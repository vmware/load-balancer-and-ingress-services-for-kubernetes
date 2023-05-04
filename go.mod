module github.com/vmware/load-balancer-and-ingress-services-for-kubernetes

go 1.18

require (
	github.com/Masterminds/semver v1.5.0
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/davecgh/go-spew v1.1.1
	github.com/fatih/structs v1.1.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-logr/logr v0.4.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/gorilla/mux v1.8.0
	github.com/jinzhu/copier v0.3.5
	github.com/onsi/gomega v1.14.0
	github.com/openshift/api v0.0.0-20201019163320-c6a5ec25f267
	github.com/openshift/client-go v0.0.0-20201020082437-7737f16e53fc
	github.com/vmware-tanzu/service-apis v0.0.0-20200901171416-461d35e58618
	github.com/vmware/alb-sdk v0.0.0-20230202152455-af9d49bac7ea
	go.uber.org/zap v1.18.1
	golang.org/x/crypto v0.0.0-20220314234659-1baeb1ce4c0b
	golang.org/x/exp v0.0.0-20230321023759-10a507213a29
	google.golang.org/protobuf v1.26.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	istio.io/client-go v1.10.0
	k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	k8s.io/klog/v2 v2.8.0
	k8s.io/utils v0.0.0-20210722164352-7f3ee0f31471
	sigs.k8s.io/controller-runtime v0.9.6
	sigs.k8s.io/service-apis v0.1.0
)

require (
	cloud.google.com/go v0.65.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/go-logr/zapr v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jinzhu/copier v0.3.5 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.11.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b // indirect
	golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/term v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20220521103104-8f96da9f5d5e // indirect
	istio.io/api v0.0.0-20210512213424-c42041d3366d // indirect
	istio.io/gogo-genproto v0.0.0-20210113155706-4daf5697332f // indirect
	k8s.io/code-generator v0.21.3 // indirect
	k8s.io/component-base v0.21.3 // indirect
	k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	github.com/davecgh/go-spew => github.com/davecgh/go-spew v1.1.1
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	github.com/golang/glog => github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gofuzz => github.com/google/gofuzz v1.2.0
	github.com/onsi/gomega => github.com/onsi/gomega v1.14.0
	go.uber.org/zap => go.uber.org/zap v1.18.1
	golang.org/x/net => golang.org/x/net v0.7.0
	golang.org/x/oauth2 => golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58
	golang.org/x/text => golang.org/x/text v0.4.0
	google.golang.org/protobuf => google.golang.org/protobuf v1.26.0
	k8s.io/api => k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.3
	k8s.io/client-go => k8s.io/client-go v0.21.3
	k8s.io/utils => k8s.io/utils v0.0.0-20210722164352-7f3ee0f31471
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.9.6
	sigs.k8s.io/service-apis => sigs.k8s.io/service-apis v0.1.0
)
