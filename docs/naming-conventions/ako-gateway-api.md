### Naming Conventions:

AKO Gateway Implementation follows the following naming conventions:

  1. **ParentVS**: `ako-gw-<cluster-name>--<gateway-namespace>-<gateway-name>-EVH`
  
  2. **ChildVS**: 
     - **With Named Route Rules** (AKO 2.1.1+, Gateway API v1.3+/v1.4+): `ako-gw-<cluster-name>–-<sha1 hash of <gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-<stringified FNV1a_32 hash of bytes(rulename)>>`
     - **Without Named Route Rules** (Legacy): `ako-gw-<cluster-name>–-<sha1 hash of <gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-<stringified FNV1a_32 hash of bytes(jsonified match)>>`
  
  3. **Pool**: 
     - **With Named Route Rules** (AKO 2.1.1+, Gateway API v1.3+/v1.4+): `ako-gw-<cluster-name>--<sha1 hash of <gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-<stringified FNV1a_32 hash of bytes(rulename)>-<backendRefs_namespace>-<backendRefs_name>-<backendRefs_port>>`
     - **Without Named Route Rules** (Legacy): `ako-gw-<cluster-name>--<sha1 hash of <gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-<stringified FNV1a_32 hash of bytes(jsonified match)>-<backendRefs_namespace>-<backendRefs_name>-<backendRefs_port>>`
  
  4. **PoolGroup**: 
     - **With Named Route Rules** (AKO 2.1.1+, Gateway API v1.3+/v1.4+): `ako-gw-<cluster-name>–-<sha1 hash of <gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-<stringified FNV1a_32 hash of bytes(rulename)>>`
     - **Without Named Route Rules** (Legacy): `ako-gw-<cluster-name>–-<sha1 hash of <gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-<stringified FNV1a_32 hash of bytes(jsonified match)>>`
  
  5. **SSLKeyAndCertificate**: `ako-gw-<cluster-name>--<sha1 hash of <gateway-namespace>-<gateway-name>-<secret-namespace>-<secret-name>>`
  
  6. **DefaultHTTPPolicySet**: `ako-gw-<cluster-name>--default-backend`