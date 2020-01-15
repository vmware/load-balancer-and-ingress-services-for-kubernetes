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

package objects

import (
	"sync"

	"gitlab.eng.vmware.com/orion/container-lib/utils"
	"k8s.io/apimachinery/pkg/labels"
)

type K8sNodeStore struct {
	ObjectMapStore
}

var nodeonce sync.Once
var nodesStoreInstance *ObjectMapStore

func SharedNodeLister() *K8sNodeStore {
	nodeonce.Do(func() {
		nodesStoreInstance = NewObjectMapStore()
	})
	return &K8sNodeStore{*nodesStoreInstance}
}

func (o *K8sNodeStore) PopulateAllNodes() {
	allNodes, _ := utils.GetInformers().NodeInformer.Lister().List(labels.Everything())
	for _, node := range allNodes {
		o.AddOrUpdate(node.Name, node)
	}
}

type NodesCache struct {
	ingSvcobjects *ObjectMapStore
}

func (o *NodesCache) GetAllObjectNames() {

}
