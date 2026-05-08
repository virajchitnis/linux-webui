<?php
if (session_status() === PHP_SESSION_NONE) {
    $secure = isset($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off';
    session_set_cookie_params([
        'lifetime' => 0,
        'path'     => '/',
        'secure'   => $secure,
        'httponly' => true,
        'samesite' => 'Strict',
    ]);
    session_start();
}

require_once dirname(__DIR__) . '/config/credentials.php';

function is_authenticated() {
    return isset($_SESSION['authenticated']) && $_SESSION['authenticated'] === true;
}

function require_auth() {
    if (!is_authenticated()) {
        $root = dirname($_SERVER['SCRIPT_NAME']);
        // Normalize to always find login.php at project root
        $depth = substr_count(trim($_SERVER['SCRIPT_NAME'], '/'), '/');
        $prefix = str_repeat('../', max(0, $depth - 1));
        header('Location: ' . $prefix . 'login.php');
        exit;
    }
}

function generate_csrf_token() {
    if (empty($_SESSION['csrf_token'])) {
        $_SESSION['csrf_token'] = bin2hex(random_bytes(32));
    }
    return $_SESSION['csrf_token'];
}

function validate_csrf($token) {
    return !empty($token)
        && !empty($_SESSION['csrf_token'])
        && hash_equals($_SESSION['csrf_token'], $token);
}

function login($username, $password) {
    if ($username === AUTH_USERNAME && password_verify($password, AUTH_PASSWORD_HASH)) {
        session_regenerate_id(true);
        $_SESSION['authenticated'] = true;
        return true;
    }
    return false;
}

function logout() {
    $_SESSION = [];
    session_destroy();
}
