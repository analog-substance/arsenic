#! /bin/bash

function _ {
  echo "[+] $@"
}

TICK=("-" "/" "|" "\\")
TICKS=0
function tick {
  echo -ne "\r[${TICK[$TICKS]}] $@" >&2
  TICKS=$(( TICKS + 1 ))
  if [ $TICKS -gt 3 ]; then
    TICKS=0
  fi
}

_ "Deleting previous detection attempt"
rm -rf recon/detect_services 2>/dev/null
sync
mkdir -p recon/detect_services/services

# Look at resolved domains
# determine what domains have the same ip
# determine what ips have the same domains
# isolate by service

cat recon/domains/resolv-domains.txt | grep "address" | sort -h | uniq | while read line; do
  tick "Reviewing resolved domains"

  domain=$(echo "$line" | awk '{print $1}')
  ip=$(echo "$line" | awk '{print $NF}')
  ip_resolv_domain=$(grep -P "^$(echo "$ip" | sed 's/\./\\./g') " recon/ips/resolv-ips.txt | awk '{print $NF}' | sed 's/\.$//g')

  if echo "$ip_resolv_domain" | grep cloudfront.net > /dev/null; then
    ip_resolv_domain="r.cloudfront.net"
  else
    ip_resolv_domain=""
  fi

  ip_resolv_domain_file="recon/detect_services/resolv-domain-$ip_resolv_domain.txt"
  domain_file="recon/detect_services/resolv-domain-$domain.txt"
  ip_file="recon/detect_services/resolv-ip-$ip.txt"


  {
    if [ -f "$ip_file" ]; then
      cat "$ip_file"
    fi

    if [ -n "$ip_resolv_domain" ]; then
      echo $ip_resolv_domain
    fi
    echo "$domain"
  } | sort -h | uniq > "$ip_file.new"
  mv "$ip_file.new" "$ip_file"

  {
    if [ -f "$domain_file" ]; then
      cat "$domain_file"
    fi
    echo "$ip"
  } | sort -h | uniq > "$domain_file.new"
  mv "$domain_file.new" "$domain_file"


  if [ -n "$ip_resolv_domain" ]; then
    {
      if [ -f "$ip_resolv_domain_file" ]; then
        cat "$ip_resolv_domain_file"
      fi
      echo "$ip"
    } | sort -h | uniq > "$ip_resolv_domain_file.new"
    mv "$ip_resolv_domain_file.new" "$ip_resolv_domain_file"
  fi
done

cat "recon/detect_services/resolv-domain-r.cloudfront.net.txt" | while read ip; do
  cat "recon/detect_services/resolv-ip-$ip.txt"
  # rm "recon/detect_services/resolv-ip-$ip.txt"
done | sort -h | uniq > recon/detect_services/cloudfront-domains.txt

cat recon/detect_services/cloudfront-domains.txt | while read domain; do
  cat "recon/detect_services/resolv-domain-$domain.txt" | while read ip; do
    {
      cat "recon/detect_services/resolv-ip-$ip.txt"
      cat recon/detect_services/cloudfront-domains.txt
    } | sort -h | uniq > "recon/detect_services/resolv-ip-$ip.txt.new"
    mv "recon/detect_services/resolv-ip-$ip.txt.new" "recon/detect_services/resolv-ip-$ip.txt"
  done
done


first_cf_domain=$(cat recon/detect_services/cloudfront-domains.txt | head -n1)
first_cf_domain_ip=$(cat "recon/detect_services/resolv-domain-$first_cf_domain.txt" | head -n1)
first_cf_file="recon/detect_services/resolv-ip-$first_cf_domain_ip.txt"
{
  cat recon/detect_services/cloudfront-domains.txt
  cat $first_cf_file
} | sort -h | uniq > "$first_cf_file.new"
mv "$first_cf_file.new" "$first_cf_file"

# cat "recon/detect_services/resolv-domain-r.cloudfront.net.txt" | while read ip; do
#   # cat "recon/detect_services/resolv-ip-$ip.txt"
#   rm "recon/detect_services/resolv-ip-$ip.txt"
# done

