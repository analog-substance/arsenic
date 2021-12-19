

export REMOVE_DOMAIN_REGEX="(\._domainkey\.|hscoscdn10\.net|sites\.hubspot\.net|amazonaws\.com|azurewebsites\.net|cloudfront\.net|azurewebsites\.windows\.net|azure\.com|cloudapp\.net|readthedocs\.io|my\.jobs|1e100\.net|googlehosted\.com|readthedocs\.org|c7dc\.com|akamaitechnologies\.com)\$"

function _ {
  echo  "[+] $@"
}

function _warn {
  echo  '[!] '"$@"
}

function _info {
  echo "[-] $@"
}

function removeInvalidDomains {
  # remove *. prefix
  # remove email addr prefixes
  # remove IP addrs
  # remove IPv6 addrs
  # remove domain regex
  sed 's/^\*\.//g' \
  | sed 's/^[^@]\+@//g' \
  | sed 's/\.$//g' \
  | tr 'A-Z' 'a-z' \
  | grep -vP "^([0-9]{1,3}\.){3}[0-9]{1,3}\$" \
  | grep -vP "$REMOVE_DOMAIN_REGEX" \
  | grep -P '^[a-z0-9_\-\.]+$' \
  | arsenic scope prune \
  | sort -d | uniq
}

function gitPull {
  if [ $GIT -eq 1 ]; then
    if ! git pull --rebase > /dev/null 2>&1 ; then
      _warn "pull failed" >&2
    fi
  fi
}

function gitCommit {
  if [ $GIT -eq 1 ]; then
    path="$1"
    msg="$2"
    set +u
    mode="$3"
    set -u
    git add "$path"

    if git commit -m "$msg" ; then
      if ! git push ; then
        _warn "First push failed"
        if ! git pull --rebase ; then
          echo '[!] pull rebase failed'
          if [ "$mode" == "reset" ] ; then
            echo '[!] reset to origin'
            git rebase --abort
            git reset --hard origin/master
          fi
          exit 2
        fi
        git push
      fi
    else
      echo "nothing happened"
    fi
  fi
}

function gitLock {
  # gitPull
  if [ $GIT -eq 1 ]; then
    if ! git pull --rebase > /dev/null 2>&1 ; then
        echo '[!] reset to origin'
        git rebase --abort
        git reset --hard origin/master
        exit 1
    fi
  fi

  if [ -f "$1" ]; then
    _warn "cant lock a file that exists"
    exit 1
  fi

  echo "lock::$(openssl rand 100 | base64 | tr -cd 'A-Za-z0-9' | cut -b1-16)" > "$1"
  gitCommit "$1" "$2" reset
}

function getAllDomains {
  # create a combined scope file
    arsenic scope domains \
    | tee scope-domains-generated-combined.txt.new
    mv scope-domains-generated-combined.txt.new scope-domains-generated-combined.txt
}

export GIT=1
if [ ! -d .git ]; then
  export GIT=0
fi
