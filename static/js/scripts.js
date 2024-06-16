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
    
    const loadingBubble = document.createElement('div');
    loadingBubble.classList.add('chat-bubble', 'ai');
    loadingBubble.innerHTML = `<div class="bubble">Luma is thinking...</div>`;
    chat.appendChild(loadingBubble);
    
    chat.scrollTop = chat.scrollHeight;

    try {
        const response = await fetch('/', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ query: message })
        });

        const data = await response.json();
        
        chat.removeChild(loadingBubble);
        
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

    } catch (error) {
        
        console.error('Error:', error);
        
        chat.removeChild(loadingBubble);

        const errorBubble = document.createElement('div');
        errorBubble.classList.add('chat-bubble', 'ai');
        errorBubble.innerHTML = `<div class="bubble">Something went wrong. Please try again.</div>`;
        chat.appendChild(errorBubble);
    }

    chat.scrollTop = chat.scrollHeight;
}

function formatResponse(response) {
    
    return response
        .replace(/\*\*(.*?)\*\*/g, '<b>$1</b>')  
        .replace(/\n\* (.*?)(?=\n|$)/g, '<br>* $1') 
        .replace(/\n/g, '<br>') 
        .replace(/\r/g, ''); 
}