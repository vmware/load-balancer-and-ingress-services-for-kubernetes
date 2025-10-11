# AKO Gateway API with Istio mTLS Support

This document describes how AKO Gateway API can be deployed in an Istio environment with strict mTLS support. This extends the existing Istio integration from AKO to support Gateway API (HTTPRoute resources) in both ClusterIP and NodePort modes.

## Overview

AKO Gateway API leverages the same Istio infrastructure as AKO, enabling strict mTLS communication between the Avi Service Engines and backend pods running in an Istio service mesh. This ensures secure, encrypted communication using certificates managed by Istio.

### Supported Modes

- **ClusterIP mode**: Fully supported
- **NodePort mode**: Fully supported

## Prerequisites

- Istio must be installed and configured in your Kubernetes cluster
- AKO must be deployed with `istioEnabled: true` in values.yaml
- Gateway API CRDs must be installed

## Implementation Details

### How It Works

When `istioEnabled` is set to `true` in values.yaml:

1. **Istio Sidecar Injection**: Envoy proxy sidecar annotations are added to the AKO StatefulSet
2. **Certificate Management**: 
   - AKO container mounts the `istio-certs` volume at `/etc/istio-output-certs/`
   - This path contains three files managed by Istio:
     - `cert-chain.pem` - Certificate chain for the workload
     - `key.pem` - Private key for the workload
     - `root-cert.pem` - Root CA certificate
3. **AVI Object Creation**:
   - AKO reads these files and creates a Kubernetes secret `istio-secret` in the AKO namespace
   - AKO creates an AVI PKIProfile: `istio-pki-<clustername>-<AKOnamespace>`
   - AKO creates an AVI SSLKeyCert: `istio-workload-<clustername>-<AKOnamespace>`
4. **Certificate Rotation**: 
   - AKO watches for certificate updates from Istio
   - Automatically updates the `istio-secret`, PKIProfile, and SSLKeyCert when certificates are rotated

### Gateway API Integration

The `ako-gateway-api` container uses the same infrastructure:

- Reads the `ISTIO_ENABLED` environment variable
- Reuses the PKIProfile and SSLKeyCert created by the `ako` container
- Automatically adds these to backend pools for Gateway API HTTPRoute resources

This approach ensures:
- Consistent certificate management across AKO and Gateway API
- Reduced maintenance overhead
- Seamless mTLS support for both ClusterIP and NodePort modes

## Configuration

### Step 1: Enable Istio in AKO

Set the `istioEnabled` flag to `true` in your AKO `values.yaml`:

```yaml
istioEnabled: true
```

This configuration:
- Enables Istio sidecar injection for AKO pods
- Sets the `ISTIO_ENABLED` environment variable for both `ako` and `ako-gateway-api` containers
- Mounts the `istio-certs` volume

### Step 2: Deploy AKO

Install or upgrade AKO with the updated configuration:

```bash
helm install ako/ako --generate-name --version 1.14.1 -f values.yaml -n avi-system
```

### Step 3: Verify Istio Integration

#### Verify Sidecar Injection

Check that the Istio proxy sidecar is running:

```bash
kubectl logs ako-0 -n avi-system -c istio-proxy
```

You should see Envoy proxy logs indicating successful startup.

#### Verify Istio Secret

Verify that the `istio-secret` is created with certificate data:

```bash
kubectl describe secret istio-secret -n avi-system
```

Expected output should show three data fields:
- `cert-chain` - Contains the certificate chain
- `key` - Contains the private key  
- `root-cert` - Contains the root CA certificate

#### Verify AVI Objects

Log in to your Avi Controller UI and verify:

1. **PKIProfile**: Navigate to Templates → Security → PKI Profile
   - Look for `istio-pki-<clustername>-avi-system`
   - Verify it contains the root CA certificate

2. **SSLKeyCert**: Navigate to Templates → Security → SSL/TLS Certificates
   - Look for `istio-workload-<clustername>-avi-system`
   - Verify it contains the workload certificate and key

### Step 4: Create Gateway API Resources

Create a Gateway with an HTTPRoute that references a backend service:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: istio-gateway
  namespace: default
spec:
  gatewayClassName: avi-lb
  listeners:
  - name: https
    protocol: HTTPS
    port: 443
    hostname: "app.example.com"
    tls:
      certificateRefs:
        - kind: Secret
          group: ""
          name: gateway-cert
      mode: Terminate
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: app-route
  namespace: default
spec:
  parentRefs:
  - name: istio-gateway
    sectionName: https
  hostnames:
  - "app.example.com"
  rules:
  - matches:
      - path:
         value: "/app"
    backendRefs:
    - name: app-service
      port: 80
