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

var uniqueHostListerInstace *UniqueHostNamespaceLister
var uniqueHostListerOnce sync.Once

func SharedUniqueNamespaceLister() *UniqueHostNamespaceLister {
	uniqueHostListerOnce.Do(func() {
		uniqueHostListerInstace = &UniqueHostNamespaceLister{
			HostnameToRoutesStore: NewObjectMapStore(),
			RouteToHostnamesStore: NewObjectMapStore(),
		}
	})
	return uniqueHostListerInstace
}

type UniqueHostNamespaceLister struct {
	HostnameLock sync.RWMutex
	// hostname --> Active and InActive Routes
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
	activeRoutes    map[string]metav1.Time
	inactiveRoutes  map[string]metav1.Time
	displacedRoutes map[string]metav1.Time
}

// =======hostname to route mapping===
// == This function should be called in add or update (in update also for newhost name ) for each hostname of the route
func (n *UniqueHostNamespaceLister) UpdateHostnameToRoute(hostName string, route RouteNamspaceName) (bool, []string, []string) {
	n.HostnameLock.Lock()
	defer n.HostnameLock.Unlock()
	var localRouteList RouteList
	var addedRoutes []string
	var deletedRoutes []string

	// fetch existing data first
	utils.AviLog.Debugf("Hostname is: %s", hostName)
	utils.AviLog.Debugf("Current route that needs to be processed is : %v", utils.Stringify(route))
	present, routeList := n.HostnameToRoutesStore.Get(hostName)
	utils.AviLog.Debugf("Details fetched from HostnameToRoutesStore is: Present: %v and routeList: %v", present, utils.Stringify(routeList))
	if !present {
		// host not present, add it to activeRoute list
		// Allocate map
		localRouteList.activeRoutes = make(map[string]metav1.Time)
		localRouteList.inactiveRoutes = make(map[string]metav1.Time)
		localRouteList.displacedRoutes = make(map[string]metav1.Time)

		localRouteList.activeRoutes[route.RouteNSRouteName] = route.CreationTime
		n.HostnameToRoutesStore.AddOrUpdate(hostName, localRouteList)
		utils.AviLog.Debugf("Active routes after appending: %v", utils.Stringify(localRouteList.activeRoutes))
		// TODO: Check what to return in different condition
		return true, nil, nil
	}
	// Compare 0th index active route namespace/name with parameter route and then decide
	// To place parameter route in active list or inactive list based on creation timestamp and namespace
	// 1. Compare namespace of parameter route with 0th index active route
	// 2.1 If namespace is same, then add it to active route list
	// 2.2 else namespace is different
	// 2.2.1 and if activeRoute list contains route, then
	// add that route to inactive list
	// 2.2.2 else fetch route with shortest timestamp and
	// then from iactive list, add all routes with same  namespace as that of shortest timestamp route to active list.
	// return (isrouteaddedToactiveList, activeRoute, displacedRoute)
	existingRoutes := routeList.(RouteList)
	// Check: route is presnt in active list or not
	if _, found := existingRoutes.activeRoutes[route.RouteNSRouteName]; found {
		utils.AviLog.Debugf("key %s found in active routes.", route.RouteNSRouteName)
		// ingest only that route
		return true, nil, nil
	}
	// check: Route is present in inactive list or not
	if _, found := existingRoutes.inactiveRoutes[route.RouteNSRouteName]; found {
		utils.AviLog.Infof("key %v found in inactive list... not processing route.", route.RouteNSRouteName)
		// Do not ingest
		return false, nil, nil
	}
	// Key: not found in active and inactive list, now compare namespace present in activeRoutes
	sameNamespace := false
	displaceRoutes := true
	currentRouteNamespace, _ := ExtractTypeNameNamespace(route.RouteNSRouteName)
	// TODO: Debug statements can be reduced
	utils.AviLog.Debugf("Namespace of current route is: %s", currentRouteNamespace)

	// go through each active route and compare namespace with active routes.
	// Active route should contains routes from same namespace
	// Need to check in case of delete of hostname, are we deleting all these list and entry from internal datastore.
	for nsName, creationTime := range existingRoutes.activeRoutes {
		utils.AviLog.Debugf("NamespaceName from active route list is: %v and creation timestamp: %v", nsName, creationTime)
		existingActiveRoutesNamespace, _ := ExtractTypeNameNamespace(nsName)
		utils.AviLog.Debugf("Extracted out namespace is: %s", existingActiveRoutesNamespace)
		if currentRouteNamespace == existingActiveRoutesNamespace {
			// add it to inactive list
			sameNamespace = true
			break
		} else {
			// add it to inactive list
			displaceRoutes = false
			break
		}
	}
	// With new logic, we need to check length of active route list. if it is empty then, we need to take another namespace
	// from inactive list and fetch all routes of it with given fqdn and ingest it.

	if sameNamespace {
		utils.AviLog.Debugf("Namespace is same. Attaching Route %v to active route list", route.RouteNSRouteName)
		existingRoutes.activeRoutes[route.RouteNSRouteName] = route.CreationTime
		n.HostnameToRoutesStore.AddOrUpdate(hostName, existingRoutes)
		// DO not pass anything. Just return true so to process given route
		return true, nil, nil
	} else {
		utils.AviLog.Debugf("Route namespace is different. Displaced route value: %v and len of existing active routes: %v", displaceRoutes, len(existingRoutes.activeRoutes))
		if displaceRoutes || len(existingRoutes.activeRoutes) == 0 {
			// move active route to displaced routes and then append it to inactive routes also.
			// displaced routes will be used to delete the Route-VS from controller
			existingRoutes.displacedRoutes = existingRoutes.activeRoutes
			utils.AviLog.Debugf("Active routes are copied to displaced route. Now displaced routes are: %v", utils.Stringify(existingRoutes.displacedRoutes))
			existingRoutes.activeRoutes = nil
			existingRoutes.activeRoutes = make(map[string]metav1.Time)

			// now copy
			existingRoutes.activeRoutes[route.RouteNSRouteName] = route.CreationTime
			addedRoutes = append(addedRoutes, route.RouteNSRouteName)
			for nsName, time := range existingRoutes.inactiveRoutes {
				inactiveRouteNamespace, _ := ExtractTypeNameNamespace(nsName)
				if inactiveRouteNamespace == currentRouteNamespace {
					existingRoutes.activeRoutes[nsName] = time
					addedRoutes = append(addedRoutes, nsName)
				}
				// delete entry from inactive routes
				delete(existingRoutes.inactiveRoutes, nsName)
			}
			utils.AviLog.Debugf("Active routes afer copying from inactive list are : %v", utils.Stringify(existingRoutes.activeRoutes))

			// Now add displaced routes to the inactive routes
			for nsName, time := range existingRoutes.displacedRoutes {
				existingRoutes.inactiveRoutes[nsName] = time
				deletedRoutes = append(deletedRoutes, nsName)
			}
			utils.AviLog.Debugf("Now after performing displaced logic... Active routes are: %v, Inactive Routes are: %v", utils.Stringify(existingRoutes.activeRoutes), utils.Stringify(existingRoutes.inactiveRoutes))
			// here we are returning displaced route (whose models needs to be deleted, active routes: models needs to be created )
			n.HostnameToRoutesStore.AddOrUpdate(hostName, existingRoutes)
			return true, addedRoutes, deletedRoutes

		} else {
			utils.AviLog.Debugf("Adding route %s to inactive route list", route.RouteNSRouteName)
			existingRoutes.inactiveRoutes[route.RouteNSRouteName] = route.CreationTime
			utils.AviLog.Debugf("Now inactive routes are : %v", utils.Stringify(existingRoutes.inactiveRoutes))
			// do not process the route
			n.HostnameToRoutesStore.AddOrUpdate(hostName, existingRoutes)
			return false, nil, nil
		}
	}

}

