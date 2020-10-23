#! /bin/bash

function getFlags {
  ip="$1"
  {

    if [ ! -d "hosts/$ip/recon/" ]; then
      mkdir -p "hosts/$ip/recon/"
    fi

    NOPORTS=0
    if [ -f "hosts/$ip/recon/nmap-punched.nmap" ] ; then
      if grep "Error #487" "hosts/$ip/recon/nmap-punched.nmap" 1>/dev/null ; then
        echo 'NOPORTS'
        NOPORTS=1
      else
        echo 'PORTS'
        cat  "hosts/$ip/recon/nmap-punched.nmap" | grep -v "may be unreliable because we could not find at least 1 open and 1 closed port" | grep open | awk '{$1=$2=$3=""; print $0}' | grep -P '[^\s]+' | sed 's/^\s\+/SVC::/g' | sort -u
      fi
    else
      if [ -f "hosts/$ip/recon/${ip}_services.xml" ]; then

        if cat "hosts/$ip/recon/${ip}_services.xml" | grep 'state="open"' > /dev/null 2>&1 ; then
          echo 'PORTS'
          cat "hosts/$ip/recon/${ip}_services.xml" | grep 'state="open"' | grep -oP '<service .+$' | sed 's/<\/\?[^ >]*>\? \?//g' |  sed 's/name="\([^"]\+\)" \(product="\([^"]\+\)"\)\?\(.*\)/\1\n\3/g' | grep "." |  sed 's/^/SVC::/g'
        else
          echo 'NOPORTS'
          NOPORTS=1
        fi
      else
        NOPORTS=1
        echo 'no-nmap'
      fi
    fi
    if compgen -G "hosts/$ip/recon/"wappalyzer* 2>&1 > /dev/null ; then
      ls "hosts/$ip/recon/"wappalyzer* | while read wappfile; do
        cat  "$wappfile" \
        | jq -r ' .applications[] | .categories[] |[ "WAPP-CAT", .[] ] | join("::")'
        cat  "$wappfile" \
        | jq -r ' .applications[] | [ "WAPP", .name ] | join("::")'
      done
    fi

    if [ -f "hosts/$ip/README.md" ] ; then
      if cat "hosts/$ip/README.md" | grep -i "response = \"no\"" 2>&1 >/dev/null; then
        echo 'unresponsive'
      fi

      if cat "hosts/$ip/README.md" | grep -i "response = \"yes\"" 2>&1 >/dev/null; then
        echo 'responsive'
      fi

      if cat "hosts/$ip/README.md" | grep -i "reviewer = \"" 2>&1 >/dev/null; then
        echo 'reviewed'
      else
        if [ $NOPORTS -eq 0 ]; then
          echo 'unreviewed'
        else
          echo 'reviewed'
        fi
      fi

      # Get existing flags
      cat hosts/$ip/README.md \
      | grep flags \
      | cut -d= -f2 \
      | sed 's/\(\[\|\]\)*//g' | sed 's/,/\n/g' | sed 's/"//g' | sed 's/^\s\+//g' | grep -vP "reviewed|responsive|WAPP::|WAPP-CAT::|NET::|SVC::|no-nmap|PORTS|(dir|go)buster|aquatone"
    fi

    if [ ! -f "hosts/$ip/recon/whois.txt" ] ; then
      whois "$ip" > "hosts/$ip/recon/whois.txt"
    fi
    grep "NetName" "hosts/$ip/recon/whois.txt" | awk '{print $NF}'|sed 's/\(PRIVATE-ADDRESS\)/\1\n\1/' |sed 's/^/NET::/g'

    if compgen -G "hosts/$ip/recon/dirbuster"* 2>&1 > /dev/null ; then
      echo dirbuster
    fi

    if compgen -G "hosts/$ip/recon/gobuster"* 2>&1 > /dev/null ; then
      echo gobuster
    fi

    if compgen -G "hosts/$ip/recon/aquatone"* 2>&1 > /dev/null  ; then
      echo aquatone
    fi

    if [ ! -z "$NEW_FLAG" ]; then
      echo $NEW_FLAG
    fi
  } | sort -u | sed 's/ /spaaaacee/g' | sed 's/^\(.\+\)$/"\1"/g'
}

function getHosts {
  if [ $GITMODE -eq 1 ]; then
    git status | grep -P "hosts/[^/]" | awk '{print $NF}' | cut -d/ -f2 | sed 's|^\(.*\)$|hosts/\1/recon/|g' | sort -u
  elif [ ! -z "$HOST" ]; then
    echo $HOST
  else
    find hosts -maxdepth 1 -type d  -print | tail -n +2
  fi
}

GITMODE=0
NEW_FLAG=""
HOST=""
if [ "$1" == "git" ]; then
  GITMODE=1
elif [ -f "hosts/$1/README.md" ] ; then
  if [ ! -z "$2" ]; then
    HOST="hosts/$1/recon/"
    NEW_FLAG="$2"
  fi
fi


getHosts | while read d; do
  ip=$(echo $d | cut -d/ -f2);
  # if [ "$ip" != "deub ip" ] ; then
  #   continue
  # fi
  flags=$( echo $(getFlags "$ip") | sed 's|\/|/|g' | sed 's/\([\/]\)/\\\1/g' | sed 's/ /,/g' | sed 's/spaaaacee/ /g')
  if [ ! -z "$flags" ]; then
    if cat hosts/$ip/README.md | grep -P "^flags = \[" > /dev/null; then
      # update existing
      if grep -P "\[$(echo $flags | sed 's/\([()]\)/\\\1/g')\]" hosts/$ip/README.md > /dev/null; then
        echo noop > /dev/null
      else
        echo "[+] Updating $flags for $ip"
        cat hosts/$ip/README.md | sed 's/flags = \[.*\]/flags = ['"$flags"']/' > hosts/$ip/README.md.new
        mv hosts/$ip/README.md.new hosts/$ip/README.md
      fi
    else
      if cat hosts/$ip/README.md | grep '+++' > /dev/null ; then
        # add to existing front matter
        echo "[+] Add $flags for $ip"
        cat hosts/$ip/README.md | sed '0,/+++/! {0,/+++/ s/+++/flags = ['"$flags"']\n+++/}' > hosts/$ip/README.md.new
        mv hosts/$ip/README.md.new hosts/$ip/README.md

      else
        # no front matter found lets add it
        echo "[+] Creating $flags for $ip"
        {
          echo "+++"
          echo "flags = [$flags]"
          echo "+++"
          echo
          cat hosts/$ip/README.md
        } > hosts/$ip/README.md.new
        mv hosts/$ip/README.md.new hosts/$ip/README.md
      fi
    fi
  fi
done
