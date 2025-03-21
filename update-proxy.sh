#!/bin/sh

exec 200>/tmp/proxy.lock || exit 1
flock -w 10 200 || exit 1
trap 'rm -f /tmp/proxy.lock' EXIT

PROXY_VERSION=$(curl -s https://api.github.com/repos/drival-ai/v10-api-proxy/releases/latest | jq -r ".tag_name")
cd /tmp/ && wget https://github.com/drival-ai/v10-api-proxy/releases/download/$PROXY_VERSION/v10-api-proxy-$PROXY_VERSION-x86_64-linux.tar.gz
tar xvzf v10-api-proxy-$PROXY_VERSION-x86_64-linux.tar.gz
sudo systemctl stop v10apiproxy
sudo cp -v v10-api-proxy /usr/local/bin/
sudo chown root:root /usr/local/bin/v10-api-proxy

sudo systemctl daemon-reload
sudo systemctl enable v10apiproxy
sudo systemctl start v10apiproxy
sudo systemctl status v10apiproxy

rm v10-api-proxy
rm v10-api-proxy-$PROXY_VERSION-x86_64-linux.tar.gz
