# AKO CRD Operator - Object Naming Conventions

## Overview

AKO CRD Operator creates and manages Avi Controller objects based on Kubernetes Custom Resource Definitions (CRDs). This document describes the naming conventions used for objects created on the Avi Controller.

## General Naming Pattern

All objects created by AKO CRD Operator follow this general pattern:

```
ako-crd-operator-<cluster-name>--<sha1-hash-of-namespace-objectName>
```

### Components:

1. **Prefix**: `ako-crd-operator-` - Identifies objects created by the CRD operator
2. **Cluster Name**: Derived from the `clusterID` in the `avi-k8s-config` ConfigMap
3. **Separator**: `--` (double dash)
4. **Encoded Value**: SHA1 hash of `<namespace>-<objectName>`

## Object-Specific Naming Conventions

### 1. HealthMonitor

**Format**: `ako-crd-operator-<cluster-name>--<sha1-hash>`

**Hash Input**: `<namespace>-<healthmonitor-name>`

**Example**:
```
CRD: HealthMonitor "my-health-check" in namespace "app-ns" with cluster name "cluster1"
Avi Object: ako-crd-operator-cluster1--0f10bce27de8131f458c31b46ea91e27358b3b22
```

**Avi Object Type**: HealthMonitor

---

### 2. ApplicationProfile

**Format**: `ako-crd-operator-<cluster-name>--<sha1-hash>`

**Hash Input**: `<namespace>-<applicationprofile-name>`

**Example**:
```
CRD: ApplicationProfile "custom-http-profile" in namespace "app-ns" with cluster name "cluster1"
Avi Object: ako-crd-operator-cluster1--25473353622c80679af5cfcc3cdee7c7e0007303
```

**Avi Object Type**: ApplicationProfile

---

### 3. PKIProfile

**Format**: `ako-crd-operator-<cluster-name>--<sha1-hash>`

**Hash Input**: `<namespace>-<pkiprofile-name>`

**Example**:
```
CRD: PKIProfile "my-ca-bundle" in namespace "app-ns" with cluster name "cluster1"
Avi Object: ako-crd-operator-cluster1--ec7217373c9ae28cf92d3452997e1a3b252053e9
```

**Avi Object Type**: PKIProfile

---

## Metadata and Markers

All objects created by AKO CRD Operator include:

### 1. Created By Field
```
created_by: ako-crd-operator-<cluster-name>
```

### 2. Markers
Markers are used for object identification and filtering:
```yaml
markers:
  - key: "clustername"
    values: ["<cluster-name>"]
  - key: "namespace"
    values: ["<namespace>"]
```

### 3. Tenant
Objects are created in the tenant associated with the namespace:
- Default: `admin` tenant
- Can be customized per namespace using namespace annotations

---

## Object Name Length Limits

- **Maximum Length**: 255 characters (Avi Controller limit)
