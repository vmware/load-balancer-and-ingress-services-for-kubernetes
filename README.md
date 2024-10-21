# Load Balancer and Ingress Services for Kubernetes

## Architecture

The Avi Kubernetes Operator (AKO) is used to provide L4-L7 load balancing for applications deployed
in a kubernetes cluster for north-south traffic.

The AKO controller ingests the Kubernetes API server object updates
to construct corresponding objects in the Avi controller. The Avi controller then programs
the datapath using appropriate APIs to enable traffic routing for requested applications.

![Alt text](ako_arch.png?raw=true "Title")


## Documentation

Take a look at the following documentation for instructions on installing [AKO - Avi Kubernetes Operator](docs/README.md)


## Contributing

We welcome new contributorss to our repository. Following are the pre-requisties that should help
you get started:

* Before contributing, please get familiar with our
[Code of Conduct](CODE_OF_CONDUCT.md).
* Check out our [Contributor Guide](CONTRIBUTING.md) for information
about setting up your development environment and our contribution workflow.
* Check out [Open Issues](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/issues).
* [ako-dev](https://groups.google.com/g/ako-dev) to participate in discussions on AKO's development.


## License

AKO is licensed under the [Apache License, version 2.0](LICENSE)
