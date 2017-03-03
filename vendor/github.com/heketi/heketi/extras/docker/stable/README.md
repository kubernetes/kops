# Heketi Docker

This is the official version of the Heketi docker container based on CentOS 7.  To run, simply type:

```
$ docker run -d -p 8080:8080 heketi/heketi
$ curl http://localhost:8080/hello
```

This will run Heketi with a mock executor allowing you to experiment with the Heketi API.  The database is stored inside the container and will not be saved.

# Custom deployment 

To provide your own configuration file, you must create your own alternative configuration file in a directory on the host machine and then mount that directory location as `/etc/heketi` inside the container.

It is also advised to save the database in a directory on the host and mouting that directory to the directory location `/var/lib/heketi` inside the container.

Here is an example:

```
$ mkdir config
$ cp custom_heketi.json config/heketi.json
$ mkdir db
$ chmod 777 db
$ docker run -d -p 8080:8080 \
   --volume $PWD/config:/etc/heketi \
   --volume $PWD/db:/var/lib/heketi \
   heketi/heketi
```

# For Developers
To build:

```
# docker build --rm --tag <username>/heketi:centos7 .
```


To run:

    # docker run -d -p 8080:8080 <username>/heketi:centos7
