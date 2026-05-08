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

define('SESSION_TIMEOUT',  1800); // 30 minutes idle
define('MAX_ATTEMPTS',        5);
define('LOCKOUT_SECONDS',   900); // 15 minutes
define('UPDATE_LOG', dirname(__DIR__) . '/data/updates.log');

// ---------------------------------------------------------------------------
// Authentication
// ---------------------------------------------------------------------------

function is_authenticated() {
    return isset($_SESSION['authenticated']) && $_SESSION['authenticated'] === true;
}

function require_auth() {
    if (!is_authenticated()) {
        _redirect_to_login();
    }

    if (isset($_SESSION['last_active']) && (time() - $_SESSION['last_active']) > SESSION_TIMEOUT) {
        logout();
        _redirect_to_login();
    }

    $_SESSION['last_active'] = time();
}

function login($username, $password) {
    if ($username === AUTH_USERNAME && password_verify($password, AUTH_PASSWORD_HASH)) {
        session_regenerate_id(true);
        $_SESSION['authenticated'] = true;
        $_SESSION['last_active']   = time();
        return true;
    }
    return false;
}

function logout() {
    $_SESSION = [];
    session_destroy();
}

function _redirect_to_login() {
    $depth  = substr_count(trim($_SERVER['SCRIPT_NAME'], '/'), '/');
    $prefix = str_repeat('../', max(0, $depth - 1));
    header('Location: ' . $prefix . 'login.php');
    exit;
}

// ---------------------------------------------------------------------------
// CSRF
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// IP-based login rate limiting
// ---------------------------------------------------------------------------

function _rl_file() {
    $dir = dirname(__DIR__) . '/data';
    if (!is_dir($dir)) {
        mkdir($dir, 0750, true);
    }
    return $dir . '/login_attempts.json';
}

function _rl_read($fp) {
    rewind($fp);
    $data = json_decode(stream_get_contents($fp), true);
    return is_array($data) ? $data : [];
}

function _rl_write($fp, array $data) {
    // Prune entries whose lockout expired more than one period ago
    $cutoff = time() - LOCKOUT_SECONDS;
    foreach ($data as $ip => $info) {
        if ($info['lockout_until'] > 0 && $info['lockout_until'] < $cutoff) {
            unset($data[$ip]);
        }
    }
    rewind($fp);
    ftruncate($fp, 0);
    fwrite($fp, json_encode($data));
}

function is_ip_locked_out() {
    $path = _rl_file();
    if (!file_exists($path)) return false;
    $fp = fopen($path, 'r');
    if (!$fp) return false;
    flock($fp, LOCK_SH);
    $data = _rl_read($fp);
    flock($fp, LOCK_UN);
    fclose($fp);
    $ip = $_SERVER['REMOTE_ADDR'] ?? '';
    return isset($data[$ip])
        && $data[$ip]['attempts'] >= MAX_ATTEMPTS
        && time() < $data[$ip]['lockout_until'];
}

function ip_lockout_remaining() {
    $path = _rl_file();
    if (!file_exists($path)) return 0;
    $fp = fopen($path, 'r');
    if (!$fp) return 0;
    flock($fp, LOCK_SH);
    $data = _rl_read($fp);
    flock($fp, LOCK_UN);
    fclose($fp);
    $ip = $_SERVER['REMOTE_ADDR'] ?? '';
    return isset($data[$ip]) ? max(0, $data[$ip]['lockout_until'] - time()) : 0;
}

function record_failed_attempt() {
    $fp = fopen(_rl_file(), 'c+');
    if (!$fp) return;
    flock($fp, LOCK_EX);
    $data = _rl_read($fp);
    $ip   = $_SERVER['REMOTE_ADDR'] ?? '';
    if (!isset($data[$ip])) {
        $data[$ip] = ['attempts' => 0, 'lockout_until' => 0];
    }
    // Reset counter if a previous lockout has fully expired
    if ($data[$ip]['lockout_until'] > 0 && time() >= $data[$ip]['lockout_until']) {
        $data[$ip] = ['attempts' => 0, 'lockout_until' => 0];
    }
    $data[$ip]['attempts']++;
    if ($data[$ip]['attempts'] >= MAX_ATTEMPTS) {
        $data[$ip]['lockout_until'] = time() + LOCKOUT_SECONDS;
    }
    _rl_write($fp, $data);
    flock($fp, LOCK_UN);
    fclose($fp);
}

function reset_ip_attempts() {
    $path = _rl_file();
    if (!file_exists($path)) return;
    $fp = fopen($path, 'c+');
    if (!$fp) return;
    flock($fp, LOCK_EX);
    $data = _rl_read($fp);
    unset($data[$_SERVER['REMOTE_ADDR'] ?? '']);
    _rl_write($fp, $data);
    flock($fp, LOCK_UN);
    fclose($fp);
}
