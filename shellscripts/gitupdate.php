<?php
require_once dirname(__DIR__) . '/common/auth.php';
require_auth();

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    header('Location: ../about.php');
    exit;
}

$csrf = isset($_POST['csrf_token']) ? $_POST['csrf_token'] : '';
if (!validate_csrf($csrf)) {
    http_response_code(403);
    exit('Forbidden');
}

$script = dirname(__DIR__) . '/shellscripts/gitupdate.sh';
exec('cd ' . escapeshellarg(dirname(__DIR__)) . ' && ' . escapeshellarg($script));

header('Location: ../about.php');
