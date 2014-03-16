<div class="header">
	<div class="header_inner">
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
    			<a href="about.php"><div class="header_brand_div">
    				<p>&nbsp;</p>
    				<h3 id="header_brand_text">linux-webui</h3>
    				<p>&nbsp;</p>
    			</div></a>
        	</td>
    	</tr>
	</table>
	</div>
</div>