const msgBox = document.getElementById("messages");
const msgInput = document.getElementById("msg");
const socket = new WebSocket("ws://localhost:8080");

function insertMsg(msg) {
    const newMsg = document.createElement("p");
    newMsg.innerHTML = msg;
    msgBox.appendChild(newMsg);

    // Scroll to the bottom
    msgBox.scrollTop = msgBox.scrollHeight;
}

(() => {
    // TODO: login /login {username} {token}
    // 1. send the username from the user
    // 2. check if username is already taken by sending a request to the server: POST /login
    // 3. check if the token is correct
    // 4. if login succeeded, store username localy

    socket.addEventListener("open", (event) => {
        // get message history (array of messages)
        // loop on them and insertMst(msg)
        console.log(event.data);
    });
    socket.addEventListener("message", (event) => {
        // insertMsg(msg)
        console.log(event.data);
    });
    socket.addEventListener("error", (error) => {
        console.error("websocket error: ", error);
    });

    msgInput.addEventListener("keydown", (event) => {
        if (msgInput.value.trim() !== "" && event.key == "Enter") {
            const msg = {
                text: msgInput.value.trim(),
            };
            socket.send(JSON.stringify(msg))
            msgInput.value = "";
        }
    })
})();
