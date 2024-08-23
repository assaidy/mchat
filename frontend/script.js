const msgBox = document.getElementById("messages");
const msgInput = document.getElementById("msg");
const socket = new WebSocket("ws://localhost:3000/ws");

function insertMsg(msg) {
    const newMsg = document.createElement("p");
    if (msg.type === "chat") {
        newMsg.innerHTML = `<span>${msg.sender}</span> ${msg.text}`;
    } else {
        newMsg.innerHTML = `${msg.text}`;
    }
    msgBox.appendChild(newMsg);
}

(() => {
    // TODO: always display user Name to the user to distinguish them while testing

    let name = "";

    socket.onopen = (event) => {
        console.log(event.data);
    };
    socket.onmessage = (event) => {
        console.log(event.data);
        insertMsg(JSON.parse(event.data));
    };
    socket.onerror = (error) => {
        console.error("websocket error: ", error);
    };

    msgInput.addEventListener("keydown", (event) => {
        let input = msgInput.value.trim();
        if (input !== "" && event.key === "Enter") {
            if (name === "") {
                if (input.startsWith("/name ")) {
                    name = input.replace("/name ", "");
                    socket.send(JSON.stringify({ name: name }))
                    msgInput.placeholder = "enter a message";
                }
            }
            else {
                const msg = {
                    text: input,
                };
                socket.send(JSON.stringify(msg))
            }
            msgInput.value = "";
        }
    })
})();
