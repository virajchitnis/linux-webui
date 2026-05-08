<?php
require_once __DIR__ . '/common/auth.php';
require_auth();
$csrf = generate_csrf_token();
?>
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
        			<h3>linux-webui (beta)</h3>
        			<p>A simple web control panel for linux servers.</p>
        			<p>&nbsp;</p>
        			<?php
        				if ((exec("./shellscripts/testgit.sh")) == "true") {
        					$git_branch = exec("git branch | grep '*' | awk '{print $2}'");
        					$branch;
        					if ($git_branch == "master") {
        						$branch = "beta";
        					}
        					else {
        						$branch = $git_branch;
        					}
        			?>
        					<p><?php echo htmlspecialchars(exec("git describe")); ?> (<?php echo htmlspecialchars($branch); ?>)</p>
        					<form method="post" action="shellscripts/gitupdate.php">
        						<input type="hidden" name="csrf_token" value="<?php echo htmlspecialchars($csrf); ?>">
        						<button type="submit">Update</button>
        					</form>
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
