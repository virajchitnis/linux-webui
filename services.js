function confirmReboot () {
	if (confirm('Are you sure you want to reboot the system?')) {
		window.location.href = 'shellscripts/reboot.php?system=y';
	}
	else {
		// Do nothing
	}
}

function checkUpdate () {
	document.getElementById('update_display').style.display = "block";
}