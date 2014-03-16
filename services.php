<html>
	<head>
		<link rel="stylesheet" type="text/css" href="design.css">
		<link rel="stylesheet" type="text/css" href="services.css">
		<script src="services.js"></script>
		<title>linux-webui - Linux Server Control Panel</title>
	</head>
    <body>
        <div class="wrapper">
        	<?php include("header.php"); ?>
        	<div class="body">
        		<div>&nbsp;</div>
        		<div class="body_content">
        			<table class="body_table">
        				<tr>
        					<th class="body_table_data">Service</th>
        					<th class="body_table_data">Status</th>
        					<th class="body_table_data">Memory usage</th>
        					<th class="body_table_data">Operations</th>
        					<th class="body_table_data">Autostart</th>
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
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep apache2")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/avahiinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Avahi</td>
        							<td class="body_table_data"><?php echo exec('sudo service avahi-daemon status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof avahi-daemon | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service avahi-daemon status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=avahi-daemon&operation=restart"><button>Restart</button></a>
        										<a href="shellscripts/service_operation.php?service=avahi-daemon&operation=stop"><button>Stop</button></a>
        								<?php
        									}
        									if ((exec("sudo service avahi-daemon status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=avahi-daemon&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep avahi-daemon")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/btsyncinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">BitTorrent Sync</td>
        							<td class="body_table_data"><?php echo exec('sudo service btsync status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof btsync | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service btsync status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=btsync&operation=restart"><button>Restart</button></a>
        										<a href="shellscripts/service_operation.php?service=btsync&operation=stop"><button>Stop</button></a>
        								<?php
        									}
        									if ((exec("sudo service btsync status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=btsync&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep btsync")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
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
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep cronie")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/cupsinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Cups</td>
        							<td class="body_table_data"><?php echo exec('sudo service cupsd status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof cupsd | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service cupsd status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=cupsd&operation=restart"><button>Restart</button></a>
        										<a href="shellscripts/service_operation.php?service=cupsd&operation=stop"><button>Stop</button></a>
        								<?php
        									}
        									if ((exec("sudo service cupsd status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=cupsd&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep cupsd")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
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
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep dnsmasq")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/gitdaemoninstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Git</td>
        							<td class="body_table_data"><?php echo exec('sudo service git-daemon status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof git-daemon | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service git-daemon status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=git-daemon&operation=restart"><button>Restart</button></a>
        										<a href="shellscripts/service_operation.php?service=git-daemon&operation=stop"><button>Stop</button></a>
        								<?php
        									}
        									if ((exec("sudo service git-daemon status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=git-daemon&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep git-daemon")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/mysqlinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">MySQL</td>
        							<td class="body_table_data"><?php echo exec('sudo service mysql status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof mysqld | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service mysql status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=mysql&operation=restart"><button>Restart</button></a>
        										<a href="shellscripts/service_operation.php?service=mysql&operation=stop"><button>Stop</button></a>
        								<?php
        									}
        									if ((exec("sudo service mysql status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=mysql&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep mysql")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
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
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep netatalk")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
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
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep nfs")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/sambainstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Samba</td>
        							<td class="body_table_data"><?php echo exec('sudo service samba status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof smbd | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service samba status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=samba&operation=restart"><button>Restart</button></a>
        										<a href="shellscripts/service_operation.php?service=samba&operation=stop"><button>Stop</button></a>
        								<?php
        									}
        									if ((exec("sudo service samba status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=samba&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep samba")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/squidinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Squid</td>
        							<td class="body_table_data"><?php echo exec('sudo service squid status'); ?></td>
        							<td class="body_table_data"><?php echo exec("sudo pmap $(pidof squid | awk '{print $1}') | tail -1 | awk '{print $2}'"); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service squid status | awk '{print $3}'")) == "started") {
        								?>
        										<a href="shellscripts/service_operation.php?service=squid&operation=restart"><button>Restart</button></a>
        										<a href="shellscripts/service_operation.php?service=squid&operation=stop"><button>Stop</button></a>
        								<?php
        									}
        									if ((exec("sudo service squid status | awk '{print $3}'")) == "stopped") {
        								?>
        										<a href="shellscripts/service_operation.php?service=squid&operation=start"><button>Start</button></a>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep squid")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
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
        							<td class="body_table_data">
        								<?php
        									if(!empty((exec("sudo rc-update show | grep sshd")))) {
        								?>
        										Enabled
        								<?php
        									}
        									else {
        								?>
        										Disabled
        								<?php
        									}
        								?>
        							</td>
        						</tr>
        				<?php
        					}
        				?>
        				<tr>
        				<tr>
        					<td class="body_table_data">Bind9</td>
        					<td class="body_table_data">status</td>
        					<td class="body_table_data">0K</td>
        					<td class="body_table_data"></td>
        				</tr>
        				<tr>
        					<td class="body_table_data">Polipo</td>
        					<td class="body_table_data">status</td>
        					<td class="body_table_data">0K</td>
        					<td class="body_table_data"></td>
        				</tr>
        				<tr>
        					<td class="body_table_data">Lighttpd</td>
        					<td class="body_table_data">status</td>
        					<td class="body_table_data">0K</td>
        					<td class="body_table_data"></td>
        				</tr>
        				<tr>
        					<td class="body_table_data">Nginx</td>
        					<td class="body_table_data">status</td>
        					<td class="body_table_data">0K</td>
        					<td class="body_table_data"></td>
        				</tr>
        			</table>
        			<div class="body_management">
        				<div>&nbsp;</div>
        				<button id="check_update_button" onclick="checkUpdate()">Check for updates</button>
        				&nbsp;
        				<button onclick="confirmReboot()">Reboot</button>
        			</div>
        			<div>&nbsp;</div>
        			<iframe class="update_display" id="update_display"></iframe>
        		</div>
        	</div>
        	<div>&nbsp;</div>
            <div class="push"></div>
        </div>
        <?php include("footer.php"); ?>
    </body>
</html>