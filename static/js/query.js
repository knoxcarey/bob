var socket;
var url;
var inElement;
var outElement;
var button;


// Connect to various elements on the page
function connect(u, i, o, b) {
    inElement = document.getElementById(i);
    inElement.onkeypress = (e) => {
	if (e.charCode == 13) {
	    bobQuery(inElement);
	}
    };
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

    if (outElement.innerHTML) {outElement.innerHTML = null;}    
    if (socket) {socket.close();}
    socket = new WebSocket(url);    
    socket.onmessage = (e) => {displayResult(e.data)};
    socket.onopen = () => {socket.send(JSON.stringify(qs))};
}


// Display a beacon result
function displayResult(r) {
    var json = JSON.parse(r);
    var result = document.createElement('div');
    result.className += 'beacon clearfix';
    result.innerHTML = '<div class="image"><img class="icon" src="/static/img/' + (json.icon || "default.png") + '"/></div>';
    result.innerHTML += '<div class="name">' + json.name + '</div>';

    for (var dataset in json.responses) {
	if (json.responses.hasOwnProperty(dataset)) {
	    result.innerHTML += '<div class="response">' + json.responses[dataset] + '</div>';
	}
    }
    
    outElement.appendChild(result);
}

