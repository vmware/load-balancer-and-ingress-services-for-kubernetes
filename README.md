## Architecture

The Avi K8s Controller is a layered collection of independent interoperable units
that are used in conjunction to provide L4-L7 load balancing for applications deployed
in a kubernetes cluster for north-south traffic.

The controller ingests the Kubernetes API server object updates namely services and ingress
to construct corresponding objects in Avi controller.

