## Passing in additional user-data to cloud-init

Kops utilizes cloud-init to initialize and setup a host at boot time. However in certain cases you may already be leaveraging certain features of cloud-init in your infrastructure and would like to continue doing so. More information on cloud-init can be found [here](http://cloudinit.readthedocs.io/en/latest/)

### Adding additonal user-data

Aditional user-user data can be passed to the host provissioning by utilizing the environment variable EXTRA_USER_DATA. By setting `EXTRA_USER_DATA` to `<file>:<content-type>` kops will load these aditional files and pass them on to the hosts. A list of valid content-types can be found [here](http://cloudinit.readthedocs.io/en/latest/topics/format.html#mime-multi-part-archive) 

### Example
By exporting `EXTRA_USER_DATA` like this, will cause the files described bellow to get loaded.
```
export NAME=myfirstcluster.example.com
export KOPS_STATE_STORE=s3://prefix-example-com-state-store
export EXTRA_USER_DATA="myscript.sh:text/x-shellscript local_repo.txt:text/cloud-config"
kops create cluster \
    --zones us-west-2a \
    ${NAME}
```

`myscript.sh`
```
#!/bin/sh
echo "Hello World.  The time is now $(date -R)!" | tee /root/output.txt
  ```

`local_repo.txt`
```
#cloud-config
apt:
  primary:
    - arches: [default]
      uri: http://local-mirror.mydomain
      search:
        - http://local-mirror.mydomain
        - http://archive.ubuntu.com
  ```