func ExtractTypeNameNamespace(nsName string) (string, string) {
	if nsName != "" {
		// changed from Split to SplitN
		spliStrings := strings.SplitN(nsName, "/", 3)
		if len(spliStrings) == 3 {
			return spliStrings[1], spliStrings[2]
		}
		if len(spliStrings) == 2 {
			return spliStrings[0], spliStrings[1]
		}
	}
	return "", ""
}

/*
// Not being used
func (n *UniqueHostNamespaceLister) GetHostnameToRoute(hostName string) (bool, RouteList) {
	n.HostnameLock.RLock()
	defer n.HostnameLock.RUnlock()
	utils.AviLog.Infof("Inside GetHostnameToRoute... hostname is: %v", hostName)
	found, routeList := n.HostnameToRoutesStore.Get(hostName)
	if found {
		utils.AviLog.Infof(Route list %v", routeList.(RouteList))
		return true, routeList.(RouteList)
	}
	utils.AviLog.Infof("returning false.")
	return false, RouteList{}
}

*/

// This has to be called for each hostname of the route
func (n *UniqueHostNamespaceLister) DeleteHostnameToRoute(hostName string, route RouteNamspaceName) ([]string, string) {
	n.HostnameLock.Lock()
	defer n.HostnameLock.Unlock()
	var routesTobeAdded []string
	present, routeList := n.HostnameToRoutesStore.Get(hostName)
	if !present {
		// hostname entry is not present in the list
		utils.AviLog.Warnf("Returning as host name %s is not present in store", hostName)
		return nil, route.RouteNSRouteName
	}
	existingRoutes := routeList.(RouteList)
	// Check route is present in active list
	if _, flag := existingRoutes.activeRoutes[route.RouteNSRouteName]; flag {
		// present
		utils.AviLog.Debugf("RouteNS to be delete: %v and Active routes are: %v", route.RouteNSRouteName, existingRoutes.activeRoutes)
		delete(existingRoutes.activeRoutes, route.RouteNSRouteName)

		utils.AviLog.Debugf("After deleting Active routes are %v", utils.Stringify(existingRoutes.activeRoutes))
		if len(existingRoutes.activeRoutes) != 0 {
			n.HostnameToRoutesStore.AddOrUpdate(hostName, existingRoutes)
			// TODO: It will good fetch data and print it for debugging.
			//_, temproutes := n.HostnameToRoutesStore.Get(hostName)
			return nil, route.RouteNSRouteName
		}

		// if active route len become zero, then copy routes from inactive list
		// with shortest creation time and same namespace
		anchorTime := metav1.Time{}
		routeNSNameWithShortestCreationTime := ""
		assignedTime := false
		// Logic to fetch route with earlist creation time (o(n))
		for routeNS, rtTime := range existingRoutes.inactiveRoutes {
			if !assignedTime || rtTime.Before(&anchorTime) {
				anchorTime = rtTime
				routeNSNameWithShortestCreationTime = routeNS
				assignedTime = true
			}
		}
		utils.AviLog.Debugf("Namespace with shortest time: %v", routeNSNameWithShortestCreationTime)
		sctNamespace, _ := ExtractTypeNameNamespace(routeNSNameWithShortestCreationTime)
		// nil it and allocate active route again
		existingRoutes.activeRoutes = nil
		existingRoutes.activeRoutes = make(map[string]metav1.Time)
		// now add the routes to active list: (o(n))
		for routeNS, rtTime := range existingRoutes.inactiveRoutes {
			namespace, _ := ExtractTypeNameNamespace(routeNS)
			if namespace == sctNamespace {
				existingRoutes.activeRoutes[routeNS] = rtTime
				routesTobeAdded = append(routesTobeAdded, routeNS)
				delete(existingRoutes.inactiveRoutes, routeNS)
			}
		}
		utils.AviLog.Debugf("After deleting route, existing routes are: %v", utils.Stringify(existingRoutes))
		n.HostnameToRoutesStore.AddOrUpdate(hostName, existingRoutes)
		return routesTobeAdded, route.RouteNSRouteName
	}
	// Check inactive list and delete it from there.
	delete(existingRoutes.inactiveRoutes, route.RouteNSRouteName)
	utils.AviLog.Debugf("After deleting route, existing routes are: %v", utils.Stringify(existingRoutes))

	return nil, route.RouteNSRouteName
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
