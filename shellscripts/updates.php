<?php
require_once dirname(__DIR__) . '/common/auth.php';
require_auth();
$script = dirname(__DIR__) . '/shellscripts/gentoocheckupdates.sh';
exec('nohup ' . escapeshellarg($script) . ' >/tmp/linux-webui_updates.log 2>&1 &');
?>
<html>
	<head>
		<script src="../js/updates.js"></script>
	</head>
	<body>
		<iframe seamless id="framedisplay" style="width: 95%; height: 95%" src="catfile.php?file=/tmp/linux-webui_updates.log"></iframe>
	</body>
</html>
