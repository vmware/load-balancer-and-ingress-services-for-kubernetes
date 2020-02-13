/*
* [2013] - [2019] Avi Networks Incorporated
* All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package lib

import (
	"os"
	"os/signal"
	"syscall"

	"gitlab.eng.vmware.com/orion/container-lib/utils"
)

var IngressApiMap = map[string]string{
	"corev1":      utils.CoreV1IngressInformer,
	"extensionv1": utils.ExtV1IngressInformer,
}
var onlyOneSignalHandler = make(chan struct{})

func GetVrf() string {
	vrfcontext := os.Getenv(utils.VRF_CONTEXT)
	if vrfcontext == "" {
		vrfcontext = utils.GlobalVRF
	}
	return vrfcontext
}

func GetIngressApi() string {
	ingressApi := os.Getenv(INGRESS_API)
	ingressApi, ok := IngressApiMap[ingressApi]
	if !ok {
		return utils.CoreV1IngressInformer
	}
	return ingressApi
}

func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGUSR1)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}
