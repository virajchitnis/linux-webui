<?php
require_once __DIR__ . '/common/auth.php';
require_auth();
$csrf = generate_csrf_token();
?>
<html>
	<head>
		<link rel="stylesheet" type="text/css" href="css/design.css">
		<link rel="stylesheet" type="text/css" href="css/services.css">
		<meta name="csrf-token" content="<?php echo htmlspecialchars(generate_csrf_token(), ENT_QUOTES); ?>">
		<script src="js/services.js"></script>
		<title>linux-webui - Linux Server Control Panel</title>
	</head>
    <body>
        <div class="wrapper">
        	<?php include("common/header.php"); ?>
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
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service apache2 status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof apache2 | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service apache2 status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('apache2','restart')">Restart</button>
        								<?php
        									}
        									if ((exec("sudo service apache2 status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('apache2','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep apache2")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/avahiinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Avahi</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service avahi-daemon status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof avahi-daemon | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service avahi-daemon status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('avahi-daemon','restart')">Restart</button>
        										<button onclick="serviceAction('avahi-daemon','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service avahi-daemon status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('avahi-daemon','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep avahi-daemon")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/btsyncinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">BitTorrent Sync</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service btsync status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof btsync | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service btsync status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('btsync','restart')">Restart</button>
        										<button onclick="serviceAction('btsync','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service btsync status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('btsync','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep btsync")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/croninstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Cron</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service cronie status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof crond | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service cronie status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('cronie','restart')">Restart</button>
        										<button onclick="serviceAction('cronie','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service cronie status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('cronie','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep cronie")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/cupsinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Cups</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service cupsd status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof cupsd | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service cupsd status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('cupsd','restart')">Restart</button>
        										<button onclick="serviceAction('cupsd','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service cupsd status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('cupsd','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep cupsd")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/dnsmasqinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Dnsmasq</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service dnsmasq status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof dnsmasq | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service dnsmasq status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('dnsmasq','restart')">Restart</button>
        										<button onclick="serviceAction('dnsmasq','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service dnsmasq status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('dnsmasq','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep dnsmasq")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/gitdaemoninstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Git</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service git-daemon status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof git-daemon | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service git-daemon status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('git-daemon','restart')">Restart</button>
        										<button onclick="serviceAction('git-daemon','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service git-daemon status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('git-daemon','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep git-daemon")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/mysqlinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">MySQL</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service mysql status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof mysqld | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service mysql status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('mysql','restart')">Restart</button>
        										<button onclick="serviceAction('mysql','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service mysql status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('mysql','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep mysql")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/netatalkinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Netatalk</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service netatalk status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof afpd | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service netatalk status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('netatalk','restart')">Restart</button>
        										<button onclick="serviceAction('netatalk','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service netatalk status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('netatalk','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep netatalk")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/nfsinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">NFS</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service nfs status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof nfsd | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service nfs status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('nfs','restart')">Restart</button>
        										<button onclick="serviceAction('nfs','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service nfs status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('nfs','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep nfs")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/polipoinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Polipo</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service polipo status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof polipo | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service polipo status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('polipo','restart')">Restart</button>
        										<button onclick="serviceAction('polipo','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service polipo status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('polipo','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep polipo")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/sambainstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Samba</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service samba status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof smbd | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service samba status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('samba','restart')">Restart</button>
        										<button onclick="serviceAction('samba','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service samba status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('samba','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep samba")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/squidinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">Squid</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service squid status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof squid | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service squid status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('squid','restart')">Restart</button>
        										<button onclick="serviceAction('squid','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service squid status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('squid','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep squid")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        					if ((exec('./shellscripts/sshinstalled.sh')) == 'true') {
        				?>
        						<tr>
        							<td class="body_table_data">SSH</td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec('sudo service sshd status')); ?></td>
        							<td class="body_table_data"><?php echo htmlspecialchars(exec("sudo pmap $(pidof sshd | awk '{print $1}') | tail -1 | awk '{print $2}'")); ?></td>
        							<td class="body_table_data">
        								<?php
        									if ((exec("sudo service sshd status | awk '{print $3}'")) == "started") {
        								?>
        										<button onclick="serviceAction('sshd','restart')">Restart</button>
        										<button onclick="serviceAction('sshd','stop')">Stop</button>
        								<?php
        									}
        									if ((exec("sudo service sshd status | awk '{print $3}'")) == "stopped") {
        								?>
        										<button onclick="serviceAction('sshd','start')">Start</button>
        								<?php
        									}
        								?>
        							</td>
        							<td class="body_table_data">
        								<?php echo !empty(exec("sudo rc-update show | grep sshd")) ? 'Enabled' : 'Disabled'; ?>
        							</td>
        						</tr>
        				<?php
        					}
        				?>
        			</table>
        			<div class="body_management">
        				<div>&nbsp;</div>
        				<button id="check_update_button" onclick="checkUpdate()">Check for updates</button>
        				&nbsp;
        				<button onclick="confirmReboot()">Reboot</button>
        			</div>
        			<div>&nbsp;</div>
        			<iframe class="update_display" id="update_display"></iframe>
        			<div>&nbsp;</div>
        			<div class="update_management" id="update_management">
        				<?php
        					if ((exec("shellscripts/linuxdistro.sh")) == "Gentoo") {
        				?>
        						<button onclick="showGentoo()">Apply updates</button>
        				<?php
        					}
        					else {
        				?>
        						<button onclick="applyUpdates()">Apply updates</button>
        				<?php
        					}
        				?>
        			</div>
        			<?php
        				if ((exec("shellscripts/linuxdistro.sh")) == "Gentoo") {
        			?>
        					<div class="gentoo_notice" id="gentoo_notice">
        						<p>Applying updates via a web interface is not recommended on Gentoo. Please apply updates via SSH.</p>
        					</div>
        			<?php
        				}
        			?>
        		</div>
        	</div>
        	<div>&nbsp;</div>
            <div class="push"></div>
        </div>
        <?php include("common/footer.php"); ?>
    </body>
</html>
