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

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	istioInformer "istio.io/client-go/pkg/informers/externalversions/networking/v1alpha3"

	akocrd "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/clientset/versioned"
	akoinformer "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/client/v1alpha1/informers/externalversions/ako/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

var aviInfraSettingEnabled bool
var hostRuleEnabled bool
var httpRuleEnabled bool

func SetCRDEnabledParams(cs akocrd.Interface) {
	timeout := int64(120)
	_, aviInfraError := cs.AkoV1alpha1().AviInfraSettings().List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &timeout})
	if aviInfraError != nil {
		utils.AviLog.Infof("ako.vmware.com/v1alpha1/AviInfraSetting not found/enabled on cluster: %v", aviInfraError)
		aviInfraSettingEnabled = false
	} else {
		utils.AviLog.Infof("ako.vmware.com/v1alpha1/AviInfraSetting enabled on cluster")
		aviInfraSettingEnabled = true
	}

	_, hostRulesError := cs.AkoV1alpha1().HostRules(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &timeout})
	if hostRulesError != nil {
		utils.AviLog.Infof("ako.vmware.com/v1alpha1/HostRule not found/enabled on cluster: %v", hostRulesError)
		hostRuleEnabled = false
	} else {
		utils.AviLog.Infof("ako.vmware.com/v1alpha1/HostRule enabled on cluster")
		hostRuleEnabled = true
	}

	_, httpRulesError := cs.AkoV1alpha1().HTTPRules(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{TimeoutSeconds: &timeout})
	if httpRulesError != nil {
		utils.AviLog.Infof("ako.vmware.com/v1alpha1/HTTPRule not found/enabled on cluster: %v", httpRulesError)
		httpRuleEnabled = false
	} else {
		utils.AviLog.Infof("ako.vmware.com/v1alpha1/HTTPRule enabled on cluster")
		httpRuleEnabled = true
	}
}

func GetAviInfraSettingEnabled() bool {
	return aviInfraSettingEnabled
}

func GetHostRuleEnabled() bool {
	return hostRuleEnabled
}

func GetHttpRuleEnabled() bool {
	return httpRuleEnabled
}

var CRDClientset akocrd.Interface

func SetCRDClientset(cs akocrd.Interface) {
	CRDClientset = cs
	SetCRDEnabledParams(cs)
}

func GetCRDClientset() akocrd.Interface {
	return CRDClientset
}

var CRDInformers *AKOCrdInformers
var IstioInformers *IstioCRDInformers

type AKOCrdInformers struct {
	HostRuleInformer        akoinformer.HostRuleInformer
	HTTPRuleInformer        akoinformer.HTTPRuleInformer
	AviInfraSettingInformer akoinformer.AviInfraSettingInformer
}

func SetCRDInformers(c *AKOCrdInformers) {
	CRDInformers = c
}

func GetCRDInformers() *AKOCrdInformers {
	return CRDInformers
}

type IstioCRDInformers struct {
	VirtualServiceInformer  istioInformer.VirtualServiceInformer
	DestinationRuleInformer istioInformer.DestinationRuleInformer
	GatewayInformer         istioInformer.GatewayInformer
}

func SetIstioCRDInformers(c *IstioCRDInformers) {
	IstioInformers = c
}

func GetIstioCRDInformers() *IstioCRDInformers {
	return IstioInformers
}
