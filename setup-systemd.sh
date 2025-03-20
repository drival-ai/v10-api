#!/bin/sh

# Setup the main API binary:

systemctl stop v10mvpapi

mkdir -p /etc/v10-mvp-api/
VERSION=$(curl -s https://api.github.com/repos/drival-ai/v10-mvp-api/releases/latest | jq -r ".tag_name")
cd /tmp/ && wget https://github.com/drival-ai/v10-mvp-api/releases/download/$VERSION/v10-mvp-api-$VERSION-x86_64-linux.tar.gz
tar xvzf v10-mvp-api-$VERSION-x86_64-linux.tar.gz
cp -v bin/v10-mvp-api /usr/local/bin/
chown root:root /usr/local/bin/v10-mvp-api
cp -v bin/setup-systemd.sh /etc/v10-mvp-api/
cp -v bin/update-systemd.sh /etc/v10-mvp-api/
chmod +x /etc/v10-mvp-api/update-systemd.sh
aws s3 cp s3://drival-mvp-api/postgres .
cp -v postgres /etc/v10-mvp-api/ && rm postgres

cat >/usr/lib/systemd/system/v10mvpapi.service <<EOL
[Unit]
Description=V10 MVP API

[Service]
Type=simple
Restart=always
RestartSec=10
ExecStart=/usr/local/bin/v10-mvp-api -logtostderr

[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
systemctl enable v10mvpapi
systemctl start v10mvpapi
systemctl status v10mvpapi

rm -rfv bin/
rm v10-mvp-api-$VERSION-x86_64-linux.tar.gz

# Setup API proxy:

systemctl stop v10mvpapiproxy

PROXY_VERSION=$(curl -s https://api.github.com/repos/drival-ai/v10-mvp-api-proxy/releases/latest | jq -r ".tag_name")
cd /tmp/ && wget https://github.com/drival-ai/v10-mvp-api-proxy/releases/download/$PROXY_VERSION/v10-mvp-api-proxy-$PROXY_VERSION-x86_64-linux.tar.gz
tar xvzf v10-mvp-api-proxy-$PROXY_VERSION-x86_64-linux.tar.gz
cp -v v10-mvp-api-proxy /usr/local/bin/
chown root:root /usr/local/bin/v10-mvp-api-proxy

cat >/usr/lib/systemd/system/v10mvpapiproxy.service <<EOL
[Unit]
Description=V10 MVP API Proxy

[Service]
Type=simple
Restart=always
RestartSec=10
ExecStart=/usr/local/bin/v10-mvp-api-proxy -logtostderr

[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
systemctl enable v10mvpapiproxy
systemctl start v10mvpapiproxy
systemctl status v10mvpapiproxy

rm v10-mvp-api-proxy
rm v10-mvp-api-proxy-$PROXY_VERSION-x86_64-linux.tar.gz
