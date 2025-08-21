/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	networkingv1alpha1 "sigs.k8s.io/service-apis/apis/v1alpha1"

	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
)

var EventScheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(EventScheme)
var ParameterCodec = runtime.NewParameterCodec(EventScheme)
var localSchemeBuilder = runtime.SchemeBuilder{
	// openshift
	routev1.AddToScheme,

	// AKO CRDs
	akov1beta1.AddToScheme,

	// WCP gateway
	networkingv1alpha1pre1.AddToScheme,

	// svcapi gateway
	networkingv1alpha1.AddToScheme,

	// kubernetes
	corev1.AddToScheme,
	networkingv1.AddToScheme,
	networkingv1beta1.AddToScheme,
	gatewayv1.Install,
}

var AddToScheme = localSchemeBuilder.AddToScheme

func init() {
	v1.AddToGroupVersion(EventScheme, schema.GroupVersion{Version: "v1"})
	utilruntime.Must(AddToScheme(EventScheme))
}

type EventRecorder struct {
	Recorder record.EventRecorder
	Enabled  bool
	Fake     bool
}

func (e *EventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	// EventTypeWarning Events are always broadcasted even if events are disbled.
	if !e.Fake && (e.Enabled || (!e.Enabled && eventtype == corev1.EventTypeWarning)) {
		e.Recorder.Eventf(object, eventtype, reason, messageFmt, args...)
	}
}

func (e *EventRecorder) Event(object runtime.Object, eventtype, reason, messageFmt string) {
	// EventTypeWarning Events are always broadcasted even if events are disbled.
	if !e.Fake && (e.Enabled || (!e.Enabled && eventtype == corev1.EventTypeWarning)) {
		e.Recorder.Event(object, eventtype, reason, messageFmt)
	}
}

func NewEventRecorder(id string, kubeClient kubernetes.Interface, fake bool) *EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(AviLog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(EventScheme, corev1.EventSource{Component: id})
	return &EventRecorder{Recorder: recorder, Fake: fake, Enabled: true}
}
