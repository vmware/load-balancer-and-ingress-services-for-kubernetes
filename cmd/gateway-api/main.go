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
	Initialize()
}

func Initialize() {
	utils.AviLog.Infof("AKO is running with version: %s", version)
	for {
		time.Sleep(8 * time.Second)
		utils.AviLog.Infof("AKO is processing the objects...")
	}
}
