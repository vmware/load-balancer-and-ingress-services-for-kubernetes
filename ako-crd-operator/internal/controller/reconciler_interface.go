/*
Copyright 2019-2025 VMware, Inc.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	avisession "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
)

// AviClientReconciler defines the interface that all reconcilers must implement
// to receive AVI client updates from the Secret Controller
type AviClientReconciler interface {
	// UpdateAviClient is called by the Secret Controller when AVI credentials are updated
	// and a new session needs to be established
	UpdateAviClient(client avisession.AviClientInterface) error

	// GetReconcilerName returns a unique identifier for this reconciler
	GetReconcilerName() string
}
