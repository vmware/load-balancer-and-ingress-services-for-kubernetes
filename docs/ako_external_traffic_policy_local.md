# ExternalTrafficPolicy Local Support in AKO

This feature is supported starting from AKO version **1.14.1** and enables AKO to support the `externalTrafficPolicy` field set as `Local` in NodePort services. This ensures that only nodes with running application pods are added as pool servers in the Avi Controller, improving traffic efficiency and preserving client IP addresses.

## Overview

The `externalTrafficPolicy` field in Kubernetes services determines whether external traffic should be routed to node-local or cluster-wide endpoints. When set to `Local`, the kube-proxy only distributes load to pods running on the same node, which:

- Preserves the client IP address
- Ensures nodes without pods are not included in service endpoints
- Reduces network hops

## Problem Statement

Prior to this feature, AKO did not respect the `externalTrafficPolicy: Local` setting and would:
- Add all cluster nodes as pool servers regardless of pod distribution
- Rely on health monitors to mark down nodes without running pods
- Result in unnecessary resource overhead from monitoring nodes that should not receive traffic

## Solution

With this feature, AKO now:
1. **Detects Local Traffic Policy**: When processing a NodePort service, AKO checks if `externalTrafficPolicy` is set to `Local`
2. **Selective Pool Server Addition**: Only adds nodes with active running pods as pool servers in the Avi Controller.

## Supported Platforms

This feature works with:
- **Kubernetes clusters** using Ingress resources
- **Kubernetes and OpenShift clusters** using Gateway API (HTTPRoute resources)
- **OpenShift clusters** using Route objects
- **LoadBalancer services in both Kubernetes and OpenShift clusters** in NodePort mode (L4 Virtual Services)

The behavior is consistent across all platforms and resource types when `externalTrafficPolicy: Local` is configured.

## Configuration

No additional configuration is required. AKO automatically detects when a NodePort service has `externalTrafficPolicy: Local` and adjusts the pool server configuration accordingly.

## Example Setup

### 1. Create a NodePort Service with Local Traffic Policy

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ew-app
  namespace: hostrule-ns
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 8080
  selector:
    app: ew-app
  sessionAffinity: None
  type: NodePort
  externalTrafficPolicy: Local
```

### 2. Create a Backend Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ew-app
  namespace: hostrule-ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ew-app
  template:
    metadata:
      labels:
        app: ew-app
    spec:
      containers:
      - image: server-os:latest
        imagePullPolicy: IfNotPresent
        name: deployment
        ports:
        - containerPort: 8080
          protocol: TCP
```

### 3. Create an Ingress Resource

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-secure-ingress-switch
  namespace: hostrule-ns
spec:
  rules:
  - host: test-hostrules-switch-1.avi.internal
    http:
      paths:
      - backend:
          service:
            name: ew-app
            port:
              number: 80
        pathType: ImplementationSpecific
        path: /foo
  tls:
  - hosts:
    - test-hostrules-switch-1.avi.internal
    secretName: ingress-secret
```

**Note**: Ensure the Kubernetes secret object `ingress-secret` (with `tls.crt` and `tls.key` data) is created before ingress creation.

### 4. Create a Gateway API HTTPRoute (Alternative to Ingress or OpenShift Route)

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: test-gateway
  namespace: hostrule-ns
spec:
  gatewayClassName: avi-lb
  listeners:
  - name: https
    protocol: HTTPS
    port: 443
    hostname: "test-hostrules-switch-1.avi.internal"
    tls:
      certificateRefs:
        - kind: Secret
          group: ""
          name: ingress-secret
      mode: Terminate
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: test-http-route
  namespace: hostrule-ns
spec:
  parentRefs:
  - name: test-gateway
    namespace: hostrule-ns
    sectionName: https
  hostnames:
  - "test-hostrules-switch-1.avi.internal"
  rules:
  - matches:
      - path:
         value: "/foo"
    backendRefs:
    - name: ew-app
      port: 80
```

**Note**: Ensure the Gateway API CRDs are installed in your cluster. The `avi-lb` GatewayClass is automatically created by AKO installation.

### 5. Create an OpenShift Route (OpenShift-specific alternative)

```yaml
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: test-secure-route
  namespace: hostrule-ns
spec:
  host: test-hostrules-switch-1.avi.internal
  path: /foo
  to:
    kind: Service
    name: ew-app
    weight: 100
  port:
    targetPort: 80
  tls:
    termination: edge
    certificate: |
      -----BEGIN CERTIFICATE-----
      <certificate-content>
      -----END CERTIFICATE-----
    key: |
      -----BEGIN PRIVATE KEY-----
      <private-key-content>
      -----END PRIVATE KEY-----
```

### 6. Create a LoadBalancer Service with Local Traffic Policy (NodePort Mode)

When using NodePort mode (`serviceType: NodePort`), LoadBalancer services also support `externalTrafficPolicy: Local`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-svc
  namespace: hostrule-ns
spec:
  type: LoadBalancer
  externalTrafficPolicy: Local
  selector:
    name: ew-app
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
```

**Note**: The NodePort service created in step 1 is not required for LoadBalancer services. Kubernetes automatically allocates a NodePort when a LoadBalancer service is created. AKO will create an L4 Virtual Service and populate the pool with only those nodes that have running pods matching the selector `name: ew-app`.

## Behavior Comparison

In a two-node cluster with a single replica deployment:

```bash
kubectl get nodes
NAME                STATUS   ROLES           AGE    VERSION
k8s-cluster-master  Ready    control-plane   120d   v1.31.5
k8s-cluster-worker  Ready    <none>          120d   v1.31.5
```

```bash
kubectl get po -n hostrule-ns -owide
NAME                     READY   STATUS    RESTARTS   AGE   IP               NODE                NOMINATED NODE   READINESS GATES
ew-app-87d56d6fc-tklph   1/1     Running   0          18s   192.168.25.134   k8s-cluster-worker   <none>           <none>
```

### Before the Feature

- Both nodes were added as pool servers
- The node(pool server) without running pods was marked as "down" by health monitors

![Previous Behavior](images/old_ext_traf_pol_local.png)

### After the Feature

- Only the node with running application pods is added as a pool server
- The pool is marked as "up" and ready to serve traffic

![New Behavior](images/new_ext_traf_pol_local.png)


## Benefits

1. **Client IP Preservation**: Traffic is routed directly to nodes with pods, preserving the original client IP address
2. **Improved Performance**: Reduces unnecessary network hops by avoiding nodes without application pods
3. **Better Resource Utilization**: Health monitors are not wasted on checking nodes that shouldn't receive traffic
4. **Automatic Configuration**: No manual intervention required - AKO automatically detects and applies the appropriate configuration

## Related Documentation

- [AKO NodePort Mode](README.md#using-nodeport-mode)
- [AKO NodePort Objects](objects.md#nodeport-mode)
- [AKO NodePort Configuration](values.md#l7settingsservicetype)