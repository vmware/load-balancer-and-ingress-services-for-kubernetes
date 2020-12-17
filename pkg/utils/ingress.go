/*
 * Copyright 2019-2020 VMware, Inc.
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

package utils

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var ingressClassEnabled *bool

func SetIngressClassEnabled(kc kubernetes.Interface) {
	if ingressClassEnabled != nil {
		return
	}

	var isPresent bool
	timeout := int64(120)
	// should only work for k8s 1.19+ clusters
	_, ingClassError := kc.NetworkingV1().IngressClasses().List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &timeout})
	if ingClassError != nil {
		AviLog.Infof("networking.k8s.io/v1/IngressClass not found/enabled on cluster: %v", ingClassError)
		isPresent = false
	} else {
		AviLog.Infof("networking.k8s.io/v1/IngressClass enabled on cluster")
		isPresent = true
	}

	ingressClassEnabled = &isPresent
}

func GetIngressClassEnabled() bool {
	return *ingressClassEnabled
}
