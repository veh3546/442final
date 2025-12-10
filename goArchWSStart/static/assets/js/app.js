async function login() {
    try {
        const form = document.getElementById('login-form');
        if (!form) return console.error('login form not found');
        const fd = new FormData(form);

        const res = await fetch('/login', {
            method: 'POST',
            body: fd,
            credentials: 'same-origin'
        });

        if (res.redirected) {
            window.location.href = res.url;
            return;
        }
        if (res.ok) {
            window.location.href = '/lobby';
            return;
        }

        const text = await res.text();
        alert('Login failed: ' + text);
    } catch (e) {
        console.error(`Login failed`, e);
        alert('Login error, check console');
    }
}

async function guestLogin() {
    const usernameEl = document.getElementById('username');
    if (usernameEl && !usernameEl.value) {
        usernameEl.value = 'Guest' + Date.now().toString().slice(-4);
    }
    await login();
}

async function getBoard() {
window.location.href = '/board';
            return;}

async function fetchTurn() {
    const res = await fetch(`/turn`),
          data = await res.json();
    document.getElementById(`turn`).textContent = data.currentTurn;
}

async function nextTurn() {
    const res = await fetch(`/next`),
          data = await res.json();
    document.getElementById(`turn`).textContent = data.nextTurn;
}

let USERNAME = null; // will be fetched from server session
let ws;

function connectWebSocket() {
            const protocol = window.location.protocol === `https:` ? `wss:` : `ws:`,
                // rely on server-side session (cookie) to authenticate the websocket
                wsUrl = `${protocol}//${window.location.host}/ws/chat`;
    
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        console.log(`WebSocket connected`);
    };
    ws.onerror = (error) => {
        console.log(`WebSocket error`, error);
    };
    ws.onclose = () => {
        console.log(`WebSocket disconnected`);
        //attempt to reconnect after 3 seconds
        setTimeout(connectWebSocket, 3000);
    };
    ws.onmessage = (event) => {
        console.log(`WebSocket`, event);
        const message = JSON.parse(event.data);
        displayMessage(message);
        if (message.type === 'userList') {
            displayOnlineUsers(message.users);
        }
    };
    
}
function displayOnlineUsers(users) {
    const userListEl = document.getElementById('user-list');
    if (!userListEl) return;
    userListEl.innerHTML = '';
    users.forEach(user => {
        const li = document.createElement('li');
        li.textContent = user.username;
        userListEl.appendChild(li);
    });
}
function displayMessage(message) {
    const chatContent = document.getElementById(`chat-content`),
                    timeText = new Date(message.time).toLocaleTimeString(`en-US`, { hour12: false });
        const sender = message.username || message.user || 'Anon';
        const html = `
                    <div class="chat-message">
                        <span class="message-time">[${timeText}]</span>
                        <strong class="message-sender">${sender}:</strong>
                        <span class="message-text"> ${message.message}</span>
                    </div>
                `;
    chatContent.insertAdjacentHTML(`beforeend`, html);
    chatContent.scrollTop = chatContent.scrollHeight;

}

function sendMessage() {
    const messageInput = document.getElementById(`chat-message`),
          messageText = messageInput.value.trim();

    if(!messageText || !ws || ws.readyState !== WebSocket.OPEN) return;
    const message = {
        message: messageText,
        time: new Date().toISOString()
    };
    ws.send(JSON.stringify(message));
    messageInput.value = ``;
}

document.addEventListener(`DOMContentLoaded`, () => {
    // wire up login page buttons if present
    const loginBtn = document.getElementById('login-btn');
    const guestBtn = document.getElementById('guest-btn');
    const requestBtn = document.getElementById('request-game-btn');

    if (loginBtn) loginBtn.addEventListener('click', (e) => { e.preventDefault(); login(); });
    if (guestBtn) guestBtn.addEventListener('click', (e) => { e.preventDefault(); guestLogin(); });
    if (requestBtn) requestBtn.addEventListener('click', (e) => { e.preventDefault(); getBoard(); });

    // Chat functionality
    const sendBtn = document.getElementById(`send-btn`),
          messageInput = document.getElementById(`chat-message`);
    
    sendBtn.addEventListener(`click`, sendMessage);
    
    messageInput.addEventListener(`keypress`, (e) => {
        if (e.key === `Enter` && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    });
    
    // On load, verify server session and fetch current user
    (async () => {
        try {
            const res = await fetch('/me');
            if (!res.ok) {
                // not authenticated -> go to login
                window.location.href = '/login.html';
                return;
            }
            const info = await res.json();
            USERNAME = info.username || info.name || null;
            const who = document.getElementById('who');
            if (who && USERNAME) who.textContent = `(${USERNAME})`;

            // Connect WebSocket after session verification so cookies are sent
            connectWebSocket();
            await fetchTurn();
        } catch (e) {
            console.error('Session check failed', e);
            window.location.href = '/login.html';
        }
    })();
});
