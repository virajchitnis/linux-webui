<html>
	<head>
		<link rel="stylesheet" type="text/css" href="design.css">
		<link rel="stylesheet" type="text/css" href="about.css">
		<script type="text/JavaScript">
			var count = 180;
			var counter = setInterval(refreshTimer, 1000 * count);
			function refreshTimer(){
				window.location.href = 'services.php';
			}
		</script>
		<title>linux-webui - Linux Server Control Panel</title>
	</head>
    <body>
        <div class="wrapper">
        	<?php include("header.php"); ?>
        	<div class="body">
        		<div>&nbsp;</div>
        		<div class="body_content">
        			<h3>Rebooting</h3>
        			<p>The server is rebooting, please wait.....</p>
        			<p>The page will auto redirect once the rebooting is complete.</p>
        		</div>
        	</div>
            <div class="push"></div>
        </div>
        <?php include("footer.php"); ?>
    </body>
</html>