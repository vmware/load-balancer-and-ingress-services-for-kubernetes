# Avi Kubernetes Controller

## Architecture

The Avi K8s Controller (AKC) is a layered collection of independent interoperable units
that are used in conjunction to provide L4-L7 load balancing for applications deployed
in a kubernetes cluster for north-south traffic.

The controller ingests the Kubernetes API server object updates namely services and ingress
to construct corresponding objects in Avi controller. Hence the AKC performs the dual
functionality of an ingress controller and a layer 4 load balancer.

![Alt text](AKC.jpg?raw=true "Title")

## Run AKC

AKC runs as a POD inside the kubernetes cluster.

#### Pre-requisites

To Run AKC you need the following pre-requisites:
 - ***Step 1***: An Avi Controller with a vCenter cloud.

 - ***Step 2***: If your POD CIDRs are not routable:
    - Create a VRF context object in Avi for the kubernetes controller.
    - Get the name the PG network which the kubernetes nodes are part of. 
    
      *NOTE: If you are using AKC for test puposes you can use the `global` vrf but you cannot manage multiple kubernetes clusters in the same cloud with this setting.*
    - Configure this PG network with the vrf mentioned in the previous step using the Avi CLI.
    - Make sure this PG network is part of the NS IPAM configured in the vCenter cloud.

 - ***Step 2.1***: If your POD CIDRs are routable then you can skip step 2. Ensure that you skip static route syncing in this case using the `disableStaticRouteSync` flag in the `values.yaml` of your helm chart.
 - ***Step 3:*** Kubernetes 1.14+.
 - ***Step 4:*** `helm` cli pointing to your kubernetes cluster.

#### Install using *helm*

*Step 1:* Create the `avi-system` namespace:

    kubectl create ns avi-system

*Step 2:* Configure `helm` cli and point it to your kubernetes cluster

*Step 3:* Clone this repository, go inside the `helm` directory and run:

    helm install ./akc --name my-akc-release --namespace=avi-system --set configs.controllerIP=10.10.10.10

Use the `helm/akc/values.yaml` to edit values related to Avi configuration. Values and their corresponding index can be found [here](#parameters) 


#### Uninstall using *helm*

Simply run:


*Step1:*

    helm delete my-akc-release -n avi-system
 
*Step 2:* 

    kubectl delete ns avi-system

## Parameters


The following table lists the configurable parameters of the AKC chart and their default values.

| **Parameter**                                   | **Description**                                         | **Default**                                                           |
|---------------------------------------------|-----------------------------------------------------|-------------------------------------------------------------------|
| `configs.controllerVersion`                      | Avi Controller version                       | 18.2.7                                                            |
| `configs.controllerIP`                         | Specify Avi controller IP    | `nil`      |
| `configs.shardVSSize`                   | Shard VS size enum values: LARGE, MEDIUM, SMALL     | LARGE      |
| `configs.fullSyncFrequency`                       | Full sync frequency       | 300                                                            |
| `configs.cloudName`                            | Name of the VCenter cloud managed in Avi                              | Default-Cloud                                                       |
| `configs.vrfRefName`                          | VRF context name to be used for the kubernetes cluster                                  | global                                                 |
| `avicredentials.username`                                 | Avi controller username                                  | admin                                                      |
| `avicredentials.password`                          | Avi controller password                          | admin                                                    |
| `image.repository`                         | Specify docker-registry that has the akc image    | 100.64.86.10:5443/avi-k8s-controller      |


## Build and Test

AKC can be built as a docker container using the `make` command. Simply clone the repository
and run:

    make docker
    
Unit tests can be run using:

    make test and make int_test

