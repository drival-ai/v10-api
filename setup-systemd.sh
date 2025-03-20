#!/bin/sh

# Setup the main API binary:

systemctl stop v10api

mkdir -p /etc/v10-api/
VERSION=$(curl -s https://api.github.com/repos/drival-ai/v10-api/releases/latest | jq -r ".tag_name")
cd /tmp/ && wget https://github.com/drival-ai/v10-api/releases/download/$VERSION/v10-api-$VERSION-x86_64-linux.tar.gz
tar xvzf v10-api-$VERSION-x86_64-linux.tar.gz
cp -v bin/v10-api /usr/local/bin/
chown root:root /usr/local/bin/v10-api
cp -v bin/setup-systemd.sh /etc/v10-api/
cp -v bin/update-systemd.sh /etc/v10-api/
chmod +x /etc/v10-api/update-systemd.sh
aws s3 cp s3://drival-mvp-api/postgres .
cp -v postgres /etc/v10-api/ && rm postgres

cat >/usr/lib/systemd/system/v10api.service <<EOL
[Unit]
Description=V10 API

[Service]
Type=simple
Restart=always
RestartSec=10
ExecStart=/usr/local/bin/v10-api -logtostderr

[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
systemctl enable v10api
systemctl start v10api
systemctl status v10api

rm -rfv bin/
rm v10-api-$VERSION-x86_64-linux.tar.gz

# Setup API proxy:

systemctl stop v10apiproxy

PROXY_VERSION=$(curl -s https://api.github.com/repos/drival-ai/v10-api-proxy/releases/latest | jq -r ".tag_name")
cd /tmp/ && wget https://github.com/drival-ai/v10-api-proxy/releases/download/$PROXY_VERSION/v10-api-proxy-$PROXY_VERSION-x86_64-linux.tar.gz
tar xvzf v10-api-proxy-$PROXY_VERSION-x86_64-linux.tar.gz
cp -v v10-api-proxy /usr/local/bin/
chown root:root /usr/local/bin/v10-api-proxy

cat >/usr/lib/systemd/system/v10apiproxy.service <<EOL
[Unit]
Description=V10 API Proxy

[Service]
Type=simple
Restart=always
RestartSec=10
ExecStart=/usr/local/bin/v10-api-proxy -logtostderr

[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
systemctl enable v10apiproxy
systemctl start v10apiproxy
systemctl status v10apiproxy

rm v10-api-proxy
rm v10-api-proxy-$PROXY_VERSION-x86_64-linux.tar.gz
