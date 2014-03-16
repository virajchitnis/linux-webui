<?php
	if (isset($_GET['file'])) {
		$file = $_GET['file'];
		echo "<pre>".shell_exec("cat ".$file)."</pre>";
	}
?>