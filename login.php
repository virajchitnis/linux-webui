<?php
require_once __DIR__ . '/common/auth.php';

if (is_authenticated()) {
    header('Location: index.php');
    exit;
}

$locked = is_ip_locked_out();
$error  = '';

if ($locked) {
    $mins  = (int) ceil(ip_lockout_remaining() / 60);
    $error = "Too many failed attempts. Try again in {$mins} minute(s).";
} elseif ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $username = $_POST['username'] ?? '';
    $password = $_POST['password'] ?? '';

    if (login($username, $password)) {
        reset_ip_attempts();
        header('Location: index.php');
        exit;
    }

    record_failed_attempt();

    if (is_ip_locked_out()) {
        $mins  = (int) ceil(LOCKOUT_SECONDS / 60);
        $error = "Too many failed attempts. Try again in {$mins} minute(s).";
        $locked = true;
    } else {
        $error = 'Invalid username or password.';
    }
}
?>
<html>
    <head>
        <link rel="stylesheet" type="text/css" href="css/design.css">
        <title>linux-webui - Login</title>
        <style>
            .login_box { max-width: 320px; margin: 80px auto; padding: 30px; background: #f5f5f5; border: 1px solid #ccc; }
            .login_box input[type=text], .login_box input[type=password] { width: 100%; padding: 8px; margin: 6px 0 14px; box-sizing: border-box; }
            .login_box button { width: 100%; padding: 10px; }
            .error { color: red; margin-bottom: 10px; }
        </style>
    </head>
    <body>
        <div class="login_box">
            <h3>linux-webui Login</h3>
            <?php if ($error): ?>
                <p class="error"><?php echo htmlspecialchars($error); ?></p>
            <?php endif; ?>
            <form method="post" action="login.php">
                <label>Username</label>
                <input type="text" name="username" autocomplete="username" <?php if ($locked) echo 'disabled'; ?> required>
                <label>Password</label>
                <input type="password" name="password" autocomplete="current-password" <?php if ($locked) echo 'disabled'; ?> required>
                <button type="submit" <?php if ($locked) echo 'disabled'; ?>>Log In</button>
            </form>
        </div>
    </body>
</html>
