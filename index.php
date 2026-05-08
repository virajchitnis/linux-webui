<?php
require_once __DIR__ . '/common/auth.php';
require_auth();
?>
<html>
	<head>
		<link rel="stylesheet" type="text/css" href="css/design.css">
		<link rel="stylesheet" type="text/css" href="css/index.css">
		<title>linux-webui - Linux Server Control Panel</title>
	</head>
    <body>
        <div class="wrapper">
        	<?php include("common/header.php"); ?>
        	<div class="body">
        		<div>&nbsp;</div>
        		<div class="body_content">
        			<div class="body_content_box">
        				<h3>OS</h3>
        					<?php echo "<pre>Distro: ".htmlspecialchars(shell_exec('./shellscripts/linuxdistro.sh'))."</pre>"; ?>
        					<?php echo "<pre>Hostname: ".htmlspecialchars(shell_exec('uname -n'))."</pre>"; ?>
        					<?php echo "<pre>Kernel: ".htmlspecialchars(shell_exec('uname -r'))."</pre>"; ?>
        					<?php echo "<pre>Uptime: ".htmlspecialchars(shell_exec("uptime | awk '{print $3,$4,$5}' | sed 's/,$//'"))."</pre>"; ?>
        					<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        			<div class="body_content_box">
        				<h3>System</h3>
        				<?php echo "<pre>CPU: ".htmlspecialchars(shell_exec("cat /proc/cpuinfo | grep \"model name\" | awk '{print $4,$5,$6,$7,$8,$9}' | tail -1"))."</pre>"; ?>
        				<?php echo "<pre>CPU Cores: ".htmlspecialchars(exec("cat /proc/cpuinfo | grep processor | tail -1 | awk '{print $3+1}'"))."</pre>"; ?>
        				<?php echo "<pre>RAM: ".htmlspecialchars(exec("free -m | grep Mem | awk '{print $2}'"))."MB</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        			<div class="body_content_box">
        				<h3>IP Addresses</h3>
        				<?php echo "<pre>".htmlspecialchars(exec("curl -s https://icanhazip.com"))."</pre>"; ?>
        				<?php echo "<pre>".htmlspecialchars(shell_exec("ifconfig | grep \"inet \" | awk '{print $2}'"))."</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        			<div class="body_content_box">
        				<h3>Memory</h3>
        				<?php echo "<pre>".htmlspecialchars(shell_exec("free -m"))."</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        			<div class="body_content_box">
        				<h3>CPU</h3>
        				<?php echo "<pre>".htmlspecialchars(exec("uptime | awk '{print $8,$9,$10,$11,$12}'"))."</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        			<div class="body_content_box">
        				<h3>Disk</h3>
        				<?php echo "<pre>".htmlspecialchars(shell_exec("df -h | grep Filesystem"))."</pre>"; ?>
        				<?php echo "<pre>".htmlspecialchars(shell_exec("df -h | grep /dev/sd"))."</pre>"; ?>
        				<?php echo "<pre>".htmlspecialchars(shell_exec("df -h | grep /dev/loop"))."</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        			<div class="body_content_box">
        				<h3>Logged in users</h3>
        				<?php echo "<pre>".htmlspecialchars(shell_exec("w"))."</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        		</div>
        	</div>
        	<div>&nbsp;</div>
            <div class="push"></div>
        </div>
        <?php include("common/footer.php"); ?>
    </body>
</html>
