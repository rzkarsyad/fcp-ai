async function sendMessage() {
    const input = document.getElementById('input');
    const message = input.value.trim();
    if (!message) return;

    const chat = document.getElementById('chat');
    const humanBubble = document.createElement('div');
    humanBubble.classList.add('chat-bubble', 'human');
    humanBubble.innerHTML = `<div class="bubble">${message}</div>`;
    chat.appendChild(humanBubble);

    input.value = '';

    const response = await fetch('/', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ query: message })
    });

    const data = await response.json();

    const aiBubble = document.createElement('div');
    aiBubble.classList.add('chat-bubble', 'ai');
    chat.appendChild(aiBubble);

    if (data.gemini_recommendations) {
        const recommendations = data.gemini_recommendations.map(candidate => {
            return candidate.content.parts.map(part => part.text).join('<br>');
        }).join('<br>');
        
        const geminiBubble = document.createElement('div');
        geminiBubble.classList.add('chat-bubble', 'ai');
        geminiBubble.innerHTML = `<div class="bubble">${formatResponse(recommendations)}</div>`;
        chat.appendChild(geminiBubble);
    }

    chat.scrollTop = chat.scrollHeight;
}

function formatResponse(response) {
    // Replacing **text** with <b>text</b> and adding line breaks for * points
    return response
        .replace(/\*\*(.*?)\*\*/g, '<b>$1</b>')  // Make text between ** bold
        .replace(/\n\* (.*?)(?=\n|$)/g, '<br>* $1') // Add line breaks for each point
        .replace(/\n/g, '<br>') // Convert newline characters to <br> tags
        .replace(/\r/g, ''); // Remove carriage return characters
}