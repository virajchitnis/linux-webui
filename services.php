<html>
	<head>
		<link rel="stylesheet" type="text/css" href="design.css">
		<link rel="stylesheet" type="text/css" href="services.css">
		<title>linux-webui - Linux Server Control Panel</title>
	</head>
    <body>
        <div class="wrapper">
        	<div class="header">
        		<table class="header_table">
        			<tr>
        				<td id="header_table_data_left">
        					<a href="."><div class="header_button">
        						<p>&nbsp;</p>
        						<h3 class="header_button_text">Info</h3>
        						<p>&nbsp;</p>
        					</div></a>
        					<?php
        						if (file_exists('services.php')) {
        					?>
        							<a href="services.php"><div class="header_button">
										<p>&nbsp;</p>
        								<h3 class="header_button_text">Services</h3>
        								<p>&nbsp;</p>
        							</div></a>
        					<?php
        						}
        						if (file_exists('storage.php')) {
        					?>
        							<a href="storage.php"><div class="header_button">
										<p>&nbsp;</p>
        								<h3 class="header_button_text">Storage</h3>
        								<p>&nbsp;</p>
        							</div></a>
        					<?php
        						}
        						if (file_exists('users.php')) {
        					?>
        							<a href="users.php"><div class="header_button">
										<p>&nbsp;</p>
        								<h3 class="header_button_text">Users</h3>
        								<p>&nbsp;</p>
        							</div></a>
        					<?php
        						}
        						if (file_exists('web.php')) {
        					?>
        							<a href="web.php"><div class="header_button">
										<p>&nbsp;</p>
        								<h3 class="header_button_text">Web</h3>
        								<p>&nbsp;</p>
        							</div></a>
        					<?php
        						}
        					?>
        					<div id="header_button_last">
								<p>&nbsp;</p>
								<h3>&nbsp;</h3>
								<p>&nbsp;</p>
        					</div>
        				</td>
        				<td id="header_table_data_right">
        					<h3 id="header_brand_text">linux-webui</h3>
        				</td>
        			</tr>
        		</table>
        	</div>
        	<div class="body">
        		<div>&nbsp;</div>
        		<div class="body_content">
        			<table class="body_table">
        				<tr>
        					<th class="body_table_data">Service</th>
        					<th class="body_table_data">Status</th>
        					<th class="body_table_data">Memory usage</th>
        					<th class="body_table_data">Operations</th>
        				</tr>
        				<?php
        					if ((exec('./shellscripts/apacheinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Apache2</td>
        							<td class="body_table_data"><?php echo exec('sudo service apache2 status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof apache2 | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<div class="service_operation_buttons">
        								<?php
        									if ((exec("sudo service apache2 status | awk '{print $3}'")) == "started") {
        								?>
        										<form action="shellscripts/service_operation.php" method="post">
        											<input type="hidden" name="service" value="apache2">
        											<input type="hidden" name="operation" value="restart">
        											<input type="submit" value="Restart">
        										</form>
        										<form action="shellscripts/service_operation.php" method="post">
        											<input type="hidden" name="service" value="apache2">
        											<input type="hidden" name="operation" value="stop">
        											<input type="submit" value="Stop">
        										</form>
        								<?php
        									}
        									if ((exec("sudo service apache2 status | awk '{print $3}'")) == "stopped") {
        								?>
        										<form action="shellscripts/service_operation.php" method="post">
        											<input type="hidden" name="service" value="apache2">
        											<input type="hidden" name="operation" value="start">
        											<input type="submit" value="Start">
        										</form>
        								<?php
        									}
        								?>
        								</div>
        							</td>
        						</tr>
        				<?php
        					}
        				?>
        			</table>
        		</div>
        	</div>
            <div class="push"></div>
        </div>
        <div class="footer">
            <p>&copy; 2014 Viraj Chitnis</p>
        </div>
    </body>
</html>