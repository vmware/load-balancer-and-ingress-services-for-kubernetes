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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type K8sNodeStore struct {
	*ObjectMapStore
}

var nodeonce sync.Once
var nodesStoreInstance *ObjectMapStore

func SharedNodeLister() *K8sNodeStore {
	nodeonce.Do(func() {
		nodesStoreInstance = NewObjectMapStore()
	})
	return &K8sNodeStore{nodesStoreInstance}
}

func (o *K8sNodeStore) PopulateAllNodes(cs *kubernetes.Clientset, isNodePort bool, nodeLabels map[string]string) {
	labelOption := metav1.ListOptions{}
	if isNodePort {
		if len(nodeLabels) == 2 && nodeLabels["key"] != "" {
			nodeLabelSelector := make(map[string]string)
			nodeLabelSelector[nodeLabels["key"]] = nodeLabels["value"]
			str := labels.Set(nodeLabelSelector).String()
			labelOption.LabelSelector = str
		}
	}
	//filter out nodes if labels are set for nodeport mode
	allNodes, err := cs.CoreV1().Nodes().List(context.TODO(), labelOption)
	if err == nil {
		utils.AviLog.Infof("Got %d nodes", len(allNodes.Items))
		for i, node := range allNodes.Items {
			o.AddOrUpdate(node.Name, &allNodes.Items[i])
		}
	}
}
