<?php
require_once dirname(__DIR__) . '/common/auth.php';
require_auth();

// Only serve the updates log — do not allow arbitrary file reads.
$allowed_files = [
    '/tmp/linux-webui_updates.log',
];

$file = isset($_GET['file']) ? $_GET['file'] : '';
if (!in_array($file, $allowed_files, true)) {
    http_response_code(403);
    exit('Forbidden');
}

if (!file_exists($file)) {
    echo '<pre>No update log yet.</pre>';
    exit;
}

echo '<pre>' . htmlspecialchars(file_get_contents($file)) . '</pre>';
