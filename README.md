[![main](https://github.com/drival-ai/v10-api/actions/workflows/main.yml/badge.svg)](https://github.com/drival-ai/v10-api/actions/workflows/main.yml)

Update binaries with the latest tag:

```sh
# Update the API binary:
$ ssh -i key.pem ec2-user@0.0.0.0 -t '/etc/v10-api/update-api.sh'

# Update the proxy binary:
$ ssh -i key.pem ec2-user@0.0.0.0 -t '/etc/v10-api/update-proxy.sh'
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
