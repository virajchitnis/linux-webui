<html>
	<head>
		<link rel="stylesheet" type="text/css" href="css/design.css">
		<link rel="stylesheet" type="text/css" href="css/about.css">
		<title>linux-webui - Linux Server Control Panel</title>
	</head>
    <body>
        <div class="wrapper">
        	<?php include("common/header.php"); ?>
        	<div class="body">
        		<div>&nbsp;</div>
        		<div class="body_content">
        			<h3>linux-webui</h3>
        			<p>A simple web control panel for linux servers.</p>
        			<p>&nbsp;</p>
        			<?php
        				if ((exec("./shellscripts/testgit.sh")) == "true") {
        					$git_branch = exec("git branch | grep '*' | awk '{print $2}'");
        					$branch;
        					if ($git_branch == "master") {
        						$branch = "stable";
        					}
        					else {
        						$branch = $git_branch;
        					}
        			?>
        					<p><?php echo exec("git describe"); ?> (<?php echo $branch; ?>)</p>
        					<p><a href="shellscripts/gitupdate.php"><button>Update</button></a></p>
        					<p>&nbsp;</p>
        			<?php
        				}
        			?>
        			<p>Written and designed by Viraj Chitnis</p>
        		</div>
        	</div>
        	<div>&nbsp;</div>
            <div class="push"></div>
        </div>
        <?php include("common/footer.php"); ?>
    </body>
</html>