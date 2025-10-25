# ApplicationProfile CRD Documentation

## Table of Contents
- [Introduction](#introduction)
- [ApplicationProfile Specification](#applicationprofile-specification)
- [Spec Fields](#spec-fields)
- [Status Fields](#status-fields)
- [Usage Examples](#usage-examples)
- [Integration with Virtual Services](#integration-with-virtual-services)
- [Troubleshooting](#troubleshooting)

## Introduction

The ApplicationProfile Custom Resource Definition (CRD) manages HTTP application layer proxy configurations for Avi Load Balancer Virtual Services. ApplicationProfile enables fine-grained control over HTTP behavior, connection management, and client IP handling in modern cloud-native architectures.

**NOTE**: ApplicationProfile CRD is specifically designed for use with Avi Load Balancer and is handled by the ako-crd-operator. This CRD supports only HTTP application profile types in the current release and is applicable only in the Gateway API context.

A sample ApplicationProfile CRD looks like this:

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: ApplicationProfile
metadata:
  name: applicationprofile-sample
  labels:
    app.kubernetes.io/name: ako-crd-operator
    app.kubernetes.io/managed-by: kustomize
spec:
  type: "HTTP"
  http_profile:
    connection_multiplexing_enabled: true
    xff_enabled: true
    xff_alternate_name: "X-Forwarded-For-Custom"
    xff_update: "REPLACE_XFF_HEADERS"
    x_forwarded_proto_enabled: false
    client_body_timeout: 30000
    keepalive_timeout: 30000
    client_max_body_size: 0
    keepalive_header: false
    use_app_keepalive_timeout: false
    max_keepalive_requests: 100
    reset_conn_http_on_ssl_port: false
    http_upstream_buffer_size: 0
    enable_chunk_merge: true
    use_true_client_ip: true
    true_client_ip:
      headers:
        - "X-Forwarded-For"
      index_in_header: 1
      direction: "LEFT"
    max_header_count: 256
    close_server_side_connection_on_error: false
```

## ApplicationProfile Specification

The ApplicationProfile CRD is defined with the following structure:

- **API Version**: `ako.vmware.com/v1alpha1`
- **Kind**: `ApplicationProfile`
- **Scope**: `Namespaced`
- **Short Name**: `ap`
- **Plural**: `applicationprofiles`
- **Singular**: `applicationprofile`

## Spec Fields

The ApplicationProfile CRD supports the following configuration options:

### Type

The `type` field specifies which application layer proxy is enabled for the virtual service. Currently, only HTTP application profile type is supported.

```yaml
type: "HTTP"
```

- Valid values: `HTTP`
- **Note**: This field is required and immutable once the ApplicationProfile is created.

### HTTP Profile

The `http_profile` section specifies the HTTP application proxy profile parameters. This section contains all HTTP-specific configuration options.

```yaml
http_profile:
  connection_multiplexing_enabled: true
  xff_enabled: true
  # ... other HTTP profile settings
```

#### Connection Multiplexing Enabled

The `connection_multiplexing_enabled` field allows HTTP requests, not just TCP connections, to be load balanced across servers. Proxied TCP connections to servers may be reused by multiple clients to improve performance.

```yaml
connection_multiplexing_enabled: true
```

- Default: `true`
- **Note**: Not compatible with Preserve Client IP

#### XFF Enabled

The `xff_enabled` field enables insertion of the client's original IP address into an HTTP request header sent to the server. Servers may use this address for logging or other purposes, rather than Avi's source NAT address.

```yaml
xff_enabled: true
```

- Default: `true`

#### XFF Alternate Name

The `xff_alternate_name` field provides a custom name for the X-Forwarded-For header sent to the servers.

```yaml
xff_alternate_name: "X-Forwarded-For-Custom"
```

- Default: "X-Forwarded-For"
- **Note**: Can only be configured if `xff_enabled` is true

#### XFF Update

The `xff_update` field configures how incoming X-Forwarded-For headers from the client are handled.

```yaml
xff_update: "REPLACE_XFF_HEADERS"
```

Valid values are:
- `REPLACE_XFF_HEADERS`: Replace all incoming X-Forward-For headers with the Avi created header
- `APPEND_TO_THE_XFF_HEADER`: All incoming X-Forwarded-For headers will be appended to the Avi created header
- `ADD_NEW_XFF_HEADER`: Simply add a new X-Forwarded-For header

- Default: `REPLACE_XFF_HEADERS`
- **Note**: Can only be configured if `xff_enabled` is true

#### X Forwarded Proto Enabled

The `x_forwarded_proto_enabled` field inserts an X-Forwarded-Proto header in the request sent to the server. When the client connects via SSL, Avi terminates the SSL, and then forwards the requests to the servers via HTTP.

```yaml
x_forwarded_proto_enabled: false
```

- Default: `false`

#### Client Body Timeout

The `client_body_timeout` field defines the maximum length of time allowed between consecutive read operations for a client request body. The value '0' specifies no timeout.

```yaml
client_body_timeout: 30000
```

- Valid range: 0-100000000 milliseconds
- Default: 30000 milliseconds

#### Keepalive Timeout

The `keepalive_timeout` field defines the max idle time allowed between HTTP requests over a Keep-alive connection.

```yaml
keepalive_timeout: 30000
```

- Valid range: 10-100000000 milliseconds
- Default: 30000 milliseconds

#### Use App Keepalive Timeout

The `use_app_keepalive_timeout` field enables use of 'Keep-Alive' header timeout sent by application instead of sending the HTTP Keep-Alive Timeout.

```yaml
use_app_keepalive_timeout: false
```

- Default: `false`

#### Client Max Body Size

The `client_max_body_size` field defines the maximum size for the client request body. This limits the size of the client data that can be uploaded/posted as part of a single HTTP Request.

```yaml
client_max_body_size: 0
```

- Units: KB
- Default: 0 (Unlimited)

#### Keepalive Header

The `keepalive_header` field enables sending HTTP 'Keep-Alive' header to the client. By default, the timeout specified in the 'Keep-Alive Timeout' field will be used unless the 'Use App Keepalive Timeout' flag is set.

```yaml
keepalive_header: false
```

- Default: `false`

#### Max Keepalive Requests

The `max_keepalive_requests` field defines the max number of HTTP requests that can be sent over a Keep-Alive connection.

```yaml
max_keepalive_requests: 100
```

- Valid range: 0-1000000
- Default: 100
- **Note**: 0 means unlimited requests on a connection

#### Reset Conn HTTP On SSL Port

The `reset_conn_http_on_ssl_port` field enables connection close instead of a 400 response when an HTTP request is made on an SSL port.

```yaml
reset_conn_http_on_ssl_port: false
```

- Default: `false`

#### HTTP Upstream Buffer Size

The `http_upstream_buffer_size` field defines the size of HTTP buffer in kB.

```yaml
http_upstream_buffer_size: 0
```

- Valid range: 0-256 KB
- Default: 0 (Auto compute the size of buffer)

#### Enable Chunk Merge

The `enable_chunk_merge` field enables chunk body merge for chunked transfer encoding response.

```yaml
enable_chunk_merge: true
```

- Default: `true`

#### Use True Client IP

The `use_true_client_ip` field enables detection of client IP from user specified header.

```yaml
use_true_client_ip: true
```

- Default: `false`

#### True Client IP

The `true_client_ip` field configures detection of client IP from user specified header at the configured index in the specified direction.

```yaml
true_client_ip:
  headers:
    - "X-Forwarded-For"
  index_in_header: 1
  direction: "LEFT"
```

**Headers**: HTTP Headers to derive client IP from. If none specified and `use_true_client_ip` is set to true, it will use X-Forwarded-For header, if present.
- Maximum items: 1
- Maximum length per header: 128 characters

**Index In Header**: Position in the configured direction, in the specified header's value, to be used to set true client IP.
- Valid range: 1-1000

**Direction**: Denotes the end from which to count the IPs in the specified header value.
- Valid values: `LEFT`, `RIGHT`
- Default: `LEFT`

- **Note**: Can only be configured if `use_true_client_ip` is true

#### Max Header Count

The `max_header_count` field defines the maximum number of headers allowed in HTTP request and response.

```yaml
max_header_count: 256
```

- Valid range: 0-4096
- Default: 256
- **Note**: 0 means unlimited headers in request and response

#### Close Server Side Connection On Error

The `close_server_side_connection_on_error` field enables closing server-side connection when an error response is received.

```yaml
close_server_side_connection_on_error: false
```

- Default: `false`

#### Status Fields

- **uuid** - Unique identifier of the application profile object in the Avi Controller
- **observedGeneration** - The generation of the ApplicationProfile resource that was most recently processed by the AKO CRD Operator
- **lastUpdated** - Timestamp when the application profile object was last updated in the Avi Controller
- **backendObjectName** - Name of the application profile object created in the Avi Controller
- **tenant** - The tenant where the application profile is created
- **controller** - Field is populated by AKO CRD operator as ako-crd-operator
- **conditions** - List of conditions for the application profile

#### Status Conditions

The ApplicationProfile status includes a `Programmed` condition that indicates whether the application profile has been successfully processed:

**Condition Type: Programmed**

- **Status: True** - The application profile has been successfully programmed on the Avi Controller
  - Reasons: `Created`, `Updated`
  
- **Status: False** - The application profile failed to be programmed
  - Reasons: `CreationFailed`, `UpdateFailed`, `UUIDExtractionFailed`, `DeletionFailed`, `DeletionSkipped`

Example of a failed status:

```yaml
status:
  conditions:
    - type: Programmed
      status: "False"
      reason: CreationFailed
      message: "Failed to create ApplicationProfile on Avi Controller: invalid configuration"
      lastTransitionTime: "2025-10-08T10:30:00Z"
```

## Usage Examples

### Basic ApplicationProfile Configuration

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: ApplicationProfile
metadata:
  name: basic-http-profile
  namespace: production
spec:
  type: "HTTP"
  http_profile:
    connection_multiplexing_enabled: true
    xff_enabled: true
    keepalive_timeout: 30000
    client_body_timeout: 30000
```

### Advanced ApplicationProfile with Client IP Detection

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: ApplicationProfile
metadata:
  name: advanced-http-profile
  namespace: production
  labels:
    app.kubernetes.io/name: ako-crd-operator
    app.kubernetes.io/managed-by: kustomize
spec:
  type: "HTTP"
  http_profile:
    connection_multiplexing_enabled: true
    xff_enabled: true
    xff_alternate_name: "X-Real-IP"
    xff_update: "REPLACE_XFF_HEADERS"
    x_forwarded_proto_enabled: true
    client_body_timeout: 60000
    keepalive_timeout: 60000
    client_max_body_size: 1024
    keepalive_header: true
    use_app_keepalive_timeout: false
    max_keepalive_requests: 200
    reset_conn_http_on_ssl_port: true
    http_upstream_buffer_size: 64
    enable_chunk_merge: true
    use_true_client_ip: true
    true_client_ip:
      headers:
        - "X-Forwarded-For"
      index_in_header: 1
      direction: "LEFT"
    max_header_count: 512
    close_server_side_connection_on_error: true
```

### Production ApplicationProfile with Custom Headers

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: ApplicationProfile
metadata:
  name: production-http-profile
  namespace: production
  labels:
    app.kubernetes.io/name: ako-crd-operator
    app.kubernetes.io/managed-by: kustomize
    environment: production
    team: platform
spec:
  type: "HTTP"
  http_profile:
    connection_multiplexing_enabled: true
    xff_enabled: true
    xff_alternate_name: "X-Forwarded-For"
    xff_update: "APPEND_TO_THE_XFF_HEADER"
    x_forwarded_proto_enabled: true
    client_body_timeout: 30000
    keepalive_timeout: 30000
    client_max_body_size: 0
    keepalive_header: true
    use_app_keepalive_timeout: false
    max_keepalive_requests: 100
    reset_conn_http_on_ssl_port: false
    http_upstream_buffer_size: 0
    enable_chunk_merge: true
    use_true_client_ip: false
    max_header_count: 256
    close_server_side_connection_on_error: false
```

## Integration with Gateway API

ApplicationProfile is designed to work with Gateway API HTTPRoute resources for HTTP load balancing. The ApplicationProfile is referenced in HTTPRoute filter configurations to define HTTP behavior.

### Using ApplicationProfile with HTTPRoute

ApplicationProfile can be referenced in Gateway API HTTPRoute resources using ExtensionRef filters:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: my-route
  namespace: default
spec:
  parentRefs:
    - name: my-gateway
  hostnames:
    - "app.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: "/"
      filters:
        - type: ExtensionRef
          extensionRef:
            group: ako.vmware.com
            kind: ApplicationProfile
            name: my-application-profile
      backendRefs:
        - name: my-service
          port: 8080
```

### ApplicationProfile Processing

When an ApplicationProfile is referenced in an HTTPRoute:

1. **Validation**: The AKO Gateway API controller validates that the ApplicationProfile exists and is in a "Programmed" state
2. **UUID Resolution**: The controller retrieves the ApplicationProfile UUID from its status
3. **Virtual Service Configuration**: The ApplicationProfile UUID is applied to the corresponding Avi Virtual Service as an ApplicationProfileRef
4. **Mapping Management**: The controller maintains bidirectional mappings between ApplicationProfile and HTTPRoute resources

### ApplicationProfile Event Handling

The AKO Gateway API controller monitors ApplicationProfile changes and automatically processes affected HTTPRoute resources:

- **Add/Update Events**: When an ApplicationProfile is created or updated, all associated HTTPRoute resources are re-processed
- **Delete Events**: When an ApplicationProfile is deleted, the ApplicationProfile reference is removed from associated HTTPRoute resources
- **Status Validation**: Only ApplicationProfile resources with "Programmed" status are considered valid for use

## Troubleshooting

### Common Issues

**1. Invalid ApplicationProfile Type**
*Symptom*: ApplicationProfile shows validation error about unsupported type

*Solution*: 
- Ensure type is set to "HTTP" (only supported type in current release)
- Verify the type field is not empty

**2. XFF Configuration Conflicts**
*Symptom*: ApplicationProfile shows validation error about XFF configuration

*Solution*:
- Ensure `xff_alternate_name` and `xff_update` are only configured when `xff_enabled` is true
- If `xff_update` is "APPEND_TO_THE_XFF_HEADER", `xff_alternate_name` must be empty

**3. True Client IP Configuration Issues**
*Symptom*: ApplicationProfile shows validation error about true client IP configuration

*Solution*:
- Ensure `true_client_ip` is only configured when `use_true_client_ip` is true
- Verify header names are at most 128 characters long
- Check that `index_in_header` is within valid range (1-1000)

**4. Controller Not Processing ApplicationProfile**
*Symptom*: ApplicationProfile status shows "Programmed: False"

*Solution*:
- Check AKO CRD Operator pod status
- Review controller logs for errors
- Verify RBAC permissions

```bash
kubectl get pods -n avi-system
kubectl logs -f deployment/ako-crd-operator -n avi-system
```

**5. HTTPRoute Not Using ApplicationProfile**
*Symptom*: HTTPRoute with ApplicationProfile ExtensionRef is not applying the profile

*Solution*:
- Verify ApplicationProfile is in "Programmed" state
- Check HTTPRoute status conditions for ResolvedRefs
- Ensure ApplicationProfile and HTTPRoute are in the same namespace
- Verify ExtensionRef configuration is correct

```bash
kubectl get applicationprofile <name> -o yaml
kubectl describe httproute <name>
```

**6. ApplicationProfile UUID Not Found**
*Symptom*: Gateway API controller logs show "ApplicationProfile UUID not found"

*Solution*:
- Ensure ApplicationProfile has been processed by AKO CRD Operator
- Check that ApplicationProfile status contains a valid UUID
- Verify ApplicationProfile tenant matches namespace tenant

```bash
kubectl get applicationprofile <name> -o jsonpath='{.status.uuid}'
```

### Debug Commands

```bash
# Check ApplicationProfile status
kubectl get applicationprofiles -A
kubectl describe applicationprofile <name> -n <namespace>

# Check ApplicationProfile events
kubectl get events --field-selector involvedObject.name=<applicationprofile-name> -n <namespace>

# Check AKO CRD Operator logs
kubectl logs deployment/ako-crd-operator -n avi-system -f

# Check Gateway API controller logs
kubectl logs deployment/ako-gateway-api -n avi-system -f

# Check HTTPRoute status and conditions
kubectl get httproute <name> -o yaml
kubectl describe httproute <name>

# Verify ApplicationProfile UUID
kubectl get applicationprofile <name> -o jsonpath='{.status.uuid}'

# Check ApplicationProfile to HTTPRoute mappings
kubectl logs deployment/ako-gateway-api -n avi-system | grep "ApplicationProfile"

# Verify Avi Controller configuration
# (Access Avi Controller UI to verify Application Profile configuration)
```

### Status Interpretation

**Programmed: True**
- ApplicationProfile has been successfully created on the Avi Controller
- Ready to be referenced by Virtual Services

**Programmed: False with reason "CreationFailed"**
- Failed to create ApplicationProfile on Avi Controller
- Check controller logs for detailed error information

**Programmed: False with reason "UUIDExtractionFailed"**
- ApplicationProfile was created but UUID extraction failed
- Check Avi Controller connectivity and permissions

### Validation Rules

The ApplicationProfile CRD enforces the following validation rules:

1. **Type Requirement**: The `type` field is required and must be "HTTP"
2. **XFF Configuration**: `xff_alternate_name` and `xff_update` can only be configured if `xff_enabled` is true
3. **XFF Update Constraint**: If `xff_update` is "APPEND_TO_THE_XFF_HEADER", `xff_alternate_name` must be empty
4. **True Client IP Configuration**: `true_client_ip` can only be configured if `use_true_client_ip` is true
5. **Header Length**: Each header name in `true_client_ip.headers` must be at most 128 characters long
6. **Index Range**: `index_in_header` must be between 1 and 1000

### Prerequisites

Before creating an ApplicationProfile CRD:

1. The AKO CRD Operator must be installed and running in the cluster
2. Ensure the Avi Controller is accessible and properly configured
3. Verify that the namespace where the ApplicationProfile will be created exists

### Namespace Considerations

- ApplicationProfile resources must be created in the same namespace as the HTTPRoute that references them using ExtensionRef
- The ApplicationProfile name must be unique within the namespace

### Immutable Fields

The following fields are immutable after the ApplicationProfile is created:

- `type`: The application profile type cannot be changed

To modify these fields, the ApplicationProfile must be deleted and recreated.

### ApplicationProfile Deletion

When an ApplicationProfile is deleted:

- The application profile object is removed from the Avi Controller if it is not being referenced by other objects
- The AKO CRD Operator retries the cleanup automatically

If an ApplicationProfile is stuck in a terminating state:

1. Check if the AKO CRD Operator is running:
   ```bash
   kubectl get pods -n avi-system | grep ako-crd-operator
   ```

2. Verify the application profile on the Avi Controller and check if it's being referenced by other objects (Virtual Services, etc.). If referenced, remove those references first.

3. If the operator is stuck and the Avi Controller object has been manually cleaned up, you can force remove the finalizer:
   ```bash
   kubectl patch applicationprofile <name> -n <namespace> -p '{"metadata":{"finalizers":[]}}' --type=merge
   ```

   **Warning**: Only use this as a last resort when confirmed the backend object is properly cleaned up on the Avi Controller. Removing the finalizer without proper cleanup may leave orphaned objects on the Avi Controller.
