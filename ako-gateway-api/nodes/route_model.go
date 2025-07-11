/*
 * Copyright 2023-2024 VMware, Inc.
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

package nodes

import (
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	akogatewayapilib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/lib"
	akogatewayapistatus "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-gateway-api/status"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
)

type RouteModel interface {
	GetName() string
	GetNamespace() string
	GetType() string
	GetSpec() interface{}
	ParseRouteConfig(key string) *RouteConfig
	Exists() bool
	GetParents() sets.Set[string]
}

func NewRouteModel(key, objType, name, namespace string) (RouteModel, error) {
	switch objType {
	case lib.HTTPRoute:
		return GetHTTPRouteModel(key, name, namespace)
	}
	return nil, fmt.Errorf("object of type %s not supported", objType)
}

type HeaderMatch struct {
	Type  string
	Name  string
	Value string
}

type PathMatch struct {
	Path string
	//Exact, PathPrefix
	Type string
}

type Match struct {
	PathMatch   *PathMatch
	HeaderMatch []*HeaderMatch
}

type Matches []*Match

func (m Matches) Len() int      { return len(m) }
func (m Matches) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m Matches) Less(i, j int) bool {
	if m[i].PathMatch != nil && m[j].PathMatch != nil {
		return m[i].PathMatch.Path < m[j].PathMatch.Path // TODO: need to check this logic
	}
	return false
}

type Header struct {
	Name  string
	Value string
}

type Headers []*Header

func (h Headers) Len() int           { return len(h) }
func (h Headers) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h Headers) Less(i, j int) bool { return h[i].Name < h[j].Name }

type HeaderFilter struct {
	Set    []*Header
	Add    []*Header
	Remove []string
}

type RedirectFilter struct {
	Host       string
	StatusCode int32
}

type HTTPUrlRewriteFilter struct {
	hostname string
	path     *gatewayv1.HTTPPathModifier
}

type ExtensionRefFilter struct {
	Group string
	Kind  string
	Name  string
}
type Filter struct {
	Type             string
	RequestFilter    *HeaderFilter
	ResponseFilter   *HeaderFilter
	RedirectFilter   *RedirectFilter
	UrlRewriteFilter *HTTPUrlRewriteFilter
	ExtensionRef     *ExtensionRefFilter
}

type Backend struct {
	Name      string
	Namespace string
	Port      int32
	Weight    int32
	Kind      string
}

type HTTPBackend struct {
	Backend *Backend
	Filters []*Filter
}

type Rule struct {
	Name     string
	Matches  []*Match
	Filters  []*Filter
	Backends []*HTTPBackend
}

type RouteConfig struct {
	Rules      []*Rule
	Hosts      []string
	IsRejected bool
}

type httpRoute struct {
	key         string
	name        string
	namespace   string
	routeConfig *RouteConfig
	spec        *gatewayv1.HTTPRouteSpec
}

func GetHTTPRouteModel(key string, name, namespace string) (RouteModel, error) {
	hr := &httpRoute{
		key:       key,
		name:      name,
		namespace: namespace,
	}

	hrObj, err := akogatewayapilib.AKOControlConfig().GatewayApiInformers().HTTPRouteInformer.Lister().HTTPRoutes(namespace).Get(name)
	if err != nil {
		return hr, err
	}
	hr.spec = hrObj.Spec.DeepCopy()
	return hr, nil
}

func (hr *httpRoute) GetName() string {
	return hr.name
}

func (hr *httpRoute) GetNamespace() string {
	return hr.namespace
}

func (hr *httpRoute) GetType() string {
	return lib.HTTPRoute
}

func (hr *httpRoute) GetSpec() interface{} {
	return hr.spec
}

func (hr *httpRoute) ParseRouteConfig(key string) *RouteConfig {
	if hr.routeConfig != nil {
		return hr.routeConfig
	}
	routeConfig := &RouteConfig{}

	routeConfig.Hosts = make([]string, len(hr.spec.Hostnames))
	for i := range hr.spec.Hostnames {
		routeConfig.Hosts[i] = string(hr.spec.Hostnames[i])
	}
	var resolvedRefCondition akogatewayapistatus.Condition
	routeConfig.Rules = make([]*Rule, 0, len(hr.spec.Rules))
	for _, rule := range hr.spec.Rules {
		routeConfigRule := &Rule{}
		routeConfigRule.Matches = make([]*Match, 0, len(rule.Matches))
		for _, ruleMatch := range rule.Matches {
			match := &Match{}

			// path match
			if ruleMatch.Path != nil {
				match.PathMatch = &PathMatch{}
				if ruleMatch.Path.Value != nil {
					match.PathMatch.Path = *ruleMatch.Path.Value
				} else {
					match.PathMatch.Path = "/"
				}
				if ruleMatch.Path.Type != nil {
					match.PathMatch.Type = string(*ruleMatch.Path.Type)
				} else {
					match.PathMatch.Type = "PathPrefix"
				}
			}

			// header match
			match.HeaderMatch = make([]*HeaderMatch, 0, len(ruleMatch.Headers))
			for _, header := range ruleMatch.Headers {
				headerMatch := &HeaderMatch{}
				if header.Type != nil {
					headerMatch.Type = string(*header.Type)
				}
				headerMatch.Name = string(header.Name)
				headerMatch.Value = header.Value
				match.HeaderMatch = append(match.HeaderMatch, headerMatch)
			}

			routeConfigRule.Matches = append(routeConfigRule.Matches, match)
		}
		sort.Sort((Matches)(routeConfigRule.Matches))
		if rule.Name != nil {
			routeConfigRule.Name = string(*rule.Name)
		}
		routeConfigRule.Filters = make([]*Filter, 0, len(rule.Filters))
		for _, ruleFilter := range rule.Filters {
			filter := &Filter{}
			filter.Type = string(ruleFilter.Type)

			// request header filter
			if ruleFilter.RequestHeaderModifier != nil {
				filter.RequestFilter = &HeaderFilter{}
				filter.RequestFilter.Add = make([]*Header, 0, len(ruleFilter.RequestHeaderModifier.Add))
				for _, addFilter := range ruleFilter.RequestHeaderModifier.Add {
					addHeader := &Header{
						Name:  string(addFilter.Name),
						Value: addFilter.Value,
					}
					filter.RequestFilter.Add = append(filter.RequestFilter.Add, addHeader)
				}
				filter.RequestFilter.Set = make([]*Header, 0, len(ruleFilter.RequestHeaderModifier.Set))
				for _, setFilter := range ruleFilter.RequestHeaderModifier.Set {
					setHeader := &Header{
						Name:  string(setFilter.Name),
						Value: setFilter.Value,
					}
					filter.RequestFilter.Set = append(filter.RequestFilter.Set, setHeader)
				}
				filter.RequestFilter.Remove = make([]string, len(ruleFilter.RequestHeaderModifier.Remove))
				copy(filter.RequestFilter.Remove, ruleFilter.RequestHeaderModifier.Remove)

				sort.Sort((Headers)(filter.RequestFilter.Add))
				sort.Sort((Headers)(filter.RequestFilter.Set))
				sort.Strings(filter.RequestFilter.Remove)
			}

			// response header filter
			if ruleFilter.ResponseHeaderModifier != nil {
				filter.ResponseFilter = &HeaderFilter{}
				filter.ResponseFilter.Add = make([]*Header, 0, len(ruleFilter.ResponseHeaderModifier.Add))
				for _, addFilter := range ruleFilter.ResponseHeaderModifier.Add {
					addHeader := &Header{
						Name:  string(addFilter.Name),
						Value: addFilter.Value,
					}
					filter.ResponseFilter.Add = append(filter.ResponseFilter.Add, addHeader)
				}
				filter.ResponseFilter.Set = make([]*Header, 0, len(ruleFilter.ResponseHeaderModifier.Set))
				for _, setFilter := range ruleFilter.ResponseHeaderModifier.Set {
					setHeader := &Header{
						Name:  string(setFilter.Name),
						Value: setFilter.Value,
					}
					filter.ResponseFilter.Set = append(filter.ResponseFilter.Set, setHeader)
				}
				filter.ResponseFilter.Remove = make([]string, len(ruleFilter.ResponseHeaderModifier.Remove))
				copy(filter.ResponseFilter.Remove, ruleFilter.ResponseHeaderModifier.Remove)

				sort.Sort((Headers)(filter.ResponseFilter.Add))
				sort.Sort((Headers)(filter.ResponseFilter.Set))
				sort.Strings(filter.ResponseFilter.Remove)
			}

			// request redirect filter
			if ruleFilter.RequestRedirect != nil {
				filter.RedirectFilter = &RedirectFilter{}
				if ruleFilter.RequestRedirect.Hostname != nil {
					filter.RedirectFilter.Host = string(*ruleFilter.RequestRedirect.Hostname)
				}
				if ruleFilter.RequestRedirect.StatusCode != nil {
					filter.RedirectFilter.StatusCode = int32(*ruleFilter.RequestRedirect.StatusCode)
				}
			}

			// URL rewrite filter
			if ruleFilter.URLRewrite != nil {
				filter.UrlRewriteFilter = &HTTPUrlRewriteFilter{}
				if ruleFilter.URLRewrite.Hostname != nil {
					filter.UrlRewriteFilter.hostname = string(*ruleFilter.URLRewrite.Hostname)
				}
				if ruleFilter.URLRewrite.Path != nil {
					filter.UrlRewriteFilter.path = ruleFilter.URLRewrite.Path.DeepCopy()
					if strings.HasPrefix(*filter.UrlRewriteFilter.path.ReplaceFullPath, "/") {
						*filter.UrlRewriteFilter.path.ReplaceFullPath = (*filter.UrlRewriteFilter.path.ReplaceFullPath)[1:]
					}

				}
			}
			// ExtensionRef filters
			if ruleFilter.ExtensionRef != nil {
				filter.ExtensionRef = &ExtensionRefFilter{}
				filter.ExtensionRef.Group = string(ruleFilter.ExtensionRef.Group)
				filter.ExtensionRef.Kind = string(ruleFilter.ExtensionRef.Kind)
				filter.ExtensionRef.Name = string(ruleFilter.ExtensionRef.Name)
			}
			routeConfigRule.Filters = append(routeConfigRule.Filters, filter)
		}
		hasInvalidBackend := false
		for _, ruleBackend := range rule.BackendRefs {
			httpBackend := &HTTPBackend{}
			backend := &Backend{}
			backend.Name = string(ruleBackend.BackendRef.Name)
			if ruleBackend.BackendRef.Namespace != nil {
				backend.Namespace = string(*ruleBackend.BackendRef.Namespace)
			} else {
				backend.Namespace = hr.namespace
			}
			if ruleBackend.BackendRef.Port != nil {
				//Default 0
				backend.Port = int32(*ruleBackend.Port)
			}
			if ruleBackend.BackendRef.Kind != nil {
				backend.Kind = string(*ruleBackend.Kind)
			}
			backend.Weight = 1
			if ruleBackend.Weight != nil {
				backend.Weight = *ruleBackend.Weight
			}
			httpBackend.Backend = backend
			httpBackend.Filters = make([]*Filter, 0, len(ruleBackend.Filters))
			for _, filter := range ruleBackend.Filters {
				// only extension ref filters suported in backend
				if filter.ExtensionRef != nil {
					httpBackendFilter := &Filter{}
					httpBackendFilter.Type = string(filter.Type)
					httpBackendFilter.ExtensionRef = &ExtensionRefFilter{}
					httpBackendFilter.ExtensionRef.Group = string(filter.ExtensionRef.Group)
					httpBackendFilter.ExtensionRef.Kind = string(filter.ExtensionRef.Kind)
					httpBackendFilter.ExtensionRef.Name = string(filter.ExtensionRef.Name)
					httpBackend.Filters = append(httpBackend.Filters, httpBackendFilter)
				}
			}
			isValidBackend, resolvedRefConditionforBackend := validateBackendReference(key, *backend, httpBackend.Filters, hr.namespace)
			if isValidBackend {
				routeConfigRule.Backends = append(routeConfigRule.Backends, httpBackend)
				if !hasInvalidBackend {
					resolvedRefCondition = resolvedRefConditionforBackend
				}
			} else {
				hasInvalidBackend = true
				resolvedRefCondition = resolvedRefConditionforBackend
			}
		}
		routeConfig.Rules = append(routeConfig.Rules, routeConfigRule)
		routeConfig.IsRejected = hasInvalidBackend
	}
	hr.routeConfig = routeConfig
	setResolvedRefConditionInHTTPRouteStatus(key, resolvedRefCondition, lib.HTTPRoute+"/"+hr.GetNamespace()+"/"+hr.GetName())
	return hr.routeConfig
}

func (hr *httpRoute) Exists() bool {
	return hr != nil
}

func (hr *httpRoute) GetParents() sets.Set[string] {
	var parents sets.Set[string]
	for _, ref := range hr.spec.ParentRefs {
		namespace := hr.namespace
		if ref.Namespace != nil {
			namespace = string(*ref.Namespace)
		}
		parents.Insert(namespace + "/" + string(ref.Name))
	}
	return parents
}
