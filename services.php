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
        							<td class="body_table_data"><?php echo shell_exec('sudo service apache2 status'); ?></td>
        							<td class="body_table_data">93 MB</td>
        							<td class="body_table_data">
        								<button>Restart</button>
        								<button>Stop</button>
        							</td>
        						</tr>
        				<?php
        					}
        				?>
        				<tr>
        					<td class="body_table_data">Apache2</td>
        					<td class="body_table_data"> * status: started</td>
        					<td class="body_table_data">93 MB</td>
        					<td class="body_table_data">
        						<button>Restart</button>
        						<button>Stop</button>
        					</td>
        				</tr>
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