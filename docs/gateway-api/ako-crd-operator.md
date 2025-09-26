# AKO CRD Operator Documentation

## Table of Contents
- [Introduction](#introduction)
- [BackendTLS Configuration](#backendtls-configuration)
- [PKIProfile CRD](#pkiprofile-crd)
- [RouteBackendExtension CRD](#routebackendextension-crd)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)

## Introduction

The AKO CRD Operator extends the Avi Kubernetes Operator (AKO) to support Custom Resource Definitions (CRDs) for advanced load balancing configurations. Following Custom Resource Definitions are introduced to support various features in AVI LB:
- PKIProfile
- ApplicationProfile
- HealthMonitor
- RouteBackendExtension

## BackendTLS Configuration

BackendTLS policies are crucial for securing communication between gateways and backend services, ensuring end-to-end encryption and trust validation in modern cloud-native architectures.

### Avi ALB Backend TLS Fields

In Avi ALB, backend TLS can be configured by setting the following fields on the Pool attached to a Virtual Service:

| Field | Description | Required |
|-------|-------------|----------|
| **PKIProfile** | Validates SSL certificates against selected PKI Profile | Optional |
| **SSLProfile** | Defines ciphers and SSL versions for re-encryption | Optional |
| **SslKeyAndCertificate** | Client SSL certificate for server validation | Optional |
| **HostCheckEnabled** | Enable common name check for server certificate | Optional |
| **DomainName** | Domain names for certificate verification. This domain name will be validated against the domain name in the certificate provided by server | Optional |
| **SniEnabled** | Enable TLS SNI for server connections | Optional |
| **ServerName** | FQDN for TLS SNI extension | Optional |
| **RewriteHostHeaderToServerName** | Rewrite Host Header to server name | Optional |

### Current Implementation

For the this implementation, the following three fields are supported:

1. **PKIProfileRef**: Reference to a PKIProfile CRD object
2. **HostCheckEnabled**: Boolean flag for hostname verification
3. **DomainName**: List of domain names for server certificate validation

#### Default SSL Profile

When BackendTLS features are configured in RouteBackendExtension, AKO automatically attaches a **System-Standard** SSL Profile to the backend pool configuration. This profile provides:

**SSL Profile Configuration:**
- **Purpose**: Enables SSL/TLS re-encryption for traffic to backend servers
- **Profile Name**: `System-Standard` (system-generated)
- **Cipher Suites**: Includes modern, secure cipher suites for backend communication
- **SSL Versions**: Supports TLS 1.2 and TLS 1.3 for secure backend connections
- **Certificate Validation**: Works in conjunction with PKIProfile for certificate verification

**Automatic Attachment Behavior:**
- **Trigger**: Automatically attached when any BackendTLS configuration is present in RouteBackendExtension
- **Scope**: Applied to all backend servers in the pool associated with the route
- **Management**: Managed by AKO CRD Operator - no manual configuration required
- **Updates**: Profile settings are updated automatically when BackendTLS configuration changes

**SSL Profile Features:**
The System-Standard includes the following security features:

| Feature | Configuration | Description |
|---------|---------------|-------------|
| **SSL Versions** | TLS 1.2, TLS 1.3 | Modern TLS versions for secure communication |
| **Cipher Suites** | ECDHE-RSA-AES256-GCM-SHA384, ECDHE-RSA-AES128-GCM-SHA256, etc. | Strong encryption algorithms |
| **Perfect Forward Secrecy** | Enabled | Ensures session keys are not compromised if private key is compromised |
| **SSL Session Reuse** | Enabled | Improves performance by reusing SSL sessions |
| **SSL Session Timeout** | 86400 seconds (24 hours) | Balances security and performance |

**Integration with PKIProfile:**
When both System-Standard and PKIProfile are configured:

1. **SSL Handshake**: System-Standard handles the SSL/TLS protocol negotiation
2. **Certificate Validation**: PKIProfile validates the server certificate against trusted CAs
3. **Hostname Verification**: HostCheckEnabled setting controls hostname validation
4. **Domain Matching**: DomainName list is used for certificate subject validation

**Example Pool Configuration:**
When BackendTLS is configured, the resulting Avi pool will have:

```json
{
  "pool": {
    "name": "cluster-namespace-service-port",
    "ssl_profile_ref": "/api/sslprofile?name=system-standard",
    "pki_profile_ref": "/api/pkiprofile?name=cluster-namespace-pkiprofile-name",
    "ssl_key_and_certificate_ref": null,
    "host_check_enabled": true,
    "server_name": "",
    "sni_enabled": false,
    "rewrite_host_header_to_server_name": false,
    "domain_name": ["backend.example.com", "api.example.com"]
  }
}
```

## PKIProfile CRD

The PKIProfile CRD manages Certificate Authorities (Root and Intermediate) used for backend certificate validation.

### Example Usage

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: PKIProfile
metadata:
  name: backend-ca-profile
  namespace: default
spec:
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

## RouteBackendExtension CRD

The RouteBackendExtension CRD extends the Gateway API with advanced load balancing, persistence, health monitoring, and TLS configurations.

### Key Features

- **Load Balancing Algorithms**: Support for various algorithms including round-robin, least connections, consistent hash
- **Session Persistence**: Client IP, HTTP cookie, TLS, and application cookie persistence
- **Health Monitoring**: Integration with Avi health monitors
- **BackendTLS**: Secure communication with backend services

### BackendTLS Configuration

```yaml
# BackendTLS defines the TLS/SSL configuration for secure communication with backend servers
backendTLS:
  # PKI Profile for certificate validation
  pkiProfile:
    kind: "CRD"  # Must be "CRD" for PKIProfile CRD references
    name: "backend-ca-profile"
  
  # Enable hostname verification during TLS handshake
  hostCheckEnabled: true
  
  # Domain names for backend certificate validation
  # Note: domainName can only be configured when hostCheckEnabled is set to true
  domainName:
    - "backend.example.com"
    - "api.example.com"
```

## Usage Examples

### Complete BackendTLS Configuration

```yaml
# First, create a PKIProfile with your CA certificates
apiVersion: ako.vmware.com/v1alpha1
kind: PKIProfile
metadata:
  name: backend-ca-profile
  namespace: production
spec:
  ca_certs:
    - certificate: |
        -----BEGIN CERTIFICATE-----
        # Your CA certificate content here
        -----END CERTIFICATE-----

---
# Create a RouteBackendExtension with BackendTLS configuration
apiVersion: ako.vmware.com/v1alpha1
kind: RouteBackendExtension
metadata:
  name: secure-backend-config
  namespace: production
spec:
  # Load balancing configuration
  lbAlgorithm: LB_ALGORITHM_LEAST_CONNECTIONS
  
  # Session persistence
  persistenceProfile: System-Persistence-Client-IP
  
  # Health monitoring
  healthMonitor:
    - kind: "AVIREF"
      name: "custom-health-monitor"
  
  # Backend TLS configuration
  backendTLS:
    pkiProfile:
      kind: "CRD"
      name: "backend-ca-profile"
    hostCheckEnabled: true
    domainName:
      - "api.production.example.com"
      - "backend.production.example.com"

---
# Reference the RouteBackendExtension in your HTTPRoute
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: secure-api-route
  namespace: production
spec:
  parentRefs:
    - name: api-gateway
  hostnames:
    - "api.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: "/api/v1"
      backendRefs:
        - name: api-service
          port: 443
          filters:
            - type: ExtensionRef
              extensionRef:
                group: ako.vmware.com
                kind: RouteBackendExtension
                name: secure-backend-config
```

## Troubleshooting

### Common Issues

#### 1. PKIProfile Not Found
**Symptom**: RouteBackendExtension shows error about missing PKIProfile

**Solution**: 
- Verify PKIProfile exists in the same namespace
- Check PKIProfile name spelling in RouteBackendExtension
- Ensure PKIProfile has valid CA certificates

```bash
kubectl get pkiprofiles -n <namespace>
kubectl describe pkiprofile <name> -n <namespace>
```

#### 2. Domain Name Validation Error
**Symptom**: Validation error about domainName configuration

**Solution**: 
- Ensure `hostCheckEnabled` is set to `true` when using `domainName`
- Verify domain names match your backend service certificates

#### 3. Controller Not Running
**Symptom**: CRDs not being processed

**Solution**:
- Check AKO CRD Operator pod status
- Review controller logs for errors
- Verify RBAC permissions

```bash
kubectl get pods -n avi-system
kubectl logs -f deployment/ako-crd-operator -n avi-system
```

#### 4. TLS Handshake Failures
**Symptom**: Backend connection failures with TLS errors

**Solution**:
- Verify CA certificates in PKIProfile match backend certificates
- Check domain names in RouteBackendExtension match backend certificate SANs
- Ensure backend services are correctly configured for TLS

### Debug Commands

```bash
# Check CRD status
kubectl get pkiprofiles,routebackendextensions -A

# Describe resources for detailed information
kubectl describe pkiprofile <name> -n <namespace>
kubectl describe routebackendextension <name> -n <namespace>

# Check AKO CRD Operator logs
kubectl logs deployment/ako-crd-operator -n avi-system -f

# Verify Avi ALB configuration
# (Access Avi Controller UI to verify Pool and SSL configurations)
```

### Best Practices

1. **Certificate Management**
   - Keep CA certificates up to date
   - Use proper certificate rotation procedures
   - Validate certificate chains before deployment

2. **Namespace Organization**
   - Keep PKIProfiles and RouteBackendExtensions in the same namespace
   - Use descriptive names for better management
   - Apply appropriate RBAC controls

3. **Monitoring**
   - Monitor controller logs for errors
   - Set up alerts for CRD validation failures
   - Track certificate expiration dates

4. **Security**
   - Restrict access to PKIProfile resources
   - Use least privilege principle for service accounts
   - Regularly audit TLS configurations