```

### Step 5: Verify mTLS in Backend Pools

In the Avi Controller UI:

1. Navigate to Applications → Pools
2. Select the pool corresponding to your HTTPRoute backend
3. Verify the following are configured:
   - **PKI Profile**: `istio-pki-<clustername>-avi-system`
   - **SSL Key and Certificate**: `istio-workload-<clustername>-avi-system`

A sample pool configuration with Istio mTLS can be found in [this image](istio_pool.png).

When these settings are present, the Avi Service Engine automatically establishes mTLS connections with the backend pods.

## Service Name for AKO Gateway API

AKO Gateway API and the Avi Service Engines use a service name based on the AKO service account and namespace:

**Format**: `cluster.local/ns/<AKOnamespace>/sa/<AKOServiceAccount>`

**Example**: `cluster.local/ns/avi-system/sa/ako-sa`

This service name should be used when configuring Istio authorization policies or peer authentication policies.

## Testing mTLS Connectivity

### Successful Traffic Test

With mTLS properly configured:

1. Send a request to your application via the Avi Controller:
   ```bash
   curl -k https://app.example.com/app
   ```

2. You should receive a successful response from your backend application

3. In the Avi Controller, verify pool members show "UP" status

### Certificate Validation Test

To verify mTLS is actually being enforced:

1. Temporarily remove the PKIProfile and SSLKeyCert from the pool in Avi Controller
2. Send a request - it should fail with SSL/TLS errors
3. Re-add the PKIProfile and SSLKeyCert
4. Traffic should succeed again

This confirms that mTLS is actively securing the connection between Service Engines and backend pods.

## Behavior in Different Modes

### ClusterIP Mode

- Service Engines connect directly to pod IPs
- mTLS is established between SE and pod Envoy sidecars
- Pod IPs are used as pool members

### NodePort Mode

- Service Engines connect to node IPs with NodePort
- mTLS is established through the node to pod Envoy sidecars
- Node IPs with NodePort are used as pool members
- Works with both `externalTrafficPolicy: Cluster` and `externalTrafficPolicy: Local`

## Limitations and Considerations

### Priority of TLS Settings

When both Istio and custom BackendTLSPolicy (via RouteBackendExtension) are configured:
- Custom TLS settings from RouteBackendExtension take precedence over Istio settings
- If you want Istio mTLS to be used, do not specify custom PKI profiles in RouteBackendExtension

### Certificate Rotation

- Istio automatically rotates certificates based on its configuration (typically every 24 hours by default)
- AKO automatically detects and applies the new certificates
- No manual intervention required during certificate rotation

### Resource Sharing

- The same `istio-secret`, PKIProfile, and SSLKeyCert are shared between:
  - AKO (for Ingress and Route objects)
  - AKO Gateway API (for Gateway and HTTPRoute objects)
- This ensures consistent mTLS behavior across all resource types

## Troubleshooting

### Sidecar Injection Not Working

**Problem**: Istio sidecar is not injected into AKO pods

**Solution**: Enable Istio injection for the AKO namespace:

```bash
kubectl label namespace avi-system istio-injection=enabled --overwrite
```

Then restart the AKO pods:

```bash
kubectl rollout restart statefulset ako -n avi-system
```

### istio-secret Not Created

**Problem**: The `istio-secret` is not present in the AKO namespace

**Solutions**:

1. Verify AKO ClusterRole has permissions to create/update secrets:
   ```bash
   kubectl get clusterrole ako-cr -o yaml | grep -A5 secrets
   ```

2. Check AKO logs for errors:
   ```bash
   kubectl logs ako-0 -n avi-system -c ako
   ```

3. Verify the istio-certs volume is mounted:
   ```bash
   kubectl exec ako-0 -n avi-system -c ako -- ls -la /etc/istio-output-certs/
   ```

### PKIProfile or SSLKeyCert Not Created in Avi

**Problem**: Istio objects are not visible in Avi Controller

**Solutions**:

1. Verify `istio-secret` has all three data fields populated:
   ```bash
   kubectl get secret istio-secret -n avi-system -o yaml
   ```

2. Check AKO logs for Avi API errors:
   ```bash
   kubectl logs ako-0 -n avi-system -c ako | grep -i "istio\|pki\|ssl"
   ```

3. Verify connectivity between AKO and Avi Controller

### Pool Members Down Despite mTLS Configuration

**Problem**: Backend pool members show "DOWN" status even with correct mTLS settings

**Solutions**:

1. Verify backend pods have Istio sidecars injected:
   ```bash
   kubectl get pods -n <namespace> -o jsonpath='{.items[*].spec.containers[*].name}'
   ```
   You should see `istio-proxy` listed.

2. Check Istio PeerAuthentication policy is not blocking traffic:
   ```bash
   kubectl get peerauthentication -A
   ```

3. Check Service Engine logs in Avi Controller for SSL handshake errors

### Certificate Rotation Issues

**Problem**: Certificates are rotated by Istio but not updated in Avi

**Solutions**:

1. Verify AKO is watching the certificate files:
   ```bash
   kubectl logs ako-0 -n avi-system -c ako | grep "certificate\|secret update"
   ```

2. Check if the `istio-secret` is being updated:
   ```bash
   kubectl describe secret istio-secret -n avi-system
   ```
   Look at the `Age` or `Last Updated` timestamp.

3. Restart AKO to force certificate refresh:
   ```bash
   kubectl delete pod ako-0 -n avi-system
   ```

## Related Documentation

- [AKO on Istio](../istio_support.md) - Original Istio support for AKO with Ingress/Routes
