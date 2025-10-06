# AKO on Istio (Tech Preview)

This feature allows AKO to be deployed on an Istio environment. Strict mTLS is supported in both ClusterIP and NodePort modes.

## Steps to deploy and verfify AKO deployment

Set `istioEnabled` flag to `true` in values.yaml. This should be enough to allow AKO to work on an Istio environment.

Verify istio sidecar injection is enabled and working.
`kubectl logs ako-0 -n avi-system -c istio-proxy`

Verify `istio-secret` secret is created in ako namespace with `cert-chain`, `key` and `root-cert` data populated. These correspond to the workload and CA certificates.
`kubectl describe secret istio-secret -n <AKOnamesapce>`

Verify pkiprofile `istio-pki-<clustername>-<AKOnamespace>` and sslkeyandcertification `istio-workload-<clustername>-<AKOnamespace>` are created on controller.

## Service Name for AKO

AKO and the AVI service engines use a service name based on the AKO service account and AKO namespace as such `cluster.local/ns/<AKOnamespace>/sa/<AKOServiceAccount>`.

Eg. `cluster.local/ns/avi-system/sa/ako-sa`

This service name should be used when updating the auth policy crd for istio.

## Unsupported features

AKO prioritizes istio pkiprofile over any other pkiprofile reference added using httprule.

**Note** AKO works only with L7.

## Workarounds and Fixes 

### Sidecar injection for AKO is not working

Try enabling injection for the ako namespace eg. `kubectl label namespace avi-system istio-injection=enabled --overwrite`

### `istio-secret` is not created

Check AKO clusterrole has permissions to create/update secrets in ako namespace.