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

type VCenterConfiguration struct {
	UserName   string `json:"user"`
	Password   string `json:"password"`
	VCenterURL string `json:"vcenter_url"`
}

type Platform struct {
	Type                 string               `json:"type"`
	VCenterConfiguration VCenterConfiguration `json:"vcenter_configuration"`
}

type Nodes struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	UserName string `json:"username"`
	Password string `json:"password"`
}

type Cluster struct {
	ClusterID              string   `json:"cluster_id"`
	ClusterName            string   `json:"cluster_name"`
	KubeConfigFilePath     string   `json:"kubeconfig_file"`
	CniPlugin              string   `json:"cniPlugin"`
	CloudName              string   `json:"cloudName"`
	DisableStaticRouteSync string   `json:"disableStaticRouteSync"`
	DefaultIngController   string   `json:"defaultIngController"`
	NetworkName            string   `json:"NetworkName"`
	VRFRefName             string   `json:"vrfRefName"`
	Platform               Platform `json:"platform"`
	KubeNodes              []Nodes  `json:"kubeNodes"`
}

type AkoParams struct {
	NumClusters int       `json:"num_clusters"`
	Clusters    []Cluster `json:"clusters"`
}

type TestParams struct {
	AkoPodName        string `json:"akoPodName"`
	Namespace         string `json:"namespace"`
	AppName           string `json:"appName"`
	ServiceNamePrefix string `json:"serviceNamePrefix"`
	IngressNamePrefix string `json:"ingressNamePrefix"`
	DnsVSUUID         string `json:"dnsVSUUID"`
}

type Networks struct {
	Mgmt string `json:"mgmt"`
}

type Vm struct {
	Datacenter string   `json:"datacenter"`
	Name       string   `json:"name"`
	Cluster    string   `json:"cluster"`
	ClusterIP  string   `json:"cluster_ip"`
	IP         string   `json:"ip"`
	Mask       string   `json:"mask"`
	Networks   Networks `json:"networks"`
	CloudName  string   `json:"cloud_name"`
	Host       string   `json:"host"`
	Static     string   `json:"static"`
	Datastore  string   `json:"datastore"`
	Type       string   `json:"type"`
	UserName   string   `json:"username"`
	Password   string   `json:"password"`
	Gateway    string   `json:"gateway"`
}

type TestbedFields struct {
	AkoParam   AkoParams  `json:"Ako_params"`
	TestParams TestParams `json:"TestParams"`
	Vm         []Vm       `json:"Vm"`
}
