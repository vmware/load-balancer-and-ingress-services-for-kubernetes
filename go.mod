module github.com/vmware/load-balancer-and-ingress-services-for-kubernetes

go 1.23.0

toolchain go1.24.1

require (
	github.com/Masterminds/semver v1.5.0
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/davecgh/go-spew v1.1.1
	github.com/fatih/structs v1.1.0
	github.com/fsnotify/fsnotify v1.7.0
	github.com/go-logr/logr v1.2.4
	github.com/gorilla/mux v1.8.0
	github.com/jinzhu/copier v0.3.5
	github.com/jupp0r/go-priority-queue v0.0.0-20160601094913-ab1073853bde
	github.com/onsi/gomega v1.27.10
	github.com/openshift/api v0.0.0-20201019163320-c6a5ec25f267
	github.com/openshift/client-go v0.0.0-20201020082437-7737f16e53fc
	github.com/prometheus/client_golang v1.17.0
	github.com/vmware-tanzu/service-apis v0.0.0-20200901171416-461d35e58618
	github.com/vmware/alb-sdk v0.0.0-20250407104238-8f88209a8f1a
	go.uber.org/zap v1.26.0
	golang.org/x/crypto v0.36.0
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
	google.golang.org/protobuf v1.33.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	istio.io/client-go v1.10.0
	k8s.io/api v0.28.3
	k8s.io/apiextensions-apiserver v0.28.3
	k8s.io/apimachinery v0.28.3
	k8s.io/client-go v0.28.3
	k8s.io/klog/v2 v2.100.1
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b
	sigs.k8s.io/controller-runtime v0.16.3
	sigs.k8s.io/gateway-api v1.0.0
	sigs.k8s.io/service-apis v0.1.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/evanphx/json-patch v5.7.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.7.0 // indirect
	github.com/go-logr/zapr v1.2.4 // indirect
	github.com/go-openapi/jsonpointer v0.20.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/oauth2 v0.13.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	istio.io/api v0.0.0-20210512213424-c42041d3366d // indirect
	istio.io/gogo-genproto v0.0.0-20210113155706-4daf5697332f // indirect
	k8s.io/component-base v0.28.3 // indirect
	k8s.io/kube-openapi v0.0.0-20231010175941-2dd684a91f00 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.3.0 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace (
	github.com/davecgh/go-spew => github.com/davecgh/go-spew v1.1.1
	github.com/golang/glog => github.com/golang/glog v1.2.4
	github.com/google/gofuzz => github.com/google/gofuzz v1.2.0
	github.com/onsi/gomega => github.com/onsi/gomega v1.14.0
	golang.org/x/oauth2 => golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58
	k8s.io/utils => k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	sigs.k8s.io/service-apis => sigs.k8s.io/service-apis v0.1.0
)
