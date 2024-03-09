// Build the socket connection to the server
function initSocket(){
    var socket = new WebSocket('ws://localhost:4269/wschat');
    socket.onopen = function (e) {
        console.log('Connected to server');
    };
    socket.onmessage = function (e) {
        console.log('Message from server: ' + e.data);
        //parse the JSON data
        var data = JSON.parse(e.data);
        //Add the message to the chat result with new chat_result div in the chat_container
        if (data.done == true) {
            document.getElementById('meta').innerHTML = "Result generated in " + (data.total_duration / 10 ** 9) + " seconds";
        }
        try{
            document.getElementById('chat_result').innerHTML += convertPlainTextToHTML(data.message.content);
        }catch(e){
            socket.close();
            console.log('Connection closed not found element chat_result');
        }

    };
    socket.onclose = function (e) {
        console.log('Connection closed');
    };
    var chat_form = document.getElementById('chat_form');
    if (chat_form != null) {
        //close the connection
        chat_form.addEventListener('submit', function (e) {
            console.log('Form submitted');
            e.preventDefault();
            //get the message from the input
            var message = document.getElementById('message').value;
        
            //clear chat result
            document.getElementById('chat_result').innerHTML = "";
            //send the message to the server
            socket.send(JSON.stringify({
                type: 1,
                content: message
            }));
            //clear the input
            document.getElementById('chat_prompt').innerHTML = "Prompt: " + message;
        });
    }else{
        socket.close();
    }
    
}

function convertPlainTextToHTML(text) {
    return text.replace(/\n/g, '<br>');
}

//handle the form submission
