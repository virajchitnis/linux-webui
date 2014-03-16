function confirmReboot () {
	if (confirm('Are you sure you want to reboot the system?')) {
		window.location.href = '.';
	}
	else {
		// Do nothing
	}
}