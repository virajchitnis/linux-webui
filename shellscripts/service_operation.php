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

$allowed_services = [
    'apache2', 'avahi-daemon', 'btsync', 'cronie', 'cupsd', 'dnsmasq',
    'git-daemon', 'mysql', 'netatalk', 'nfs', 'polipo', 'samba', 'squid', 'sshd',
];
$allowed_operations = ['start', 'stop', 'restart'];

$service   = isset($_POST['service'])   ? $_POST['service']   : '';
$operation = isset($_POST['operation']) ? $_POST['operation'] : '';

if (!in_array($service, $allowed_services, true) || !in_array($operation, $allowed_operations, true)) {
    http_response_code(400);
    exit('Invalid service or operation');
}

exec('sudo service ' . escapeshellarg($service) . ' ' . escapeshellarg($operation));

header('Location: ../services.php');
