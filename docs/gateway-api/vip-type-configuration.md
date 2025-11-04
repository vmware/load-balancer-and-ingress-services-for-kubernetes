# Gateway VIP Type Configuration (NSX-T VPC Mode)

## Overview

When using Gateway API with AKO in NSX-T cloud VPC mode, you can control whether Virtual IPs (VIPs) are allocated from the **Public** or **Private** IP address pool within your VPC. This configuration allows you to specify the type of VIP allocation for each Gateway resource, providing flexibility in network design and security requirements.

By default, AKO allocates VIPs from the **Public** IP address pool. You can override this behavior using a Gateway annotation.

## When to Use This Feature

This feature is applicable only when:
- **Cloud Type**: NSX-T Cloud
- **Mode**: VPC Mode (enabled via `akoSettings.vpcMode: true`)
- **Resource**: Gateway API Gateway objects

In VPC mode, AKO automatically configures VIP networks and allocates IP addresses from IP address pools within the specified VPC, eliminating the need for manual VIP network configuration.

## Configuration

### Annotation

Add the following annotation to your Gateway object to specify the VIP type:

```yaml
metadata:
  annotations:
    networking.vmware.com/lb-vip-type: "<type>"
```

### Supported Values

| Annotation Value | Description | Use Case |
|-----------------|-------------|----------|
| `public` | Allocates VIP from the Public IP address pool | For services that need external accessibility or internet-facing workloads |
| `private` | Allocates VIP from the Private IP address pool | For internal services, improved security, or private network requirements |

**Note**: The annotation values are case-insensitive. AKO accepts lowercase values (`public`, `private`) and converts them to uppercase internally (`PUBLIC`, `PRIVATE`).

## Examples

### Example 1: Gateway with Public VIP (Default Behavior)

If you do not specify the annotation, AKO defaults to allocating a Public VIP:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: public-gateway
  namespace: production
spec:
  gatewayClassName: avi-lb
  listeners:
  - name: http
    protocol: HTTP
    port: 80
```

### Example 2: Gateway with Explicit Public VIP

To explicitly configure a Public VIP (though this is the default):

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: public-gateway
  namespace: production
  annotations:
    networking.vmware.com/lb-vip-type: "public"
spec:
  gatewayClassName: avi-lb
  listeners:
  - name: http
    protocol: HTTP
    port: 80
  - name: https
    protocol: HTTPS
    port: 443
    tls:
      certificateRefs:
      - kind: Secret
        name: tls-secret
```

### Example 3: Gateway with Private VIP

To allocate a VIP from the Private IP address pool:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: internal-gateway
  namespace: internal-services
  annotations:
    networking.vmware.com/lb-vip-type: "private"
spec:
  gatewayClassName: avi-lb
  listeners:
  - name: http
    protocol: HTTP
    port: 80
```

### Example 4: Mixed Configuration

You can configure different VIP types for different Gateways in the same cluster:

```yaml
# Public-facing Gateway
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: external-api-gateway
  namespace: api
  annotations:
    networking.vmware.com/lb-vip-type: "public"
spec:
  gatewayClassName: avi-lb
  listeners:
  - name: https
    protocol: HTTPS
    port: 443
    tls:
      certificateRefs:
      - kind: Secret
        name: api-cert
---
# Internal Gateway
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: internal-admin-gateway
  namespace: admin
  annotations:
    networking.vmware.com/lb-vip-type: "private"
spec:
  gatewayClassName: avi-lb
  listeners:
  - name: https
    protocol: HTTPS
    port: 443
    tls:
      certificateRefs:
      - kind: Secret
        name: admin-cert
```

## Important Notes

### Limitations

1. **NSX-T VPC Mode Only**: This annotation only takes effect when AKO is running in NSX-T cloud with VPC mode enabled. In non-VPC mode or other cloud types, this annotation is ignored.

2. **Invalid Annotation Values**: If you provide an annotation value that is not supported (e.g., `internal` or `mixed`), AKO will ignore it and fall back to the default behavior (Public VIP allocation).

3. **Annotation Updates**: You should not update the annotation value on an existing Gateway if VIP, associated with that Gateway, is present on the AviController.

### VPC Mode Prerequisites

Before using this feature, ensure VPC mode is properly configured:

1. Enable VPC mode in `values.yaml`:
   ```yaml
   akoSettings:
     vpcMode: true
   ```

2. Configure the nsxtT1LR field with VPC path:
   ```yaml
   NetworkSettings:
     nsxtT1LR: "/orgs/<org-id>/projects/<project-id>/vpcs/<vpc-id>"
   ```

3. Disable VIP network list (required for VPC mode):
   ```yaml
   NetworkSettings:
     vipNetworkList: []
   ```

For more information about VPC mode configuration, refer to the [VPC Mode documentation](../vpc_mode.md).

## How It Works

1. When AKO processes a Gateway object, it checks for the `networking.vmware.com/lb-vip-type` annotation.

2. If the annotation is present and contains a supported value (`public` or `private`), AKO uses that value to configure the VIP type in the underlying Avi Virtual Service VIP (VSVIP) object.

3. If the annotation is missing or contains an unsupported value, AKO defaults to `PUBLIC` VIP allocation.

4. In NSX-T VPC mode, AKO constructs the VIP network identifier using the project, VPC, and VIP type information, which determines from which IP address pool the VIP is allocated.

## Troubleshooting

### Annotation Not Taking Effect

If your VIP type annotation appears to be ignored:

1. **Verify VPC Mode**: Confirm that `vpcMode` is set to `true` in your AKO configuration.
   ```bash
   kubectl get configmap avi-k8s-config -n <ako-namespace> -o yaml | grep vpcMode
   ```

2. **Check Cloud Type**: Ensure you are using NSX-T cloud by verifying cloud, defined by `cloudName`, on AviController:
   ```bash
   kubectl get configmap avi-k8s-config -n <ako-namespace> -o yaml | grep cloudName
   ```

3. **Verify Annotation Syntax**: Check that the annotation key and value are correct:
   ```bash
   kubectl get gateway <gateway-name> -n <namespace> -o yaml | grep lb-vip-type
   ```

4. **Check AKO Logs**: Review AKO logs for any warnings or errors related to VIP configuration:
   ```bash
   kubectl logs -n <ako-namespace> <ako-pod-name> | grep -i "vip\|vs vip"
   ```

### Common Issues

- **Wrong Annotation Value**: Using values other than `public` or `private` (case-insensitive) will be silently ignored. Always use `public` or `private`.

- **VPC Mode Not Enabled**: The annotation only works in VPC mode. In standard mode, VIP networks are configured via `vipNetworkList` and this annotation has no effect.

## Related Documentation

- [AKO VPC Mode Configuration](../vpc_mode.md)
- [Gateway API Documentation](gateway-api-v1.md)
- [AviInfraSetting CRD](../crds/avinfrasetting.md)

