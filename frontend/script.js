const msgBox = document.getElementById("messages");
const msgInput = document.getElementById("msg");
const socket = new WebSocket("ws://localhost:3000/ws");

function insertMsg(msg) {
    const newMsg = document.createElement("p");
    newMsg.innerHTML = `<span>${msg.sender}</span> ${msg.text}`;
    msgBox.appendChild(newMsg);
}

(() => {
    // TODO: enter in the input filed: /token {token}
    // 1. check token
    // 3. check if the token is correct
    // 4. if login succeeded, store username localy

    let name = "";

    socket.onopen = (event) => {
        // get message history (array of messages)
        // loop on them and insertMst(msg)

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
        if (input !== "" && event.key == "Enter") {
            if (name == "") {
                if (input.startsWith("/name ")) {
                    name = input.replace("/name ", "");
                    socket.send(JSON.stringify({ name: name }))
                    msgInput.value = "";
                } else {
                    msgInput.value = "please enter your name like: /name {your name}";
                    setTimeout(() => {
                        msgInput.value = "";
                    }, 2000);
                }
            }
            else {
                const msg = {
                    text: input,
                };
                socket.send(JSON.stringify(msg))
                msgInput.value = "";
            }
        }
    })
})();
