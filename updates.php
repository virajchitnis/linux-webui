<html>
	<head>
		<script type="text/JavaScript">
			var count = 2;
			var counter = setInterval(refreshTimer, 1000 * count);
			function refreshTimer(){
				document.getElementById('framedisplay').contentWindow.location.reload();
			}
		</script>
	</head>
	<body>
		<iframe seamless id="framedisplay" style="width: 100%, height: 100%" src="shellscripts/catfile.php?file=/tmp/sysinfo.cgi.log"></iframe>
	</body>
</html>