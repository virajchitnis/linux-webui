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
        								<?php
        									if ((exec("sudo service apache2 status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=apache2&operation=restart"><button>Restart</button></a>
        								<?php
        									}
        									if ((exec("sudo service apache2 status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=apache2&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/croninstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Cron</td>
        							<td class="body_table_data"><?php echo exec('sudo service cronie status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof crond | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service cronie status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=cronie&operation=restart"><button>Restart</button></a>
        										<a href="shellscripts/service_operation.php?service=cronie&operation=stop"><button>Stop</button></a>
        								<?php
        									}
        									if ((exec("sudo service cronie status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=cronie&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/dnsmasqinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Dnsmasq</td>
        							<td class="body_table_data"><?php echo exec('sudo service dnsmasq status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof dnsmasq | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service dnsmasq status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=dnsmasq&operation=restart"><button>Restart</button></a>
        										<a href="shellscripts/service_operation.php?service=dnsmasq&operation=stop"><button>Stop</button></a>
        								<?php
        									}
        									if ((exec("sudo service dnsmasq status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=dnsmasq&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/netatalkinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Netatalk</td>
        							<td class="body_table_data"><?php echo exec('sudo service netatalk status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof afpd | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service netatalk status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=netatalk&operation=restart"><button>Restart</button></a>
        										<a href="shellscripts/service_operation.php?service=netatalk&operation=stop"><button>Stop</button></a>
        								<?php
        									}
        									if ((exec("sudo service netatalk status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=netatalk&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/nfsinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">NFS</td>
        							<td class="body_table_data"><?php echo exec('sudo service nfs status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof nfsd | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service nfs status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=nfs&operation=restart"><button>Restart</button></a>
        										<a href="shellscripts/service_operation.php?service=nfs&operation=stop"><button>Stop</button></a>
        								<?php
        									}
        									if ((exec("sudo service nfs status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=nfs&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/sshinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">SSH</td>
        							<td class="body_table_data"><?php echo exec('sudo service sshd status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof sshd | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service sshd status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=sshd&operation=restart"><button>Restart</button></a>
        										<a href="shellscripts/service_operation.php?service=sshd&operation=stop"><button>Stop</button></a>
        								<?php
        									}
        									if ((exec("sudo service sshd status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=sshd&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        				?>
        			</table>
        			<div class="body_management">
        				<div>&nbsp;</div>
        				<button>Check for updates</button>
        				&nbsp;
        				<a href="shellscripts/reboot.php?system=y"><button>Reboot</button></a>
        			</div>
        		</div>
        	</div>
            <div class="push"></div>
        </div>
        <div class="footer">
            <p>&copy; 2014 Viraj Chitnis</p>
        </div>
    </body>
</html>