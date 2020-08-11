# Avi Kubernetes Operator

## Architecture

The Avi K8s Operator (AKO) is used to provide L4-L7 load balancing for applications deployed
in a kubernetes cluster for north-south traffic.

The AKO controller ingests the Kubernetes API server object updates
to construct corresponding objects in the Avi controller. The Avi controller then programs
the datapath using appropriate APIs to enable traffic routing for requested applications.

![Alt text](AKO_Arch.png?raw=true "Title")

## Run AKO

AKO runs as a POD inside the kubernetes cluster.

#### Pre-requisites

To Run AKO you need the following pre-requisites (on-prem clusters):

 - ***Step 1***: An Avi Controller with a vCenter cloud or No Access Cloud configured.

 - ***Step 2***: If your POD CIDRs are not routable:
 
    - Configure the vip network in the values.yaml of AKO.
    - Also provide the backend network information (from 1.2.1 onwards).
    
  For additional settings related to values.yaml pls refer [here](https://github.com/avinetworks/avi-helm-charts/blob/master/docs/AKO/values.md)

 - ***Step 2.1***: If your POD CIDRs are routable (or you are using the NodePort mode), you don't have to worry about the backend network.
 - ***Step 3:*** Kubernetes 1.16+.
 - ***Step 4:*** `helm` cli pointing to your kubernetes cluster.
 
 NOTE: We only support `helm` 3.0 and above. For a more detailed installation instruction pls refer [here](https://avinetworks.com/docs/ako/1.1/ako-installation/)

#### Install using *helm*

*Step 1:* Create the `avi-system` namespace:

    kubectl create ns avi-system

*Step 2:* Clone this repository, go inside the `helm` directory and run:

    helm install ./ako --name my-ako-release --namespace=avi-system --set configs.controllerIP=10.10.10.10

Use the `helm/ako/values.yaml` to edit values related to Avi configuration. A list of editable parameters can be found [here](https://github.com/avinetworks/avi-helm-charts/blob/master/docs/AKO/README.md#parameters)


#### Uninstall using *helm*

Simply run:

*Step1:*

    helm delete <ako-release-name> -n avi-system
 
*Step 2:* 

    kubectl delete ns avi-system


## Build and Test

AKO can be built as a docker container using the `make` command. Simply clone the repository
and run:

    make docker
    
AKO runs a simulation of the Kubernetes APIs using the kubernetes `FakeClient` and it also
simulates the Avi controller by exploiting the `httptest` server from golang. In order to run
the end to end unit tests, you can execute:

    make int_tests

    
## Contributing

We welcome new contributors to our repository. Following are the pre-requisties that should help
you get started:

* Before contributing, please get familiar with our
[Code of Conduct](CODE_OF_CONDUCT.md).
* Check out our [Contributor Guide](CONTRIBUTING.md) for information
about setting up your development environment and our contribution workflow.

## License

AKO is licensed under the [Apache License, version 2.0](LICENSE)
