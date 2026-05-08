function postForm(action, fields) {
	var form = document.createElement('form');
	form.method = 'POST';
	form.action = action;
	Object.keys(fields).forEach(function(key) {
		var input = document.createElement('input');
		input.type = 'hidden';
		input.name = key;
		input.value = fields[key];
		form.appendChild(input);
	});
	document.body.appendChild(form);
	form.submit();
}

function serviceAction(service, operation) {
	postForm('shellscripts/service_operation.php', {
		service: service,
		operation: operation,
		csrf_token: CSRF_TOKEN
	});
}

function confirmReboot () {
	if (confirm('Are you sure you want to reboot the system?')) {
		postForm('shellscripts/reboot.php', { csrf_token: CSRF_TOKEN });
	}
}

function checkUpdate () {
	document.getElementById('update_display').style.display = "block";
	document.getElementById('update_display').src = "shellscripts/updates.php";
	document.getElementById('check_update_button').onclick = hideUpdate;
	document.getElementById('check_update_button').innerHTML = "Hide Updates";
	document.getElementById('update_management').style.display = "block";
}

function hideUpdate () {
	document.getElementById('update_display').src = "";
	document.getElementById('update_display').style.display = "none";
	document.getElementById('check_update_button').onclick = checkUpdate;
	document.getElementById('check_update_button').innerHTML = "Check for updates";
	document.getElementById('update_management').style.display = "none";
	document.getElementById('gentoo_notice').style.display = "none";
}

function applyUpdates () {
	document.getElementById('update_display').src = "shellscripts/applyupdates.php";
}

function showGentoo () {
	document.getElementById('gentoo_notice').style.display = "block";
	document.getElementById('update_management').style.display = "none";
}
