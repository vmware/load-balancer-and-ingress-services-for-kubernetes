## Utility Scripts

#### update-codegen-akocrd.sh
This script should be used to generate deepcopy functions, registers, clientsets, listers and informers for AKO CRDs i.e. HostRule and HTTPRule. 
<br/>
Run this script whenever we upgrade AKOs hostrule and httprule CRD versions. Change the `AKOCRD_VERSION` env variable in the script to the intended version. 
<br/>
Usage from the AKO workspace directory, Run `./hack/update-codegen-akocrd.sh`

#### update-codegen-service-apis.sh
This script should be used to generate clientsets, listers and informers for service APIs. The service API structs i.e. Gateway and GatewayClass are taken from vmware-tanzu/service-apis (support/v1alpha0 branch) repository.
<br/>
Run this script whenever we update service-apis types from upstream.
<br/>
Usage from the AKO workspace directory, Run `./hack/update-codegen-service-apis.sh`