# SCTP support in AKO for L4 services

This feature supports SCTP protocol for L4 services in AKO. Starting with 1.9.1, AKO will support SCTP protocol with Kubernetes/Openshift LoadBalancer services, and Gateway objects and their corresponding backend services. For more information on using Gateway class and Gateway objects with AKO, please refer to this document, https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/gateway-api/gateway-api.md.

## Overview

With version 22.1.3, AVI Controller supports SCTP traffic for L4 virtual services. Starting with AKO 1.9.1, AKO will support SCTP protocol in port definitions of LoadBalancer services, and Gateway objects and their corresponding backend services. Prior to 1.9.1, only TCP and UDP protocols were supported.

The AVI Controller has introduced SCTP specific properties for virtual services and pools. These include `System-SCTP-Proxy` TCP/UDP (network) profile for supporitng SCTP traffic in virtual services, an SCTP based `System-SCTP` health monitor for pools, and `SCTP` protocol match option, in L4 Policy Set match rules.

The user needs to create a LoadBalancer service, or a Gateway based L4 service, with SCTP protocol in port definition. The AKO running in the Kubernetes/Openshift cluster, will consume the service and gateway definitions. AKO will create the corresponding virtual service in AVI Controller, with appropriate `System-SCTP-Proxy` TCP/UDP (network) profile, and the corresponding pools, with appropriate `System-SCTP` health monitor. The L4PolicySet is also created with appropriate match rules for `SCTP` protocol.

> **Note**: SCTP protocol support is not available for service type NodePortLocal, as Antrea CNI does not support SCTP Service ports, for NodePortLocal type services.

## Configuration

As already stated, the configuration mainly includes creating LoadBalancer services, and Gateway objects and their corresponding backend services, with SCTP protocol in port definitions. Some sample yaml definitions are shown below.

### LoadBalancer Service

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sctp-demo
  labels:
    app: server
spec:
  replicas: 2
  selector:
    matchLabels:
      app: server
  template:
    metadata:
      labels:
        app: server
    spec:
      containers:
      - name: sctp-demo
        image: <sctp enabled container image>
        ports:
        - containerPort: 9090
          protocol: SCTP
---
apiVersion: v1
kind: Service
metadata:
  name: server
  namespace: default
spec:
  ports:
  - port: 80
    protocol: SCTP
    targetPort: 9090
  selector:
    app: server
  type: LoadBalancer
```

### Gateway Object

```
apiVersion: networking.x-k8s.io/v1alpha1
kind: GatewayClass
metadata:
  name: avi-lb
spec:
  controller: ako.vmware.com/avi-lb
  parametersRef:
    group: ako.vmware.com
    kind: AviInfraSetting
    name: my-infrasetting
---
apiVersion: ako.vmware.com/v1alpha1
kind: AviInfraSetting
metadata:
  name: my-infrasetting
---
apiVersion: networking.x-k8s.io/v1alpha1
kind: Gateway
metadata:
  name: my-gateway
  namespace: svcapi
spec:
  gatewayClassName: avi-lb
  listeners:
  - port: 6060
    protocol: SCTP
    routes:
      group: v1
      kind: services
      selector:
        matchLabels:
          ako.vmware.com/gateway-name: my-gateway
          ako.vmware.com/gateway-namespace: svcapi
---
apiVersion: v1
kind: Service
metadata:
  labels:
    ako.vmware.com/gateway-name: my-gateway
    ako.vmware.com/gateway-namespace: svcapi
  name: avisvc-svcapi
  namespace: svcapi
spec:
  ports:
  - name: sixtysixty
    port: 6060
    protocol: SCTP
    targetPort: 9090
  selector:
    app: avi-server
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: avi-server
  namespace: svcapi
spec:
  replicas: 1
  selector:
    matchLabels:
      app: avi-server
  template:
    metadata:
      labels:
        app: avi-server
    spec:
      containers:
      - image: <sctp enabled container image>
        imagePullPolicy: IfNotPresent
        name: avi-server
        ports:
        - containerPort: 9090
          protocol: SCTP
```
> **Note**: The above example for Gateway defines service of type ClusterIP as the backend service. But, service of type NodePort can also be used.
