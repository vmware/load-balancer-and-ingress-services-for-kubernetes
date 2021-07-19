## Setting up routing rules using CRDs

This document outlines the use of AKO specific CRD objects that allows the users to express Avi related properties.

### What are Custom Resource Definitions (CRDs)? 

Custom Resource Definitions or CRDs are used to extend the Kubernetes APIs server with additional schemas.
More about CRDs can be read [here](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)

AKO ships a bunch of CRD objects (installed through helm). The CRDs are envisioned for two types of audiences:

* __Operators__: Users of this category are aware of Avi related semantics, have access to the Avi controller. They manage the lifecycle
  of AKO.
    
* __Developers__: They are owners of microservices deployed in Kubernetes. They are assumed to know basic routing principles but don't
  know specifics of Avi atributes. 
  

### Why are CRDs better?

Some loadbalancers allow configuration options via annotations. The following reasons were considered to choose CRDs:

* __Versioning__: CRDs, allow AKO to version fields appropriately due it's the dependency on the Avi Controller Versions. In general
this allows users to preserve unique states across various deployment versions.

* __Syntactical Validations__: CRDs can be used to verify syntax at the time of creation of the CR object. This saves a lot of API cost
and allows quicker feedback to the user using a combination of field constraints and effective `status` messages.

* __Role segregation__: CRDs can benefit from the RBAC policies of Kubernetes and allow stricter access to a group of users.

### CRD Types in AKO

AKO categorizes the CRDs in the following buckets:

1. __Layer 7__: These CRD objects are used to express layer 7 traffic routing rules. Following are the list of CRDs currently available:
  
    * [HostRule](https://github.com/avinetworks/avi-helm-charts/tree/master/docs/AKO/crds/hostrule.md)
    * [HTTPRule](https://github.com/avinetworks/avi-helm-charts/tree/master/docs/AKO/crds/httprule.md)
  
2. __Layer 4__: These CRD objects are used to express layer 4 trafffic routing rules. (Unreleased)

3. __Infrastructure__: These CRD objects are used to control Avi's infrastructure components like Ingress Class, SE group properties etc. (Unreleased)
