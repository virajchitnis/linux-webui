<html>
	<head>
		<link rel="stylesheet" type="text/css" href="design.css">
		<link rel="stylesheet" type="text/css" href="index.css">
		<title>linux-webui - Linux Server Control Panel</title>
	</head>
    <body>
        <div class="wrapper">
        	<?php include("header.php"); ?>
        	<div class="body">
        		<div>&nbsp;</div>
        		<div class="body_content">
        			<h3>linux-webui</h3>
        			<h4>Written and designed by Viraj Chitnis</h4>
        			<p>&nbsp;</p>
        			<p><?php echo exec("git describe"); ?></p>
        		</div>
        	</div>
            <div class="push"></div>
        </div>
        <?php include("footer.php"); ?>
    </body>
</html>