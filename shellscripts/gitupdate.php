<?php
	exec("cd ..; sudo ./shellscripts/update.sh");
	header('Location: about.php');
?>