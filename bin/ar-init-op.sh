#! /bin/bash

if [ -z "$OP_PATH" ]; then
  if [ ! -z "$1" ]; then
    OP_PATH="$1"
  else
    # clever i know
    OP_PATH="op"
  fi
fi

function _ {
  echo "[+] $@"
}

ARSENIC_PATH="$( cd "$(dirname "$0")/../" >/dev/null 2>&1 ; pwd -P )"
OP_NAME=$(basename "$OP_PATH")

_ "Creating op: $OP_NAME"

mkdir -p "$OP_PATH"
cd "$OP_PATH"

mkdir -p apps bin report/{findings,sections,static} hosts recon/domains
touch {apps,bin,recon}/.keep

mkdir -p hosts/127.0.0.1/recon
touch hosts/127.0.0.1/README.md

_ "Setup hugo"
git clone https://github.com/defektive/arsenic-hugo.git

rm -rf arsenic-hugo/.git
mv arsenic-hugo/example .hugo
mkdir .hugo/themes
mv arsenic-hugo .hugo/themes/arsenic

mv .hugo/README.md report/sections/
ln -s report/sections/README.md

cd .hugo
mv config.toml ../
ln -s ../config.toml
mv sample-finding ../report/findings/first-finding

cd content
ln -srf ../../recon
ln -srf ../../hosts
ls -d ../../report/* | xargs -n 1 ln -srf

_ "Hugo Setup complete"

cd ../../

if [ ! -f Makefile ]; then
  {
    echo -e "report::\n\tcd .hugo; \\"
    echo -e "\thugo server"
  } >> Makefile
fi
