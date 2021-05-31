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

package objects

import (
	"context"
	"sync"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type CtrlPropStore struct {
	*ObjectMapStore
}

var ctrlproponce sync.Once
var ctrlPropStoreInstance *ObjectMapStore

func SharedCtrlPropLister() *CtrlPropStore {
	ctrlproponce.Do(func() {
		ctrlPropStoreInstance = NewObjectMapStore()
	})
	return &CtrlPropStore{ctrlPropStoreInstance}
}

func (o *CtrlPropStore) PopulateCtrlProp(cs kubernetes.Interface) error {
	aviSecret, err := cs.CoreV1().Secrets("avi-system").Get(context.TODO(), "avi-secret", metav1.GetOptions{})
	if err != nil {
		return err
	}
	ctrlUsername := string(aviSecret.Data["username"])
	o.AddOrUpdate(utils.ENV_CTRL_USERNAME, ctrlUsername)
	if aviSecret.Data["password"] != nil {
		ctrlPassword := string(aviSecret.Data["password"])
		o.AddOrUpdate(utils.ENV_CTRL_PASSWORD, ctrlPassword)
	} else {
		o.AddOrUpdate(utils.ENV_CTRL_PASSWORD, "")
	}
	if aviSecret.Data["authtoken"] != nil {
		ctrlAuthToken := string(aviSecret.Data["authtoken"])
		o.AddOrUpdate(utils.ENV_CTRL_AUTHTOKEN, ctrlAuthToken)
	} else {
		o.AddOrUpdate(utils.ENV_CTRL_AUTHTOKEN, "")
	}
	return nil
}
