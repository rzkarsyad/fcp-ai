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
    aiBubble.innerHTML = `<div class="bubble"><strong>TAPAS:</strong> ${data.tapas_answer || "No answer"}</div><div class="bubble"><strong>GPT-2:</strong> ${data.gpt2_answer || "No answer"}</div>`;
    chat.appendChild(aiBubble);

    chat.scrollTop = chat.scrollHeight;
}