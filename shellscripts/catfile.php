<?php
require_once dirname(__DIR__) . '/common/auth.php';
require_auth();

$allowed_files = [UPDATE_LOG];

$file = $_GET['file'] ?? '';
if (!in_array($file, $allowed_files, true)) {
    http_response_code(403);
    exit('Forbidden');
}

if (!file_exists($file)) {
    echo '<pre>No update log yet.</pre>';
    exit;
}

echo '<pre>' . htmlspecialchars(file_get_contents($file)) . '</pre>';
