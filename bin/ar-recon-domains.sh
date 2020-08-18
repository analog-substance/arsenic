#! /bin/bash

mkdir -p recon/domains hosts

declare -a domain_commands=("whois" "host" "ar-crtsh-slrp")
declare -a ip_commands=("whois")

function domain_recon {
  domain=$(echo $1 | tr 'A-Z' 'a-z')
  case $domain in
    *"azurewebsites.net"|*"my.jobs"|*"azure-mobile.net"|*"readthedocs.io"|*"cloudflaressl.com"|*"amazonaws.com"|*"cloudflare-dns.com")
      echo "[!] Skipping common domain: $domain"
      ;;
    *)
      for cmd in "${domain_commands[@]}";  do
        if [ ! -f "recon/domains/$domain-$cmd.txt" ]; then
          echo "[+] running $domain $cmd"

          $cmd $domain | tr 'A-Z' 'a-z' > "recon/domains/$domain-$cmd.txt" &
        else
          echo "[!] skipping $domain $cmd"
        fi
      done
      wait
    esac
}

function ip_recon {
  ip=$1
  case $ip in
    "1.1.1.1"|"1.0.0.1"|10.*)
      echo "Ignore"
      ;;
    *)
      mkdir -p "hosts/$ip/recon/"
      for cmd in "${ip_commands[@]}";  do
        if [ ! -f "hosts/$ip/recon/$cmd.txt" ]; then
          echo "[+] running $ip $cmd"

          $cmd $ip > "hosts/$ip/recon/$cmd.txt" &
        else
          echo "[!] skipping $ip $cmd"
        fi
      done
      wait
      ;;
  esac
}


REMOVE_DOMAIN_REGEX="(hscoscdn10\.net|sites\.hubspot\.net|amazonaws\.com|azurewebsites\.net|azurewebsites\.windows\.net|cloudapp\.net|readthedocs\.io|my\.jobs|googlehosted\.com|readthedocs\.org)$"
cat scope-domains-* | sort -u | while read domain; do
  domain_recon "$domain"
done

cat recon/domains/*-host.txt | grep address | awk '{print $NF}' | sort -u | tee scope-ips-client-domains.txt

cat scope-ips-client-domains.txt | while read ip; do
  ip_recon "$ip"
done

cat recon/domains/*-ar-crtsh-slrp.txt | tr 'A-Z' 'a-z' | sort -u | grep -vP "$REMOVE_DOMAIN_REGEX" | tee scope-domains-ar-crtsh-slrp.txt

cat scope-domains* | sort -u | while read domain; do
  domain_recon "$domain"
done

cat recon/domains/*-host.txt | grep address | awk '{print $NF}' | sort -u | grep -vP "^($(echo $(cat scope-ips-client-domains.txt) | sed 's/ /|/g'))$" | grep -vP "1\.1\.1\.1|1\.0\.0\.1" |  tee scope-ips-discovered-domains.txt

cat scope-ips-* | sort -u | while read ip; do
  ip_recon "$ip"
done

cat scope-ips-* | sort -u > recon/all-ips.txt

nmap -p443,8443 -iL recon/all-ips.txt -oA recon/nmap-https-check-all-ips --open

cat recon/nmap-https-check-all-ips.gnmap|grep Ports: | awk '{print $2,$5}' | cut -d/ -f1 | while read l; do
  host=$(echo $l | awk '{print $1}');
  port=$(echo $l | awk '{print $2}');
  ar-get-cert-domains.sh $host $port > "hosts/$host/recon/ar-get-cert-domains.txt"
done

cat hosts/*/recon/ar-get-cert-domains.txt | tr 'A-Z' 'a-z' | sed 's/\*\.//g' | sort -u | tee scope-domains-from-certs.txt
cat scope-domains* | sort -u | grep -vP "^([0-9]{1,3}\.){3}([0-9]{1,3})" \
| grep -vP "^($(echo $(cat scope-domains-client-provided.txt) | sed 's/ /|/g'))$" \
| grep -vP "$REMOVE_DOMAIN_REGEX" \
| tee scope-domains-discovered.txt

cat scope-domains* | sort -u | while read domain; do
  domain_recon "$domain"
done

cat recon/domains/*-host.txt | grep address | awk '{print $NF}' | sort -u | grep -vP "^($(echo $(cat scope-ips-client-domains.txt) | sed 's/ /|/g'))$" | grep -vP "1\.1\.1\.1|1\.0\.0\.1" |  tee scope-ips-discovered-domains.txt

cat scope-ips-* | sort -u | while read ip; do
  ip_recon "$ip"
done

cat scope-domains-discovered.txt | while read d; do
  if [ -f recon/domains/$d-host.txt ]; then
    cat recon/domains/$d-host.txt
  fi
done \
| grep -P "(is an alias|has (IPv6 )?address)" \
| awk '{print $1}' \
| grep -vP "$REMOVE_DOMAIN_REGEX" \
| sort -u > scope-domains-discovered-valid.txt


cat scope-domains-discovered-valid.txt | while read d; do
    cat recon/domains/$d-host.txt
done

cat scope-domains-discovered-valid.txt | wc -l
