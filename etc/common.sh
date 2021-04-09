
function _ {
  echo  "[+] $@"
}

function _warn {
  echo  '[!] '"$@"
}

function _info {
  echo "[-] $@"
}

function ensureInScope {
  in_scope=$(echo $(cat scope-domains.txt | sed 's/\./\\./g;s/^/(.+\\.)?/g') | sed 's/ /|/g')
  if [[ -n "$in_scope" ]]; then
    grep -P "^$in_scope$"
  else
    grep -P '.*'
  fi
}

function removeInvalidThings {
  # remove *. prefix
  # remove email addr prefixes
  # remove IP addrs
  # remove IPv6 addrs
  # remove domain regex
  sed 's/^\*\.//g' \
  | sed 's/^[^@]\+@//g' \
  | tr 'A-Z' 'a-z' \
  | grep -vP "^([0-9]{1,3}\.){3}[0-9]{1,3}$" \
  | grep -vP "$REMOVE_DOMAIN_REGEX" \
  | grep -P '^[a-z0-9_\-\.]+$' \
  | as-prune-blacklisted-domains \
  | sort -h | uniq
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
      if ! git pull --rebase ; then
        echo '[!] pull rebase failed'
        if [ "$mode" == "reset" ] ; then
          echo '[!] reset to origin'
          git reset --hard origin/master
          exit 2
        fi
        echo '[!] not sure what to do. i guess i will git add and hope it works'
        git commit -am "not sure what this is"
        git pull --rebase
      fi
      git push
    else
      echo "nothing happened"
    fi
  fi
}

function gitLock {
  echo lock > "$1"
  gitCommit "hosts/$host/recon/nmap-punched-udp.nmap" "new host: $host" reset
}

export GIT=1
if [ ! -d .git ]; then
  export GIT=0
fi
