var socket;

function connect(url, input, output) {
    var inElement = document.getElementById(input);
    var outElement = document.getElementById(output);

    socket = new WebSocket(url);
    socket.onopen = () => {console.log("Connected to websocket")};
    socket.onmessage = (e) => {outElement.innerHTML += e.data};

    button = document.getElementById("queryButton");
    button.onclick = () => {bobQuery(inElement)};
}

function bobQuery(queryElement) {
    var qs = {};
    var qa = queryElement.value.split(/[^0-9a-zA-Z]/);

    if (qa[0]) qs.chromosome     = [qa[0]];
    if (qa[1]) qs.start          = [qa[1]];
    if (qa[2]) qs.referenceBases = [qa[2]];
    if (qa[3]) qs.alternateBases = [qa[3]];
    qs.assemblyId = ["GRCh37"];    // FIXME: assembly should not be hardcoded
    
    socket.send(JSON.stringify(qs));
}


