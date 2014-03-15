<?php
	if ((isset($_POST['service'])) && (isset($_POST['operand']))) {
		$service = $_POST['service'];
		$operand = $_POST['operand'];
		exec("sudo service ".$service." ".$operand);
	}
	
	header('Location ../services.php');
?>