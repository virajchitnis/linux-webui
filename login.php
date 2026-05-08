<?php
require_once __DIR__ . '/common/auth.php';

if (is_authenticated()) {
    header('Location: index.php');
    exit;
}

const MAX_ATTEMPTS    = 5;
const LOCKOUT_SECONDS = 900; // 15 minutes

$attempts    = isset($_SESSION['login_attempts'])    ? (int) $_SESSION['login_attempts']    : 0;
$locked_until = isset($_SESSION['login_lockout_until']) ? (int) $_SESSION['login_lockout_until'] : 0;
$locked       = $attempts >= MAX_ATTEMPTS && time() < $locked_until;

$error = '';

if ($locked) {
    $mins  = (int) ceil(($locked_until - time()) / 60);
    $error = "Too many failed attempts. Try again in {$mins} minute(s).";
} elseif ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $username = isset($_POST['username']) ? $_POST['username'] : '';
    $password = isset($_POST['password']) ? $_POST['password'] : '';

    if (login($username, $password)) {
        $_SESSION['login_attempts']    = 0;
        unset($_SESSION['login_lockout_until']);
        header('Location: index.php');
        exit;
    }

    $attempts++;
    $_SESSION['login_attempts'] = $attempts;
    if ($attempts >= MAX_ATTEMPTS) {
        $_SESSION['login_lockout_until'] = time() + LOCKOUT_SECONDS;
        $mins  = (int) ceil(LOCKOUT_SECONDS / 60);
        $error = "Too many failed attempts. Try again in {$mins} minute(s).";
    } else {
        $remaining = MAX_ATTEMPTS - $attempts;
        $error = "Invalid username or password. {$remaining} attempt(s) remaining.";
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
