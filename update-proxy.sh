#!/bin/sh

PROXY_VERSION=$(curl -s https://api.github.com/repos/drival-ai/v10-api-proxy/releases/latest | jq -r ".tag_name")
cd /tmp/ && wget https://github.com/drival-ai/v10-api-proxy/releases/download/$PROXY_VERSION/v10-api-proxy-$PROXY_VERSION-x86_64-linux.tar.gz
tar xvzf v10-api-proxy-$PROXY_VERSION-x86_64-linux.tar.gz
systemctl stop v10apiproxy
cp -v v10-api-proxy /usr/local/bin/
chown root:root /usr/local/bin/v10-api-proxy

systemctl daemon-reload
systemctl enable v10apiproxy
systemctl start v10apiproxy
systemctl status v10apiproxy

rm v10-api-proxy
rm v10-api-proxy-$PROXY_VERSION-x86_64-linux.tar.gz
