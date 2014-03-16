<?php
	if ((isset($_GET['system'])) && (($_GET['system']) == 'y')) {
		exec("sudo reboot");
	}
	
	header('Location: ../rebooting.php');
?>