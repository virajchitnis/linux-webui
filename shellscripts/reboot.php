<?php
	if ((isset($_GET['system'])) && (($_GET['system']) == 'y')) {
		exec("nohup ./shellscripts/reboot.sh >/dev/null 2>&1 &");
	}
	
	header('Location: ../rebooting.php');
?>