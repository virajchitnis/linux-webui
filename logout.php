<?php
require_once __DIR__ . '/common/auth.php';
logout();
header('Location: login.php');
exit;
