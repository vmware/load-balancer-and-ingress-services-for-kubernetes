# PKIProfile CRD Documentation

## Table of Contents
- [Introduction](#introduction)
- [PKIProfile Specification](#pkiprofile-specification)
- [Spec Fields](#spec-fields)
- [Status Fields](#status-fields)
- [Usage Examples](#usage-examples)
- [Integration with RouteBackendExtension](#integration-with-routebackendextension)
- [Troubleshooting](#troubleshooting)

## Introduction

The PKIProfile Custom Resource Definition (CRD) manages Certificate Authorities (Root and Intermediate) used for backend certificate validation in Avi Load Balancer. PKIProfile is a crucial component for securing communication between gateways and backend services, ensuring trust validation in modern cloud-native architectures.

**NOTE**: PKIProfile CRD is specifically designed for use with Gateway API and is not supported with traditional Ingress resources.This CRD is handled by the ako-crd-operator.

A sample PKIProfile CRD looks like this:

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: PKIProfile
metadata:
  labels:
    app.kubernetes.io/name: ako-crd-operator
    app.kubernetes.io/managed-by: kustomize
  name: pkiprofile-sample
spec:
  # CaCerts is a list of Certificate Authorities (Root and Intermediate) trusted that is used for certificate validation.
  # Matches AVI SDK PKIProfile.CaCerts structure - only contains certificate data (no private keys)
  ca_certs:
    - certificate: |
        -----BEGIN CERTIFICATE-----
        MIIDBTCCAe2gAwIBAgIUJcVHBy6lHaUBNPInPTpN52HY0ikwDQYJKoZIhvcNAQEL
        BQAwEjEQMA4GA1UEAwwHdGVzdC1jYTAeFw0yNTA5MDEyMDQ0NDVaFw0yNjA5MDEy
        MDQ0NDVaMBIxEDAOBgNVBAMMB3Rlc3QtY2EwggEiMA0GCSqGSIb3DQEBAQUAA4IB
        DwAwggEKAoIBAQCX5F3IojSUkzzK8eu811mTT7WZT6QthMp9ipTMnsHfQBm4+4tb
        t2ZSfaJG7QFsb094uZPEZXEBHzgsxPqQAe3YKD5Mtqb87Z1G5yVczF0kqxihYyMK
        +ds1Gv8LYLueAdsgtI1Ukc78kNLAuOGUYBqfxz0m6/Zwh9mUIwoYhYOqQtLtrwXt
        WV1UlhR6zroBTQyyQNydBzVKQ2wJK+ocWvJX1GpflNsnelsDCo9et7izsZzDwJI9
        Xu2Hy4bERHZdbl15JwmUf9/8anF5fhuPA1gV/nKydc8vT/nZNkI/DrRP91jZenOV
        CpbhEDi/YPVVX++vRMNwup0SJbwWUJFJxfOLAgMBAAGjUzBRMB0GA1UdDgQWBBSu
        vszq4He7ySDfZ7lUol3coE/1NjAfBgNVHSMEGDAWgBSuvszq4He7ySDfZ7lUol3c
        oE/1NjAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBkwnwA2zMG
        bVgTckic2tTNCeO3udVYWX2jFKBhqrpEvKesNqHoDPn9MRmY55z262WVmcrgwoLr
        830V0RxYMNdeyZGjrXfQwx0VQUoVgXTxPhHbH9gv1XzY/VjxrxzTwdSmkScvLuf+
        gmSvvLFekBmwuJ6djGc+tRd84nwHdGWRAEfDOOIGuyG3c2C2HhOdcCj8ccTm3yY3
        69lFg4/PBZ3kF2tyou+QkLK8CxJTQTuyVzs+uaD1pouVzZiDaHy8Dv96sWI3WP7D
        bL/JyB72jbNrCpV5rYIEz2/1/+eNF/UbC4FQIlclBvbPpl6hSrfgD9wMY+axe0cE
        JpDF9pm/jPjP
        -----END CERTIFICATE-----
```

## PKIProfile Specification

The PKIProfile CRD is defined with the following structure:

- **API Version**: `ako.vmware.com/v1alpha1`
- **Kind**: `PKIProfile`
- **Scope**: `Namespaced`
- **Short Name**: `pkip`
- **Plural**: `pkiprofiles`
- **Singular**: `pkiprofile`

## Spec Fields

The PKIProfile CRD supports the following configuration options:

### ca_certs (required)

**Description**: A list of Certificate Authorities (Root and Intermediate) trusted that is used for certificate validation. This field matches the AVI SDK PKIProfile.CaCerts structure and only contains certificate data (no private keys).The certificate field is required, must contain at least one PEM-encoded certificate, and each certificate must be at least 1 character long.

**Structure**:
```yaml
ca_certs:
  - certificate: |
      -----BEGIN CERTIFICATE-----
      # Your CA certificate content here
      -----END CERTIFICATE-----
  - certificate: |
      -----BEGIN CERTIFICATE-----
      # Additional CA certificate content here
      -----END CERTIFICATE-----
```

**Certificate Requirements**:
- Must be PEM-encoded
- Must contain valid X.509 certificate data
- Can include Root CA certificates
- Can include Intermediate CA certificates
- Must not contain private keys

## Status Fields

The PKIProfile CRD provides status information through the following fields:

### conditions (optional)

**Type**: `[]metav1.Condition`

**Description**: Represents the latest available observations of the PKIProfile's current state.

**Supported Condition Types**:
- **"Programmed"**: Indicates whether the PKIProfile has been successfully processed and programmed on the Avi Controller
  - **True Reasons**: "Created", "Updated"
  - **False Reasons**: "CreationFailed", "UpdateFailed", "UUIDExtractionFailed", "DeletionFailed", "DeletionSkipped"

### uuid (optional)

**Type**: `string`

**Description**: Unique identifier of the PKI profile object in the Avi Controller.

### observedGeneration (optional)

**Type**: `int64`

**Description**: The observed generation by the operator, used to track when the spec has been processed.

### lastUpdated (optional)

**Type**: `*metav1.Time`

**Description**: Timestamp when the object was last updated.

### backendObjectName (optional)

**Type**: `string`

**Description**: The name of the backend object in the Avi Controller.

### tenant (optional)

**Type**: `string`

**Description**: The tenant where the PKI profile is created.

### controller (optional)

**Type**: `string`

**Description**: Field populated by AKO CRD operator as "ako-crd-operator".

## Usage Examples

### Basic PKIProfile Configuration

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: PKIProfile
metadata:
  name: basic-ca-profile
  namespace: production
spec:
  ca_certs:
    - certificate: |
        -----BEGIN CERTIFICATE-----
        # Your Root CA certificate here
        -----END CERTIFICATE-----
```

### Multiple CA Certificates

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: PKIProfile
metadata:
  name: multi-ca-profile
  namespace: production
spec:
  ca_certs:
    - certificate: |
        -----BEGIN CERTIFICATE-----
        # Root CA certificate
        -----END CERTIFICATE-----
    - certificate: |
        -----BEGIN CERTIFICATE-----
        # Intermediate CA certificate
        -----END CERTIFICATE-----
    - certificate: |
        -----BEGIN CERTIFICATE-----
        # Additional CA certificate
        -----END CERTIFICATE-----
```

### Production PKIProfile with Labels

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: PKIProfile
metadata:
  name: production-ca-profile
  namespace: production
  labels:
    app.kubernetes.io/name: ako-crd-operator
    app.kubernetes.io/managed-by: kustomize
    environment: production
    team: platform
spec:
  ca_certs:
    - certificate: |
        -----BEGIN CERTIFICATE-----
        # Production Root CA certificate
        -----END CERTIFICATE-----
    - certificate: |
        -----BEGIN CERTIFICATE-----
        # Production Intermediate CA certificate
        -----END CERTIFICATE-----
```

## Integration with RouteBackendExtension

PKIProfile is designed to work with RouteBackendExtension for secure backend communication. The PKIProfile is referenced in the `backendTLS.pkiProfile` field of RouteBackendExtension.

For detailed integration examples and BackendTLS configuration, refer to the [RouteBackendExtension documentation](routebackendextension.md).


## Troubleshooting

### Common Issues

**1. Invalid Certificate Format**
*Symptom*: PKIProfile shows validation error about certificate format

*Solution*: 
- Ensure certificates are PEM-encoded
- Verify certificate content is valid X.509 format
- Check for proper certificate headers and footers

```bash
# Validate certificate format
openssl x509 -in certificate.pem -text -noout
```

**2. Certificate Validation Failures**
*Symptom*: Backend TLS connections fail with certificate errors

*Solution*:
- Verify CA certificates match the backend service certificates
- Check certificate chain completeness
- Ensure certificates are not expired

**3. AKO CRD Operator Not Processing PKIProfile**
*Symptom*: PKIProfile status shows "Programmed: False"

*Solution*:
- Check AKO CRD Operator pod status
- Review controller logs for errors
- Verify RBAC permissions

```bash
kubectl get pods -n avi-system
kubectl logs -f deployment/ako-crd-operator -n avi-system
```

### Debug Commands

```bash
# Check PKIProfile status
kubectl get pkiprofiles -A
kubectl describe pkiprofile <name> -n <namespace>

# Check PKIProfile events
kubectl get events --field-selector involvedObject.name=<pkiprofile-name> -n <namespace>

# Check AKO CRD Operator logs
kubectl logs deployment/ako-crd-operator -n avi-system -f

# Verify Avi Controller configuration
# (Access Avi Controller UI to verify PKI Profile configuration)
```

### Status Interpretation

**Programmed: True**
- PKIProfile has been successfully created on the Avi Controller
- Ready to be referenced by RouteBackendExtension

**Programmed: False with reason "CreationFailed"**
- Failed to create PKIProfile on Avi Controller
- Check controller logs for detailed error information

**Programmed: False with reason "UUIDExtractionFailed"**
- PKIProfile was created but UUID extraction failed
- Check Avi Controller connectivity and permissions