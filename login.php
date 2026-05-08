<?php
require_once __DIR__ . '/common/auth.php';

if (is_authenticated()) {
    header('Location: index.php');
    exit;
}

$error = '';
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $username = isset($_POST['username']) ? $_POST['username'] : '';
    $password = isset($_POST['password']) ? $_POST['password'] : '';
    if (login($username, $password)) {
        header('Location: index.php');
        exit;
    }
    $error = 'Invalid username or password.';
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
                <input type="text" name="username" autocomplete="username" required>
                <label>Password</label>
                <input type="password" name="password" autocomplete="current-password" required>
                <button type="submit">Log In</button>
            </form>
        </div>
    </body>
</html>
