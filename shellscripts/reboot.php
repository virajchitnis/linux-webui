<?php
require_once dirname(__DIR__) . '/common/auth.php';
require_auth();

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    header('Location: ../services.php');
    exit;
}

$csrf = isset($_POST['csrf_token']) ? $_POST['csrf_token'] : '';
if (!validate_csrf($csrf)) {
    http_response_code(403);
    exit('Forbidden');
}

exec('nohup ' . dirname(__DIR__) . '/shellscripts/reboot.sh >/dev/null 2>&1 &');

header('Location: ../rebooting.php');
