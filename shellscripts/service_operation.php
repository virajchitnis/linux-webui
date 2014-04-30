<?php
	if ((isset($_GET['service'])) && (isset($_GET['operation']))) {
		$service = escapeshellcmd($_GET['service']);
		$operation = escapeshellcmd($_GET['operation']);
		exec("sudo service ".$service." ".$operation);
	}
	
	header('Location: ../services.php');
?>
