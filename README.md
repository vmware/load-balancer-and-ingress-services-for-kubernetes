# Avi Kubernetes Controller

### Architecture

The Avi K8s Controller (AKC) is a layered collection of independent interoperable units
that are used in conjunction to provide L4-L7 load balancing for applications deployed
in a kubernetes cluster for north-south traffic.

The controller ingests the Kubernetes API server object updates namely services and ingress
to construct corresponding objects in Avi controller. Hence the AKC performs the dual
functionality of an ingress controller and a layer 4 load balancer.

![Alt text](AKC.jpg?raw=true "Title")

### Run AKC

AKC runs as a POD inside the kubernetes cluster.

#### Run using helm

To Run AKC you need the following pre-requisites:
 - ***Step 1***: An Avi Controller with a vCenter cloud.

 - ***Step 2***: If your POD CIDRs are not routable:
    - Create a VRF context object in Avi for the kubernetes controller.
    - Get the name the PG network which the kubernetes nodes are part of. 
    
      *NOTE: If you are using AKC for test puposes you can use the `global` vrf but you cannot manage multiple kubernetes clusters in the same cloud with this setting.*
    - Configure this PG network with the vrf mentioned in the previous step.
    - Make sure this PG network is part of the NS IPAM configured in the vCenter cloud.

 - ***Step 2.1***: If your POD CIDRs are routable then you can skip step 2.

##### Using Helm

If you have helm configured in your kubernetes cluster then simply clone this repository, go inside the `helm` directory and run (to be updated):

    helm install ./akc

Use the `values.yaml` to edit values related to Avi configuration. Values and their corresponding index can be found here.

##### Using `kubectl`
If you do not have helm and would want to deploy kubernetes in the crude way, pls execute the following commands:
  - Obtain the avi network credentials. Encode them in base64 as follows:
        
        echo -n "admin" | base64 -w 0

  - In the `secret.yaml` update the `username` and `password`.
  - In the `configmap.yaml` update the values in accordance to their fields.
  - Execute the following:
        
        kubectl create -f secret.yaml
        kubectl create -f configmap.yaml
        kubectl create -f deployment.yaml

Ensure to edit the deployment.yaml with the right image tag if you are using a private docker repository.


### Build and Test

AKC can be built as a docker container using the `make` command. Simply clone the repository
and run:

    make docker
    
Unit tests can be run using:

    make test and make int_test

