<html>
	<head>
		<link rel="stylesheet" type="text/css" href="design.css">
		<link rel="stylesheet" type="text/css" href="index.css">
		<title>linux-webui - Linux Server Control Panel</title>
	</head>
    <body>
        <div class="wrapper">
        	<div class="header">
        		<table class="header_table">
        			<tr>
        				<td id="header_table_data_left">
        					<div class="header_button">
        						<p>&nbsp;</p>
        						<p class="header_button_text">Info</p>
        						<p>&nbsp;</p>
        					</div>
        					<div class="header_button">
								<p>&nbsp;</p>
        						<p class="header_button_text">NAS</p>
        						<p>&nbsp;</p>
        					</div>
        				</td>
        				<td id="header_table_data_right">
        					<p id="header_brand_text">linux-webui</p>
        				</td>
        			</tr>
        		</table>
        	</div>
        	<div class="body">
        		<div>&nbsp;</div>
        		<div class="body_content">
        			<div class="body_content_box">
        				<h3>OS Info</h3>
        					<?php echo "<pre>Distro: ".shell_exec('./shellscripts/linuxdistro.sh')."</pre>"; ?>
        					<?php echo "<pre>Hostname: ".shell_exec('uname -n')."</pre>"; ?>
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