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
        				<h3>OS Info</h3>
        					<?php echo "<pre>Distro: ".shell_exec('./shellscripts/linuxdistro.sh')."</pre>"; ?>
        					<?php echo "<pre>Hostname: ".shell_exec('uname -n')."</pre>"; ?>
        					<?php echo "<pre>Kernel: ".shell_exec('uname -r')."</pre>"; ?>
        					<?php echo "<pre>Uptime: ".shell_exec("uptime | awk '{print $3,$4,$5}' | sed 's/,$//'")."</pre>"; ?>
        					<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        			<div class="body_content_box">
        				<h3>System Info</h3>
        				<?php echo "<pre>CPU: ".shell_exec("cat /proc/cpuinfo | grep \"model name\" | awk '{print $4,$5,$6,$7,$8,$9}' | tail -1")."</pre>"; ?>
        				<?php echo "<pre>RAM: ".exec("free -m | grep Mem | awk '{print $2}'")."MB</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        			<div class="body_content_box">
        				<h3>IP Addresses</h3>
        				<?php echo "<pre>".exec("curl -s icanhazip.com")."</pre>"; ?>
        				<?php echo "<pre>".shell_exec("ifconfig | grep \"inet \" | awk '{print $2}'")."</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        			<div class="body_content_box">
        				<h3>Memory Info</h3>
        				<?php echo "<pre>".shell_exec("free -m")."</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        			<div class="body_content_box">
        				<h3>Disk Info</h3>
        				<?php echo "<pre>".shell_exec("df -h | grep Filesystem")."</pre>"; ?>
        				<?php echo "<pre>".shell_exec("df -h | grep /dev/sd")."</pre>"; ?>
        				<?php echo "<pre>".shell_exec("df -h | grep /dev/loop")."</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        			<div class="body_content_box">
        				<h3>Logged in users</h3>
        				<?php echo "<pre>".shell_exec("w")."</pre>"; ?>
        				<?php echo "<pre>&nbsp;</pre>"; ?>
        			</div>
        			<div class="body_content_box">
        				<h3>Processes</h3>
        				<div class="body_content_processes">
        					<iframe class="process_iframe" src="shellscripts/process.php"></iframe>
        				</div>
        			</div>
        		</div>
        	</div>
        	<div>&nbsp;</div>
            <div class="push"></div>
        </div>
        <?php include("common/footer.php"); ?>
    </body>
</html>