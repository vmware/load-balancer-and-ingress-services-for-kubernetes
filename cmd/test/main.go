package main

import (
	"strconv"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func main() {

	for i := 0; i < 100; i++ {
		go func(i int) {
			utils.AviLog.WithContext(i).Infof("hello" + strconv.Itoa(i))
		}(i)
	}
	time.Sleep(1 * time.Second)
}

func LoggerWithContext() {
	utils.AviLog.WithContext(1).Infof("hello" + strconv.Itoa(1))
}

func Logger() {
	utils.AviLog.Infof("hello" + strconv.Itoa(1))
}
