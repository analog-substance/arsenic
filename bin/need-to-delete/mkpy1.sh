#! /bin/bash
PY='/usr/bin/python2.7'
if [ ! -z $2 ] ; then 
  PY="$2"
fi
echo "echo \"import base64; exec(base64.b64decode('$(cat $1 | base64 -w0)'))\" | $PY"
