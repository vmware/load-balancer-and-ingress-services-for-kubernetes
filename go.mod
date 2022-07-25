module github.com/vmware/load-balancer-and-ingress-services-for-kubernetes

go 1.15

require (
	github.com/Masterminds/semver v1.5.0
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/davecgh/go-spew v1.1.1
	github.com/go-logr/logr v0.4.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/onsi/gomega v1.14.0
	github.com/openshift/api v0.0.0-20201019163320-c6a5ec25f267
	github.com/openshift/client-go v0.0.0-20201020082437-7737f16e53fc
	github.com/vmware-tanzu/service-apis v0.0.0-20200901171416-461d35e58618
	github.com/vmware/alb-sdk v0.0.0-20210721142023-8e96475b833b
	go.uber.org/zap v1.18.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58 // indirect
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

replace (
	github.com/davecgh/go-spew => github.com/davecgh/go-spew v1.1.1
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	github.com/golang/glog => github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gofuzz => github.com/google/gofuzz v1.2.0
	github.com/onsi/gomega => github.com/onsi/gomega v1.14.0
	go.uber.org/zap => go.uber.org/zap v1.18.1
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/oauth2 => golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58
	google.golang.org/protobuf => google.golang.org/protobuf v1.26.0
	k8s.io/api => k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.3
	k8s.io/client-go => k8s.io/client-go v0.21.3
	k8s.io/utils => k8s.io/utils v0.0.0-20210722164352-7f3ee0f31471
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.9.6
	sigs.k8s.io/service-apis => sigs.k8s.io/service-apis v0.1.0
)
