var socket = new WebSocket("ws://localhost:8080/queryws");

socket.onopen = function () {
    console.log("Connected to websocket")
}



function bobQuery(queryElement, resultsElement) {
    socket.onmessage = function (e) {
	resultsElement.innerHTML += e.data
    }

    var qa = queryElement.value.split(/[^0-9a-zA-Z]/)
    var qs = {chromosome: [qa[0]], start: [qa[1]], referenceBases: [qa[2]], alternateBases: [qa[3]]};
    
    socket.send(JSON.stringify(qs));
}


