module github.com/vmware/load-balancer-and-ingress-services-for-kubernetes

require (
	github.com/Azure/go-autorest v11.1.2+incompatible // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/avinetworks/sdk v0.0.0-20200910040148-2ed48a016241
	github.com/davecgh/go-spew v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/onsi/gomega v1.8.1
	github.com/openshift/api v0.0.0-20200311183032-85e16cc5dd7c
	github.com/openshift/client-go v0.0.0-20191022152013-2823239d2298
	github.com/vmware-tanzu/service-apis v0.0.0-20200901171416-461d35e58618
	go.uber.org/zap v1.15.0
	golang.org/x/sys v0.0.0-20191105231009-c1f44814a5cd // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.2.5 // indirect
	k8s.io/api v0.17.1
	k8s.io/apimachinery v0.17.1
	k8s.io/client-go v0.17.0
)

go 1.13
