<?php
	if ((isset($_GET['service'])) && (isset($_GET['operation']))) {
		$service = $_GET['service'];
		$operand = $_GET['operation'];
		exec("sudo service ".$service." ".$operation);
	}
	
	header('Location ../services.php');
?>