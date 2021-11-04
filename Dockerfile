# Defektive hates Mac OS so you can build and exec into this.
# docker build -t arsenic:ubuntu18.04 .
# docker run --rm -v $PWD:/pwd -d --name arsenic -i arsenic:ubuntu18.04
# docker exec -it arsenic /bin/bash

FROM ubuntu:18.04
ENV LC_CTYPE C.UTF-8
ENV DEBIAN_FRONTEND=noninteractive
ENV GOROOT="/usr/local/go"
ENV GOPATH="/root/go"
ENV PATH="/root/go/bin:/usr/local/go/bin:${PATH}"

RUN apt-get update && \
apt-get install -y build-essential jq curl wget rubygems gcc dnsutils netcat net-tools vim python python3 python3-pip python3-dev libssl-dev libffi-dev git make python-pip curl nmap sed grep figlet libunbound-dev whois tar

RUN cd /var/opt/ && \
git clone https://github.com/michenriksen/aquatone && \
git clone https://github.com/analog-arsenic/arsenic && \
git clone https://github.com/analog-arsenic/hugo && \
git clone https://github.com/OJ/gobuster.git

RUN cd /tmp && \
wget https://dl.google.com/go/go1.15.6.linux-amd64.tar.gz && \
tar -xvf go1.15.6.linux-amd64.tar.gz && \
mv go /usr/local && \
echo "export GOROOT=/usr/local/go" ~/.bashrc && \
echo "export GOPATH=$HOME/go" >> ~/.bashrc && \
echo "export PATH=$GOPATH/bin:$GOROOT/bin:$PATH" >> ~/.bashrc && \
echo "source /var/opt/arsenic/arsenic.rc" >> ~/.bashrc

RUN go get github.com/analog-arsenic/fast-resolv
RUN cp /root/go/src/github.com/analog-arsenic/fast-resolv/fast-resolv.conf /root/go/bin/
