#!/bin/sh

SAMPLE_VERSION=$(curl -s https://api.github.com/repos/drival-al/v10-mvp-api/releases/latest | jq -r ".tag_name")
cd /tmp/ && wget https://github.com/drival-al/v10-mvp-api/releases/download/$SAMPLE_VERSION/v10-mvp-api-$SAMPLE_VERSION-x86_64-linux.tar.gz
tar xvzf v10-mvp-api-$SAMPLE_VERSION-x86_64-linux.tar.gz
cp -v v10-mvp-api /usr/local/bin/v10-mvp-api
chown root:root /usr/local/bin/v10-mvp-api

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
