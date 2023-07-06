package main

import (
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var (
	masterURL  string
	kubeconfig string
	version    = "dev"
)

func main() {
	InitializeAKCwithGateway()
}

func InitializeAKCwithGateway() {
	if !utils.IsGatewayEnabled() {
		utils.AviLog.Fatalf("Shutting down, Gateway is disabled")
		return
	}
	for {
		time.Sleep(8 * time.Second)
	}

}