echo
_ "Domain review complete"

PRIVATE_IP_REGEX="\b(127\.[0-9]{1,3}\.|10\.[0-9]{1,3}\.|192\.168\.|172\.(1[6-9]|2[0-9]|3[01])\.)[0-9]{1,3}\.[0-9]{1,3}\b"
ls recon/detect_services/resolv-ip* | sort -h | uniq | grep -vP "$PRIVATE_IP_REGEX" | while read ip_file; do
  cat "$ip_file" | while read domain; do
    tick "Reviewing resolved IPs"

    if [ -n "$last_domain" ] ; then
      first_domain=$( head -n 1 "$ip_file" )
      if diff "recon/detect_services/resolv-domain-$first_domain.txt" "recon/detect_services/resolv-domain-$domain.txt" > /dev/null ; then
        # echo "$first_domain $domain no diff"
        nodif=1
      else
        diff_file="recon/detect_services/services/$first_domain/domains-with-resolv-differences"
        # echo "$domain has resolv differences, but shares some with $first_domain"
        # echo "This coulld mean they point to a CDN, or DDOS protection service."
        {
          if [ -f "$diff_file" ]; then
            cat "$diff_file"
          fi
          echo "$domain"
        } | sort -h | uniq > "$diff_file.new"
        mv "$diff_file.new" "$diff_file"


        {
          cat "$diff_file"
          cat "recon/detect_services/services/$first_domain/recon/other-hostnames.txt"
        } | sort -h | uniq > "recon/detect_services/services/$first_domain/recon/other-hostnames.txt.new"
        mv "recon/detect_services/services/$first_domain/recon/other-hostnames.txt.new" "recon/detect_services/services/$first_domain/recon/other-hostnames.txt"

      fi
    else
      mkdir -p "recon/detect_services/services/$domain/recon"
      cp "$ip_file" "recon/detect_services/services/$domain/recon/other-hostnames.txt"
      cp "recon/detect_services/resolv-domain-$domain.txt" "recon/detect_services/services/$domain/recon/other-ips.txt"
    fi
    last_domain="$domain"
  done
done

echo
_ "Updating existing hosts"

ls -d recon/detect_services/services/* | cut -d/ -f4 | while read service ; do

  if [ -e "hosts/$service" ]; then

    _ "Updating existing $service"
    {
      cat "hosts/$service/recon/hostnames.txt"
      cat "recon/detect_services/services/$service/recon/other-hostnames.txt"
    } | sort -h | uniq > "hosts/$service/recon/hostnames.txt.new"
    mv "hosts/$service/recon/hostnames.txt.new" "hosts/$service/recon/hostnames.txt"
  else
    # $service doesn't exist, lelts see if the domains
    if grep -P "^($(echo $(cat "recon/detect_services/services/$service/recon/other-hostnames.txt") | sed 's/\./\\./g;s/ /|/g'))\$" hosts/*/recon/*hostnames.txt > /dev/null ; then

      exsting_service=$(grep -P "^($(echo $(cat "recon/detect_services/services/$service/recon/other-hostnames.txt") | sed 's/\./\\./g;s/ /|/g'))\$" hosts/*/recon/*hostnames.txt \
      | cut -d/ -f2 | sort -h | uniq | head -n1)
      _ "Adding domains to $exsting_service from $service"

      {
        cat "hosts/$exsting_service/recon/hostnames.txt"
        cat "recon/detect_services/services/$service/recon/other-hostnames.txt"
      } | sort -h | uniq > "hosts/$exsting_service/recon/hostnames.txt.new"
      mv "hosts/$exsting_service/recon/hostnames.txt.new" "hosts/$exsting_service/recon/hostnames.txt"
    else
      # no existing domains found lets create a new service
      _ "Creating new service $service"
      mv "recon/detect_services/services/$service/" "hosts/$service"
      cp "hosts/$service/recon/other-hostnames.txt" "hosts/$service/recon/hostnames.txt"
    fi

  fi
done
