<?php
	exec("cd ..; sudo ./shellscripts/gitupdate.sh");
	header('Location: ../about.php');
?>