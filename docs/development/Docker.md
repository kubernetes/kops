In order to develop inside a Docker container you must mount your local copy of 
the Kops repo into the container's `GOPATH`. For the official Golang Docker 
image this is simply a matter of running the following command:

```bash
docker run -it -v /path/to/local/kops/repo:/go/src/k8s.io/kops golang bash
```

You should now be able to test if everything is working by building the project 
using `make kops` or running the tests with `make test`. In order to simulate 
the tests ran on the CI server then use the target `make ci`.
