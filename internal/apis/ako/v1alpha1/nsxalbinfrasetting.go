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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NsxAlbInfraSetting is a top-level type
type NsxAlbInfraSetting struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Status NsxAlbInfraSettingStatus `json:"status,omitempty"`

	Spec NsxAlbInfraSettingSpec `json:"spec,omitempty"`
}

// NsxAlbInfraSettingSpec consists of the main NsxAlbInfraSetting settings
type NsxAlbInfraSettingSpec struct {
	Network    NsxAlbInfraSettingNetwork  `json:"network,omitempty"`
	SegGroup   NsxAlbInfraSettingSegGroup `json:"seGroup,omitempty"`
	L7Settings NsxAlbInfraL7Settings      `json:"l7Settings,omitempty"`
}

type NsxAlbInfraSettingNetwork struct {
	Name string `json:"name,omitempty"`
	Rhi  bool   `json:"rhi,omitempty"`
}

type NsxAlbInfraSettingSegGroup struct {
	Name string `json:"name,omitempty"`
}

type NsxAlbInfraL7Settings struct {
	ShardSize string `json:"shardSize,omitempty"`
}

// NsxAlbInfraSettingStatus holds the status of the NsxAlbInfraSetting
type NsxAlbInfraSettingStatus struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NsxAlbInfraSettingList has the list of NsxAlbInfraSetting objects
type NsxAlbInfraSettingList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []NsxAlbInfraSetting `json:"items"`
}
