package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

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

	var rawResponse map[string]interface{}
	err = json.Unmarshal(body, &rawResponse)
	if err != nil {
		return Response{}, err
	}

	response := Response{
		Answer:      rawResponse["answer"].(string),
		Coordinates: parseCoordinates(rawResponse["coordinates"].([]interface{})),
		Cells:       parseCells(rawResponse["cells"].([]interface{})),
		Aggregator:  rawResponse["aggregator"].(string),
	}

	return response, nil
}

func convertAnswer(answerStr string) string {
	parts := strings.Split(answerStr, ", ")
	var sum float64

	for _, part := range parts {
		numStr := strings.TrimPrefix(part, "SUM > ")
		num, err := strconv.ParseFloat(numStr, 64)
		if err == nil {
			sum += num
		}
	}

	answerConverted := fmt.Sprintf("%.1f kWh", sum)

	return answerConverted
}

func parseCoordinates(coordinatesRaw []interface{}) [][]int {
	coordinates := make([][]int, len(coordinatesRaw))
	for i, coord := range coordinatesRaw {
		coordArr := coord.([]interface{})
		coordinates[i] = []int{int(coordArr[0].(float64)), int(coordArr[1].(float64))}
	}
	return coordinates
}

func parseCells(cellsRaw []interface{}) []string {
	cells := make([]string, len(cellsRaw))
	for i, cell := range cellsRaw {
		switch v := cell.(type) {
		case float64:
			cells[i] = fmt.Sprintf("%.1f", v)
		default:
			cells[i] = fmt.Sprintf("%v", v)
		}
	}
	return cells
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

	query := "What is the total energy consumption for tomorrow?"

	connector := &AIModelConnector{
		Client: &http.Client{},
	}

	inputs := Inputs{
		Table: table,
		Query: query,
	}

	token := os.Getenv("HUGGINGFACE_TOKEN")
	if token == "" {
		fmt.Println("HUGGINGFACE_TOKEN environment variable not set.")
		return
	}

	response, err := connector.ConnectAIModel(inputs, token)
	if err != nil {
		fmt.Println("Error connecting to AI model:", err)
		return
	}

	answerConverted := convertAnswer(response.Answer)

	fmt.Println("AnswerConverted:", answerConverted)
	fmt.Println("Answer:", response.Answer)
	fmt.Println("Coordinates:", response.Coordinates)
	fmt.Println("Cells:", response.Cells)
	fmt.Println("Aggregator:", response.Aggregator)
}
