var count = 1;
var counter = setInterval(refreshTimer, 1000 * count);
function refreshTimer(){
	if (count == 180) {
		window.location.href = 'services.php';
	}
	else {
		if (((count % 10) == 0) || (count == 1)) {
			document.getElementById("progress_dots").innerHTML = ".";
		}
		else {
			document.getElementById("progress_dots").innerHTML += ".";
		}
	}
	count++;
}