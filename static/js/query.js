var socket;
var url;
var inElement;
var outElement;
var button;


// Connect to various elements on the page
function connect(u, i, o, b) {
    inElement = document.getElementById(i);
    outElement = document.getElementById(o);
    url = u;
    
    button = document.getElementById(b);
    button.onclick = () => {bobQuery(inElement)};
}


// Query the beacon of beacons asynchronously
function bobQuery(queryElement) {
    var qs = {};
    var qa = queryElement.value.split(/[^0-9a-zA-Z]/);

    if (qa[0]) qs.chromosome     = [qa[0]];
    if (qa[1]) qs.start          = [qa[1]];
    if (qa[2]) qs.referenceBases = [qa[2]];
    if (qa[3]) qs.alternateBases = [qa[3]];
    qs.assemblyId = ["GRCh37"];    // FIXME: assembly should not be hardcoded

    if (socket) {socket.close();}
    socket = new WebSocket(url);    
    socket.onmessage = (e) => {outElement.innerHTML += e.data};
    socket.onopen = () => {socket.send(JSON.stringify(qs))};
}


