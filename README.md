# Avi Kubernetes Operator

## Architecture

The Avi K8s Operator (AKO) is a layered collection of independent interoperable units
that are used in conjunction to provide L4-L7 load balancing for applications deployed
in a kubernetes cluster for north-south traffic.

The controller ingests the Kubernetes API server object updates namely services and ingress
to construct corresponding objects in Avi controller. Hence the AKC performs the dual
functionality of an ingress controller and a layer 4 load balancer.

![Alt text](AKO.jpg?raw=true "Title")

## Run AKO

AKO runs as a POD inside the kubernetes cluster.

#### Pre-requisites

To Run AKO you need the following pre-requisites:
 - ***Step 1***: An Avi Controller with a vCenter cloud.

 - ***Step 2***: If your POD CIDRs are not routable:
    - Create a VRF context object in Avi for the kubernetes controller.
    - Get the name the PG network which the kubernetes nodes are part of. 
    
      *NOTE: If you are using AKO for test puposes you can use the `global` vrf but you cannot manage multiple kubernetes clusters in the same cloud with this setting.*
    - Configure this PG network with the vrf mentioned in the previous step using the Avi CLI.
    - Make sure this PG network is part of the NS IPAM configured in the vCenter cloud.

 - ***Step 2.1***: If your POD CIDRs are routable then you can skip step 2. Ensure that you skip static route syncing in this case using the `disableStaticRouteSync` flag in the `values.yaml` of your helm chart.
 - ***Step 3:*** Kubernetes 1.14+.
 - ***Step 4:*** `helm` cli pointing to your kubernetes cluster.

#### Install using *helm*

*Step 1:* Create the `avi-system` namespace:

    kubectl create ns avi-system


*Step 2:* Clone this repository, go inside the `helm` directory and run:

    helm install ./ako --name my-ako-release --namespace=avi-system --set configs.controllerIP=10.10.10.10

Use the `helm/ako/values.yaml` to edit values related to Avi configuration. For information regarding configurable parameters to be provided in `values.yaml` during AKO install, please refer [AKO helm chart params](https://github.com/avinetworks/avi-helm-charts#parameters)

Detailed descriptions for these params are provided in [Description of tunables of AKO](https://github.com/avinetworks/avi-helm-charts/blob/master/docs/values.md)


#### Uninstall using *helm*

Simply run:


*Step1:*

    helm delete my-ako-release -n avi-system
 
*Step 2:* 

    kubectl delete ns avi-system


## Build and Test

AKO can be built as a docker container using the `make` command. Simply clone the repository
and run:

    make docker
    
Unit tests can be run using:

    make test and make int_test

