module github.com/vmware/load-balancer-and-ingress-services-for-kubernetes

require (
	github.com/Masterminds/semver v1.5.0
	github.com/avinetworks/ako v0.0.0-20200818183048-9235bc726579 // indirect
	github.com/avinetworks/container-lib v0.0.0-20200805113307-80c6b5ecc46e // indirect
	github.com/avinetworks/sdk v0.0.0-20200812060914-ba100c75801c
	github.com/davecgh/go-spew v1.1.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/gorilla/mux v1.7.4
	github.com/onsi/gomega v1.7.0
	github.com/openshift/api v0.0.0-20200311183032-85e16cc5dd7c
	github.com/openshift/client-go v0.0.0-20191022152013-2823239d2298
	go.uber.org/zap v1.15.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	k8s.io/api v0.17.1
	k8s.io/apimachinery v0.17.1
	k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
)

go 1.13
