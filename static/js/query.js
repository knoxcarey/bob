function bobQuery(queryElement) {

    var qa = queryElement.value.split(/[^0-9a-zA-Z]/)
    
    var xhr = new XMLHttpRequest();

    xhr.onreadystatechange = function () {
	if (xhr.readyState === 4) {
	    if (xhr.status === 200) {
		console.log(xhr.responseText);
	    } else {
		console.log('Error: ' + xhr.status);
	    }
	}
    };

    var qs = '/query?chromosome=' + qa[0] + '&start=' + qa[1] + '&referenceBases=' + qa[2] + '&alternateBases=' + qa[3] + '&assemblyId=GRCh37';

    console.log(qs);
    
    xhr.open('GET', qs);
    xhr.send(null);
}
