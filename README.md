[![main](https://github.com/drival-ai/v10-mvp-api/actions/workflows/main.yml/badge.svg)](https://github.com/drival-ai/v10-mvp-api/actions/workflows/main.yml)

Update API binary with the latest tag:

```sh
$ ssh -i key.pem ec2-user@0.0.0.0 \
  -t 'sudo systemctl stop v10mvpapi && /etc/v10-mvp-api/update-systemd.sh'
```

Setup systemd service (only needed for the first time):

```sh
# From local env:
$ scp -i key.pem setup-systemd.sh ec2-user@0.0.0.0:/tmp/
$ ssh -i key.pem ec2-user@0.0.0.0

# From within VM:
$ sudo su
$ cd /tmp/
$ chmod +x setup-systemd.sh
$ ./setup-systemd.sh
```
