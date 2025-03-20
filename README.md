Update API binary with the latest tag:

```sh
$ ssh -i key.pem \
  ec2-user@0.0.0.0 \
  -t 'sudo systemctl stop v10mvpapi && /etc/v10-mvp-api/update-systemd.sh'
```
