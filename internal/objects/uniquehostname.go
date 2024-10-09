/*
 * Copyright 2024 VMware, Inc.
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
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var uniqueHostListerInstance *UniqueHostNamespaceLister
var uniqueHostListerOnce sync.Once

func SharedUniqueNamespaceLister() *UniqueHostNamespaceLister {
	uniqueHostListerOnce.Do(func() {
		uniqueHostListerInstance = &UniqueHostNamespaceLister{
			HostnameToRoutesStore: NewObjectMapStore(),
			RouteToHostnamesStore: NewObjectMapStore(),
		}
	})
	return uniqueHostListerInstance
}

type UniqueHostNamespaceLister struct {
	HostnameLock sync.RWMutex
	// hostname --> Active Routes
	HostnameToRoutesStore *ObjectMapStore
	// Route (namespace/name) --> Hostnames (not being used)
	RouteToHostnamesStore *ObjectMapStore
}

type RouteNamspaceName struct {
	// each entry will be objecttype/route-namespace/route-name
	RouteNSRouteName string
	// store creation time
	CreationTime metav1.Time
}

// key --> Routenamespacename -> string
// value --> creation time  --> string
type RouteList struct {
	activeRoutes map[string]metav1.Time
	// Commenting it for now. In future if we decide to dynamically update the lists, unblock following
	/*
		inactiveRoutes  map[string]metav1.Time
		displacedRoutes map[string]metav1.Time
	*/
}

// =======hostname to route mapping===
// == This function should be called in add or update (in update also for newhost name ) for each hostname of the route
// Keeping signature same for future use.
func (n *UniqueHostNamespaceLister) UpdateHostnameToRoute(hostName string, route RouteNamspaceName) (bool, []string, []string) {
	n.HostnameLock.Lock()
	defer n.HostnameLock.Unlock()
	var localRouteList RouteList

	// fetch existing data first
	utils.AviLog.Debugf("Hostname is: %s", hostName)
	utils.AviLog.Debugf("Current route that needs to be processed is : %v", utils.Stringify(route))
	present, routeList := n.HostnameToRoutesStore.Get(hostName)
	utils.AviLog.Debugf("Details fetched from HostnameToRoutesStore is: Present: %v and routeList: %v", present, utils.Stringify(routeList))
	if !present {
		// host not present, add it to activeRoute list
		// Allocate map
		localRouteList.activeRoutes = make(map[string]metav1.Time)

		localRouteList.activeRoutes[route.RouteNSRouteName] = route.CreationTime
		n.HostnameToRoutesStore.AddOrUpdate(hostName, localRouteList)
		utils.AviLog.Debugf("Active routes after appending: %v", utils.Stringify(localRouteList.activeRoutes))
		return true, nil, nil
	}
	existingRoutes := routeList.(RouteList)
	// Check: route is presnt in active list or not
	if _, found := existingRoutes.activeRoutes[route.RouteNSRouteName]; found {
		utils.AviLog.Debugf("key %s found in active routes.", route.RouteNSRouteName)
		// ingest only that route
		return true, nil, nil
	}

	// Key: not found in active , now compare namespace present in activeRoutes
	sameNamespace := false
	currentRouteNamespace, _ := ExtractNamespaceName(route.RouteNSRouteName)
	// TODO: Debug statements can be reduced
	utils.AviLog.Debugf("Namespace of current route is: %s", currentRouteNamespace)

	// go through each active route and compare namespace with active routes.
	// Active route should contains routes from same namespace
	// Need to check in case of delete of hostname, are we deleting all these list and entry from internal datastore.
	for nsName, creationTime := range existingRoutes.activeRoutes {
		utils.AviLog.Debugf("NamespaceName from active route list is: %v and creation timestamp: %v", nsName, creationTime)
		existingActiveRoutesNamespace, _ := ExtractNamespaceName(nsName)
		utils.AviLog.Debugf("Extracted out namespace is: %s", existingActiveRoutesNamespace)
		if currentRouteNamespace == existingActiveRoutesNamespace {
			// add it to active list
			sameNamespace = true
			break
		}
	}
	// With new logic, we need to check length of active route list. if it is empty then, we need to take another namespace

	if sameNamespace {
		utils.AviLog.Debugf("Namespace is same. Attaching Route %v to active route list", route.RouteNSRouteName)
		existingRoutes.activeRoutes[route.RouteNSRouteName] = route.CreationTime
		n.HostnameToRoutesStore.AddOrUpdate(hostName, existingRoutes)
		// DO not pass anything. Just return true so to process given route
		return true, nil, nil
	}
	utils.AviLog.Debugf("Route %v is not accepted.", route.RouteNSRouteName)
	return false, nil, nil
}

func ExtractNamespaceName(nsName string) (string, string) {
	if nsName != "" {
		// changed from Split to SplitN
		splitStrings := strings.SplitN(nsName, "/", 3)
		if len(splitStrings) == 3 {
			return splitStrings[1], splitStrings[2]
		}
		if len(splitStrings) == 2 {
			return splitStrings[0], splitStrings[1]
		}
	}
	return "", ""
}

// This has to be called for each hostname of the route
func (n *UniqueHostNamespaceLister) DeleteHostnameToRoute(hostName string, route RouteNamspaceName) ([]string, string) {
	n.HostnameLock.Lock()
	defer n.HostnameLock.Unlock()
	present, routeList := n.HostnameToRoutesStore.Get(hostName)
	if !present {
		// hostname entry is not present in the list
		utils.AviLog.Warnf("Returning as host name %s is not present in store", hostName)
		return nil, ""
	}
	existingRoutes := routeList.(RouteList)
	// Check route is present in active list
	if _, flag := existingRoutes.activeRoutes[route.RouteNSRouteName]; flag {
		// present
		utils.AviLog.Debugf("RouteNS to be deleted: %v and Active routes are: %v", route.RouteNSRouteName, existingRoutes.activeRoutes)
		delete(existingRoutes.activeRoutes, route.RouteNSRouteName)

		utils.AviLog.Debugf("After deleting Active routes are %v", utils.Stringify(existingRoutes.activeRoutes))
		if len(existingRoutes.activeRoutes) != 0 {
			n.HostnameToRoutesStore.AddOrUpdate(hostName, existingRoutes)
			// TODO: It will good fetch data and print it for debugging.
			//_, temproutes := n.HostnameToRoutesStore.Get(hostName)
			return nil, route.RouteNSRouteName
		}
		utils.AviLog.Debugf("Active route list is empty for hostname %s. Deleting map entry", hostName)
		n.HostnameToRoutesStore.Delete(hostName)
	}

	return nil, ""
}

//=== route to hostname mapping (Not being used. Done for future use)====

// Add/update list of hostnames associated with given route
func (n *UniqueHostNamespaceLister) UpdateRouteToHostname(routeNSName string, hostNames []string) {
	n.HostnameLock.Lock()
	defer n.HostnameLock.Unlock()
	n.RouteToHostnamesStore.AddOrUpdate(routeNSName, hostNames)
}

func (n *UniqueHostNamespaceLister) GetRouteToHostname(routeName string) (bool, []string) {
	n.HostnameLock.RLock()
	defer n.HostnameLock.RUnlock()
	found, hostNameList := n.RouteToHostnamesStore.Get(routeName)
	if found {
		return true, hostNameList.([]string)
	}
	return false, []string{}
}

func (n *UniqueHostNamespaceLister) DeleteRouteToHostname(routeName string) {
	n.HostnameLock.Lock()
	defer n.HostnameLock.Unlock()
	n.RouteToHostnamesStore.Delete(routeName)
}
