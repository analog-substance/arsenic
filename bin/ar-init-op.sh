#! /bin/bash

OP_PATH="$OP_PATH"

set -euo pipefail

DEFAULT_OP="op"
OP_ARG=${1:-}

if [ ! -z "$OP_ARG" ]; then
  DEFAULT_OP="$OP_ARG"
fi

if [ -z "$OP_PATH" ]; then
  OP_PATH="$DEFAULT_OP"
fi

function _ {
  echo "[+] $@"
}

ARSENIC_PATH="$( cd "$(dirname "$0")/../" >/dev/null 2>&1 ; pwd -P )"
ARSENIC_OPT_PATH=$(dirname $ARSENIC_PATH)

export OP_NAME=$(basename "$OP_PATH")

_ "Creating op: $OP_NAME"

mkdir -p "$OP_PATH"
cd "$OP_PATH"

mkdir -p apps bin report/{findings,sections,static} hosts recon/domains
touch {apps,bin,recon}/.keep report/static/.keep

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


ls -d $ARSENIC_OPT_PATH/*/scripts/ar-init-op.sh 2>/dev/null | while read hook; do
  echo "[+] running $hook"
  bash "$hook"
done
