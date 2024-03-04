package main

import (
	"flag"
	"os"
	"strings"

	akoclean "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-clean"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var (
	ctrlIP                            = ""
	username                          = ""
	authToken                         = ""
	password                          = ""
	albCtrlCert                       = ""
	clusterID                         = ""
	ExitCodeRequiredArgsMissing       = 1
	ExitCodeCleanupALBResourcesFailed = 2
)

func main() {
	flag.StringVar(&ctrlIP, "ctrl-ip", "", "NSX ALB Controller IP")
	flag.StringVar(&username, "username", "nsxt-alb", "NSX ALB Controller username")
	flag.StringVar(&authToken, "token", "", "NSX ALB Controller authentication token")
	flag.StringVar(&password, "password", "", "NSX ALB Controller authentication password")
	flag.StringVar(&albCtrlCert, "cacert", "", "NSX ALB Controller authentication certificate")
	flag.StringVar(&clusterID, "cluster-id", "", "AKO cluster name")
	flag.Parse()

	cfg := akoclean.NewAKOCleanupConfig(ctrlIP, username, password, authToken, albCtrlCert, clusterID)

	err := cfg.Cleanup()
	if err != nil {
		utils.AviLog.Errorf("Failed to cleanup Avi resources, err: %s", err.Error())
		exitCode := ExitCodeCleanupALBResourcesFailed
		if strings.Contains(err.Error(), "invalid config") {
			exitCode = ExitCodeRequiredArgsMissing
		}
		os.Exit(exitCode)
	}
}
