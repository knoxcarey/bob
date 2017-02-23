var socket = new WebSocket("ws://localhost:8080/queryws");

socket.onopen = function () {
    console.log("Connected to websocket")
}



function bobQueryWS(queryElement, resultsElement) {
    socket.onmessage = function (e) {
	resultsElement.innerHTML += e.data
    }

    var qa = queryElement.value.split(/[^0-9a-zA-Z]/)
    var qs = {chromosome: [qa[0]], start: [qa[1]], referenceBases: [qa[2]], alternateBases: [qa[3]]};
    console.log(qs)
    console.log(JSON.stringify(qs))
    
    socket.send(JSON.stringify(qs));
}



function bobQuery(queryElement, resultsElement) {
    var qa = queryElement.value.split(/[^0-9a-zA-Z]/)
    
    var xhr = new XMLHttpRequest();

    xhr.onreadystatechange = function () {
	if (xhr.readyState === 4) {
	    if (xhr.status === 200) {
		resultsElement.innerHTML = xhr.responseText;
	    } else {
		console.log('Error: ' + xhr.status);
	    }
	}
    };

    var qs = '/query?chromosome=' + qa[0] + '&start=' + qa[1] + '&referenceBases=' + qa[2] + '&alternateBases=' + qa[3] + '&assemblyId=GRCh37';

    xhr.open('GET', qs);
    xhr.send(null);
}
