# AKO CRD Operator

## Overview

The AKO CRD Operator is a Kubernetes operator that manages Avi Load Balancer objects directly through Custom Resource Definitions (CRDs). Unlike AKO which translates Kubernetes resources (Ingress, Services, Gateway API) into Avi objects, the AKO CRD Operator provides direct lifecycle management of specific Avi Controller objects, enabling fine-grained control over load balancer configurations.

The operator watches for CRD objects in Kubernetes namespaces and synchronizes them with corresponding objects on the Avi Controller, providing declarative management of Avi resources through Kubernetes-native workflows.

## Key Features

- **Direct Avi Object Management**: Create and manage Avi Controller objects directly from Kubernetes
- **Declarative Configuration**: Use Kubernetes CRDs to define Avi resources
- **Status Tracking**: Real-time status updates with Kubernetes Conditions API
- **Multi-Tenancy**: Namespace-scoped resources with tenant isolation

## Supported CRDs

The AKO CRD Operator manages the following Custom Resource Definitions:

### 1. HealthMonitor

Configure health monitoring for backend services with support for:
- TCP Health Monitors
- PING Health Monitors
- HTTP Health Monitors

[HealthMonitor Documentation](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/healthmonitor.md)

### 2. ApplicationProfile

Define application profiles corresponding to Avi.

[ApplicationProfile Documentation](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/applicationprofile.md)

### 3. PKIProfile

Manage PKI profiles for certificate validation:
- Configure trusted Certificate Authorities (Root and Intermediate)
- Enable secure backend communication
- Certificate validation for TLS connections

[PKIProfile Documentation](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/pkiprofile.md)

## Installation and Configuration

### Using Helm

The AKO CRD Operator is bundled with AKO as a dependency and can be installed with AKO. It can be configured via values.yaml provided with AKO.

## Status and Conditions

The AKO CRD Operator updates the status of each CRD object with detailed information:

### Status Fields

- **UUID**: Unique identifier of the object on Avi Controller
- **ObservedGeneration**: Generation of the spec that was last processed
- **LastUpdated**: Timestamp of the last update
- **BackendObjectName**: Name of the object on Avi Controller
- **Tenant**: Avi tenant where the object is created
- **Controller**: Set to "ako-crd-operator"

### Conditions

The operator uses Kubernetes Conditions API to report status:

- **Programmed**: Indicates whether the object has been successfully created/updated on Avi Controller
  - **Reasons**: Created, Updated, CreationFailed, UpdateFailed, DeletionFailed

Example status:

```yaml
status:
  uuid: "healthmonitor-12345-abcde"
  observedGeneration: 1
  lastUpdated: "2025-01-15T10:30:00Z"
  backendObjectName: "my-k8s-cluster--default-http-health-check"
  tenant: "admin"
  controller: "ako-crd-operator"
  conditions:
    - type: Programmed
      status: "True"
      reason: Created
      message: "HealthMonitor successfully created on Avi Controller"
      lastTransitionTime: "2025-01-15T10:30:00Z"
```

## Monitoring and Troubleshooting

### Health Checks

The operator exposes health endpoints:

- **Liveness**: `http://localhost:8081/healthz`
- **Readiness**: `http://localhost:8081/readyz`

### Logs

View operator logs:

```bash
kubectl logs -n avi-system deployment/ako-crd-operator -f
```

### Events

Check Kubernetes events for CRD objects:

```bash
kubectl describe healthmonitor <name> -n <namespace>
kubectl get events -n <namespace> --field-selector involvedObject.name=<name>
```

### Common Issues

1. **Object not created on Avi Controller**
   - Check operator logs for errors
   - Verify Avi Controller credentials
   - Ensure network connectivity to Avi Controller

2. **Status shows CreationFailed**
   - Check CRD spec for validation errors
   - Verify tenant and cloud configuration on Avi Controller
   - Check operator logs for detailed error messages

3. **Object stuck in deletion**
   - Check if object is referenced by other Avi objects
   - Verify operator has permissions to delete objects
   - Check finalizers on the CRD object

## Upgrade

AKO CRD Operator is a dependency of AKO and can be upgraded when upgrading AKO.

## Uninstallation

To uninstall the AKO CRD Operator you need to uninstall AKO:

```bash
# Delete all CRD objects first
kubectl delete healthmonitors --all -A
kubectl delete applicationprofiles --all -A
kubectl delete pkiprofiles --all -A
```
Continue with normal AKO uninstallation.

**Note**: Deleting CRD objects will also delete the corresponding objects from Avi Controller.

## Version Compatibility

| AKO Version (includes AKO CRD Operator) | Avi Controller Version | Kubernetes Version | OpenShift Version |
|-----------------------------------------|------------------------|-------------------|-------------------|
| 2.1.1                                   | 30.1.1+                | 1.29 - 1.34       | 4.16 - 4.18       |

**Note**: AKO CRD Operator is bundled with AKO and shares the same version number.

## Changelog

See [CHANGELOG.md](../ako-crd-operator/CHANGELOG.md) for version history and release notes.

## Additional Resources

- [HealthMonitor CRD Documentation](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/healthmonitor.md)
- [ApplicationProfile CRD Documentation](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/applicationprofile.md)
- [PKIProfile CRD Documentation](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/pkiprofile.md)
- [RouteBackendExtension CRD Documentation](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/routebackendextension.md)
- [CRD Overview](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/overview.md)

