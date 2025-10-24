### Naming Conventions:

AKO Gateway Implementation follows following naming convention:

  1. ParentVS              `ako-gw-<cluster-name>--<gateway-namespace>-<gateway-name>-EVH`
  2. ChildVS               `ako-gw-<cluster-name>–-<sha1 hash of <gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-<stringified FNV1a_32 hash of bytes(jsonified match)>>` 
  3. Pool                  `ako-gw-<cluster-name>--<sha1 hash of <gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-<stringified FNV1a_32 hash of bytes(jsonified match)>-<backendRefs_namespace>-<backendRefs_name>-<backendRefs_port>>` 
  4. PoolGroup             `ako-gw-<cluster-name>–-<sha1 hash of <gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-<stringified FNV1a_32 hash of bytes(jsonified match)>>` 
  5. SSLKeyAndCertificate  `ako-gw-<cluster-name>--<sha1 hash of <gateway-namespace>-<gateway-name>-<secret-namespace>-<secret-name>>`
  6. DefaultHTTPPolicySet  `ako-gw-<cluster-name>--default-backend`