var socket = new WebSocket("ws://127.0.0.1:8080/queryws");
// FIXME: replace fixed address with parameterized address


socket.onopen = function () {
    console.log("Connected to websocket")
}



function bobQuery(queryElement, resultsElement) {
    socket.onmessage = function (e) {
	resultsElement.innerHTML += e.data
    }

    var qa = queryElement.value.split(/[^0-9a-zA-Z]/)
    // FIXME: eliminate stuff that's not there
    var qs = {chromosome: qa[0], start: qa[1], referenceBases: qa[2], alternateBases: qa[3]};
    
    socket.send(JSON.stringify(qs));
}


