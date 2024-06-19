package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type AIModelConnector struct {
	Client *http.Client
}

type Inputs struct {
	Table map[string][]string `json:"table"`
	Query string              `json:"query"`
}

type Response struct {
	Answer      string   `json:"answer"`
	Coordinates [][]int  `json:"coordinates"`
	Cells       []string `json:"cells"`
	Aggregator  string   `json:"aggregator"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func CsvToSlice(data string) (map[string][]string, error) {
	r := csv.NewReader(strings.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	table := make(map[string][]string)
	headers := records[0]

	for _, header := range headers {
		table[header] = make([]string, 0)
	}

	for _, record := range records[1:] {
		for i, value := range record {
			table[headers[i]] = append(table[headers[i]], value)
		}
	}

	return table, nil
}

func (c *AIModelConnector) ConnectAIModel(payload interface{}, token string) (Response, error) {
	url := "https://api-inference.huggingface.co/models/google/tapas-base-finetuned-wtq"
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return Response{}, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return Response{}, err
	}

	return response, nil
}

func (c *AIModelConnector) GeminiRecommendation(query string, table map[string][]string, token string) (GeminiResponse, error) {

	prompt := query + "\n"
	for column, values := range table {
		prompt += column + ": " + strings.Join(values, ", ") + "\n"
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=" + token

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": prompt,
					},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return GeminiResponse{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return GeminiResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return GeminiResponse{}, err
	}
	defer resp.Body.Close()

	var geminiResponse GeminiResponse
	err = json.NewDecoder(resp.Body).Decode(&geminiResponse)
	if err != nil {
		return GeminiResponse{}, err
	}

	return geminiResponse, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	file, err := os.Open("data-series.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	table, err := CsvToSlice(string(data))
	if err != nil {
		fmt.Println("Error parsing CSV:", err)
		return
	}

	router := gin.Default()

	router.Static("/static", "./static")

	router.GET("/", func(c *gin.Context) {
		c.File("templates/index.html")
	})

	router.POST("/", func(c *gin.Context) {
		var input struct {
			Query string `json:"query"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		connector := &AIModelConnector{
			Client: &http.Client{},
		}

		inputs := Inputs{
			Table: table,
			Query: input.Query,
		}

		token := os.Getenv("HUGGINGFACE_TOKEN")
		if token == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "HUGGINGFACE_TOKEN environment variable not set"})
			return
		}

		response, err := connector.ConnectAIModel(inputs, token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to AI model"})
			return
		}

		apiKey := os.Getenv("API_KEY_GEMINI")
		if apiKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "API_KEY_GEMINI environment variable not set"})
			return
		}

		geminiResponse, err := connector.GeminiRecommendation(input.Query, table, apiKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting Gemini recommendation"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tapas_answer":           response.Answer,
			"gemini_recommendations": geminiResponse.Candidates,
		})
	})

	router.Run(":8080")
}
