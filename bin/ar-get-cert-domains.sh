#! /bin/bash
set -euo pipefail

HOST="$1"
PORT="443"

if [ ! -z "$2" ]; then
  PORT="$2"
fi

{
  echo \
    | openssl s_client -showcerts -connect $HOST:$PORT 2>/dev/null \
    | timeout 5 openssl x509 -inform pem -noout -text \
    | grep -A1 -iP "X509v3 subject alternative" \
    | tail -n 1 \
    | sed 's/,/\n/g' \
    | awk -F':' '{print $2}'

    echo \
      | openssl s_client -showcerts -connect $HOST:$PORT 2>/dev/null \
      | timeout 5 openssl x509 -inform pem -noout -text \
      | grep -i "Subject:" \
      | grep -oP "CN ?=[ ]*[^ ]+" \
      | cut -d= -f2 \
      | sed 's/^[ ]*//g'
} | sort -u
