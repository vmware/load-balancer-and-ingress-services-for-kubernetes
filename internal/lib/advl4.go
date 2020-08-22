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

package lib

import (
	advl4crd "github.com/vmware-tanzu/service-apis/pkg/client/clientset/versioned"
	advl4informer "github.com/vmware-tanzu/service-apis/pkg/client/informers/externalversions/apis/v1alpha1pre1"
)

var AdvL4Clientset advl4crd.Interface

// crd "github.com/vmware-tanzu/service-apis/pkg/client/clientset/versioned/typed/ako/v1alpha1"
func SetAdvL4Clientset(cs advl4crd.Interface) {
	AdvL4Clientset = cs
}

func GetAdvL4Clientset() advl4crd.Interface {
	return AdvL4Clientset
}

var AKOAdvL4Informers *AdvL4Informers

type AdvL4Informers struct {
	GatewayInformer      advl4informer.GatewayInformer
	GatewayClassInformer advl4informer.GatewayClassInformer
}

func SetAdvL4Informers(c *AdvL4Informers) {
	AKOAdvL4Informers = c
}

func GetAdvL4Informers() *AdvL4Informers {
	return AKOAdvL4Informers
}
