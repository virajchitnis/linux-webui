var count = 2;
var counter = setInterval(refreshTimer, 1000 * count);
function refreshTimer(){
	document.getElementById('framedisplay').contentWindow.location.reload();
}