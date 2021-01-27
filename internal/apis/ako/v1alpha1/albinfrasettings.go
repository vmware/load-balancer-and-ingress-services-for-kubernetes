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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlbInfraSettings is a top-level type
type AlbInfraSettings struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Status AlbInfraSettingsStatus `json:"status,omitempty"`

	Spec AlbInfraSettingsSpec `json:"spec,omitempty"`
}

// AlbInfraSettingsSpec consists of the main AlbInfraSettings settings
type AlbInfraSettingsSpec struct {
	Network  AlbInfraSettingsNetwork  `json:"network,omitempty"`
	SegGroup AlbInfraSettingsSegGroup `json:"segroup,omitempty"`
}

// AlbInfraSettingsVirtualHost defines properties for a host
type AlbInfraSettingsNetwork struct {
	Name string `json:"name,omitempty"`
	Rhi  string `json:"rhi,omitempty"`
}

// AlbInfraSettingsTLS holds secure host specific properties
type AlbInfraSettingsSegGroup struct {
	Name string `json:"name,omitempty"`
}

// AlbInfraSettingsStatus holds the status of the AlbInfraSettings
type AlbInfraSettingsStatus struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AlbInfraSettingsList has the list of AlbInfraSettings objects
type AlbInfraSettingsList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []AlbInfraSettings `json:"items"`
}
