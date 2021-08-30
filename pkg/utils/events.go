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

package utils

import (
	routev1 "github.com/openshift/api/route/v1"
	networkingv1alpha1pre1 "github.com/vmware-tanzu/service-apis/apis/v1alpha1pre1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	networkingv1alpha1 "sigs.k8s.io/service-apis/apis/v1alpha1"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
)

var EventScheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(EventScheme)
var ParameterCodec = runtime.NewParameterCodec(EventScheme)
var localSchemeBuilder = runtime.SchemeBuilder{
	// openshift
	routev1.AddToScheme,

	// AKO CRDs
	akov1alpha1.AddToScheme,

	// WCP gateway
	networkingv1alpha1pre1.AddToScheme,

	// svcapi gateway
	networkingv1alpha1.AddToScheme,

	// kubernetes
	corev1.AddToScheme,
	networkingv1.AddToScheme,
	networkingv1.AddToScheme,
}

var AddToScheme = localSchemeBuilder.AddToScheme

func init() {
	v1.AddToGroupVersion(EventScheme, schema.GroupVersion{Version: "v1"})
	utilruntime.Must(AddToScheme(EventScheme))
}

// Event is a wrapper for broadcasting Events from AKO, for AKO managed objects.
// eventType can be one of EventTypeNormal, EventTypeWarning (from corev1).
func Event(object runtime.Object, eventType, reason, message string, formatArgs ...string) {
	// TODO: Get recorder and call the event API like so.
	// TODO: Check if this is blocking.
	// recorder.Event(svc, corev1.EventTypeWarning, "Invalid", "service not synced to Avi")
}
