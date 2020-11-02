module github.com/vmware/load-balancer-and-ingress-services-for-kubernetes

require (
	github.com/Masterminds/semver v1.5.0
	github.com/avinetworks/sdk v0.0.0-20200910070359-d9ffda19a7dd
	github.com/coreos/etcd v3.3.25+incompatible // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/onsi/gomega v1.8.1
	github.com/openshift/api v0.0.0-20201019163320-c6a5ec25f267
	github.com/openshift/client-go v0.0.0-20201020082437-7737f16e53fc
	github.com/vmware-tanzu/service-apis v0.0.0-20200901171416-461d35e58618
	go.uber.org/zap v1.15.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
)

go 1.13
