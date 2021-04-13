

export REMOVE_DOMAIN_REGEX="(\._domainkey\.|hscoscdn10\.net|sites\.hubspot\.net|amazonaws\.com|azurewebsites\.net|cloudfront\.net|azurewebsites\.windows\.net|azure\.com|cloudapp\.net|readthedocs\.io|my\.jobs|googlehosted\.com|readthedocs\.org|c7dc\.com|akamaitechnologies\.com)\$"
# Right now just gonna ignore these.
export NON_ROOT_DOMAIN_REGEX="co\.|com\.|herokuapp\."

function _ {
  echo  "[+] $@"
}

function _warn {
  echo  '[!] '"$@"
}

function _info {
  echo "[-] $@"
}

function ensureDomainInScope {
  in_scope=$(echo $(cat scope-domains.txt | sed 's/\./\\./g;s/^/(.+\\.)?/g') | sed 's/ /|/g')
  if [[ -n "$in_scope" ]]; then
    grep -P "^$in_scope\$"
  else
    grep -P '.*'
  fi
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

function getRootDomains {
  ## Lets get a unique list of root domains
  # cat all domains
  # remove *. prefix
  # remove email addr prefixes
  # remove problematic domains
  # print last 2 octets in the domain
  # remove things like co.uk, com.uk
  cat scope-domains* \
  | removeInvalidDomains \
  | awk -F. '{print $(NF-1) "." $NF}' \
  | grep -vP "$NON_ROOT_DOMAIN_REGEX" \
  | sort -h | uniq \
  | tee scope-domains-generated-root.txt
}

function getAllDomains {
  # create a combined scope file
    cat scope-domains* \
    | removeInvalidDomains \
    | cat - scope-domains.txt \
    | sort -h |  uniq \
    | tee scope-domains-generated-combined.txt
}

export GIT=1
if [ ! -d .git ]; then
  export GIT=0
fi
