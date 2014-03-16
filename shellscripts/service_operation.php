<?php
	if ((isset($_GET['service'])) && (isset($_GET['operand']))) {
		$service = $_GET['service'];
		$operand = $_GET['operand'];
		exec("sudo service ".$service." ".$operand);
	}
	
	header('Location ../services.php');
?>