#!/bin/sh

VERSION=$(curl -s https://api.github.com/repos/drival-ai/v10-mvp-api/releases/latest | jq -r ".tag_name")
cd /tmp/ && wget https://github.com/drival-ai/v10-mvp-api/releases/download/$VERSION/v10-mvp-api-$VERSION-x86_64-linux.tar.gz
tar xvzf v10-mvp-api-$VERSION-x86_64-linux.tar.gz
sudo cp -v bin/v10-mvp-api /usr/local/bin/v10-mvp-api
sudo chown root:root /usr/local/bin/v10-mvp-api

sudo systemctl daemon-reload
sudo systemctl enable v10mvpapi
sudo systemctl start v10mvpapi
sudo systemctl status v10mvpapi

rm -rfv bin/
rm v10-mvp-api-$VERSION-x86_64-linux.tar.gz
