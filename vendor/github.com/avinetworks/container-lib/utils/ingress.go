/*
 * [2013] - [2018] Avi Networks Incorporated
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
	"os"

	extension "k8s.io/api/extensions/v1beta1"
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var IngressApiMap = map[string]string{
	"corev1":      CoreV1IngressInformer,
	"extensionv1": ExtV1IngressInformer,
}

// GetIngressApi reads the INGRESS_API environment variable
func GetIngressApi() string {
	ingressApi := os.Getenv("INGRESS_API")
	ingressApi, ok := IngressApiMap[ingressApi]
	if !ok {
		return CoreV1IngressInformer
	}
	return ingressApi
}

var (
	ExtensionsIngress = schema.GroupVersionResource{
		Group:    "extensions",
		Version:  "v1beta1",
		Resource: "ingresses",
	}

	NetworkingIngress = schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1beta1",
		Resource: "ingresses",
	}
)

func fromExtensions(old *extension.Ingress) (*networking.Ingress, error) {
	networkingIngress := &networking.Ingress{}

	err := runtimeScheme.Convert(old, networkingIngress, nil)
	if err != nil {
		return nil, err
	}

	return networkingIngress, nil
}

func fromNetworking(old *networking.Ingress) (*extension.Ingress, error) {
	extensionsIngress := &extension.Ingress{}

	err := runtimeScheme.Convert(old, extensionsIngress, nil)
	if err != nil {
		return nil, err
	}

	return extensionsIngress, nil
}

// ToNetworkingIngress converts obj interface to networking.Ingress
func ToNetworkingIngress(obj interface{}) (*networking.Ingress, bool) {
	oldVersion, inExtension := obj.(*extension.Ingress)
	if inExtension {
		ing, err := fromExtensions(oldVersion)
		if err != nil {
			AviLog.Warnf("unexpected error converting Ingress from extensions package: %v", err)
			return nil, false
		}

		return ing, true
	}

	if ing, ok := obj.(*networking.Ingress); ok {
		return ing, true
	}

	return nil, false
}

// ToExtensionIngress converts obj interface to extension.Ingress
func ToExtensionIngress(obj interface{}) (*extension.Ingress, bool) {
	oldVersion, inExtension := obj.(*networking.Ingress)
	if inExtension {
		ing, err := fromNetworking(oldVersion)
		if err != nil {
			AviLog.Warnf("unexpected error converting Ingress from networking package: %v", err)
			return nil, false
		}

		return ing, true
	}

	if ing, ok := obj.(*extension.Ingress); ok {
		return ing, true
	}

	return nil, false
}
