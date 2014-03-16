<html>
	<head>
		<script src="../js/updates.js"></script>
		<?php exec("nohup ./gentoocheckupdates.sh >/tmp/linux-webui_updates.log 2>&1 &"); ?>
	</head>
	<body>
		<iframe seamless id="framedisplay" style="width: 95%; height: 95%" src="catfile.php?file=/tmp/linux-webui_updates.log"></iframe>
	</body>
</html>