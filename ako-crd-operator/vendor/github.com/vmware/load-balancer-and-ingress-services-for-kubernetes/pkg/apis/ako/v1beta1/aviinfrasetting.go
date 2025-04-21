/*
 * Copyright 2020-2021 VMware, Inc.
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

package v1beta1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AviInfraSetting is a top-level type
type AviInfraSetting struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Status AviInfraSettingStatus `json:"status,omitempty"`

	Spec AviInfraSettingSpec `json:"spec,omitempty"`
}

// AviInfraSettingSpec consists of the main AviInfraSetting settings
type AviInfraSettingSpec struct {
	Network     AviInfraSettingNetwork `json:"network,omitempty"`
	SeGroup     AviInfraSettingSeGroup `json:"seGroup,omitempty"`
	L7Settings  AviInfraL7Settings     `json:"l7Settings,omitempty"`
	NSXSettings AviInfraNSXSettings    `json:"nsxSettings,omitempty"`
}

type AviInfraNSXSettings struct {
	Project *string `json:"project,omitempty"`
	T1LR    *string `json:"t1lr,omitempty"`
}

type AviInfraSettingNetwork struct {
	VipNetworks    []AviInfraSettingVipNetwork  `json:"vipNetworks,omitempty"`
	NodeNetworks   []AviInfraSettingNodeNetwork `json:"nodeNetworks,omitempty"`
	EnableRhi      *bool                        `json:"enableRhi,omitempty"`
	EnablePublicIP *bool                        `json:"enablePublicIP,omitempty"`
	BgpPeerLabels  []string                     `json:"bgpPeerLabels,omitempty"`
	Listeners      []AviInfraListeners          `json:"listeners,omitempty"`
}

type AviInfraListeners struct {
	Port        *int  `json:"port,omitempty"`
	EnableSSL   *bool `json:"enableSSL,omitempty"`
	EnableHTTP2 *bool `json:"enableHTTP2,omitempty"`
}
type AviInfraSettingVipNetwork struct {
	NetworkName string `json:"networkName,omitempty"`
	NetworkUUID string `json:"networkUUID,omitempty"`
	Cidr        string `json:"cidr,omitempty"`
	V6Cidr      string `json:"v6cidr,omitempty"`
}

type AviInfraSettingNodeNetwork struct {
	NetworkName string   `json:"networkName,omitempty"`
	NetworkUUID string   `json:"networkUUID,omitempty"`
	Cidrs       []string `json:"cidrs,omitempty"`
}

type AviInfraSettingSeGroup struct {
	Name string `json:"name,omitempty"`
}

type AviInfraL7Settings struct {
	ShardSize string `json:"shardSize,omitempty"`
}

// AviInfraSettingStatus holds the status of the AviInfraSetting
type AviInfraSettingStatus struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AviInfraSettingList has the list of AviInfraSetting objects
type AviInfraSettingList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AviInfraSetting `json:"items"`
}
