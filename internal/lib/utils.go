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
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/runtime"
)

type NPLAnnotation struct {
	PodPort  int    `json:"podPort"`
	NodeIP   string `json:"nodeIP"`
	NodePort int    `json:"nodePort"`
}

type PodsWithTargetPort struct {
	Pods       []utils.NamespaceName
	TargetPort int32
}

func ExtractTypeNameNamespace(key string) (string, string, string) {
	segments := strings.SplitN(key, "/", 3)
	if len(segments) == 3 {
		return segments[0], segments[1], segments[2]
	}
	if len(segments) == 2 {
		return segments[0], "", segments[1]
	}
	return "", "", segments[0]
}

func isServiceLBType(svcObj *corev1.Service) bool {
	// If we don't find a service or it is not of type loadbalancer - return false.
	if svcObj.Spec.Type == "LoadBalancer" {
		return true
	}
	return false
}

func IsServiceNodPortType(svcObj *corev1.Service) bool {
	if svcObj.Spec.Type == NodePort {
		return true
	}
	return false
}

func IsServiceClusterIPType(svcObj *corev1.Service) bool {
	if svcObj.Spec.Type == "ClusterIP" {
		return true
	}
	return false
}

func HasSpecLoadBalancerIP(svcObj *corev1.Service) bool {
	if svcObj.Spec.LoadBalancerIP != "" {
		return true
	}
	return false
}

func HasLoadBalancerIPAnnotation(svcObj *corev1.Service) bool {
	if svcObj.Annotations[LoadBalancerIP] != "" {
		return true
	}
	return false
}

func GetSvcKeysForNodeCRUD() (svcl4Keys []string, svcl7Keys []string) {
	// For NodePort if the node matches the  selector update all L4 services.

	svcObjs, err := utils.GetInformers().ServiceInformer.Lister().Services(metav1.NamespaceAll).List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("Unable to retrieve the services : %s", err)
		return
	}
	for _, svc := range svcObjs {
		var key string
		if isServiceLBType(svc) && !GetLayer7Only() {
			label := utils.ObjKey(svc)
			ns := strings.Split(label, "/")
			//Do not append L4 service if namespace is invalid
			if !utils.IsServiceNSValid(ns[0]) {
				continue
			}
			key = utils.L4LBService + "/" + utils.ObjKey(svc)
			svcl4Keys = append(svcl4Keys, key)
		}
		if IsServiceNodPortType(svc) {
			key = utils.Service + "/" + utils.ObjKey(svc)
			svcl7Keys = append(svcl7Keys, key)
		}
	}
	return svcl4Keys, svcl7Keys

}

func GetPodsFromService(namespace, serviceName string, targetPortName intstr.IntOrString) ([]utils.NamespaceName, int32) {
	var pods []utils.NamespaceName
	var targetPort int32
	svcKey := namespace + "/" + serviceName
	svc, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(serviceName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return pods, targetPort
		}
		if found, podsIntf := objects.SharedSvcToPodLister().Get(svcKey); found {
			savedPods, ok := podsIntf.(PodsWithTargetPort)
			if ok {
				return savedPods.Pods, savedPods.TargetPort
			}
		}
		return pods, targetPort
	}

	if len(svc.Spec.Selector) == 0 {
		return pods, targetPort
	}

	podList, err := utils.GetInformers().PodInformer.Lister().Pods(namespace).List(labels.SelectorFromSet(labels.Set(svc.Spec.Selector)))
	if err != nil {
		utils.AviLog.Warnf("Got error while listing Pods with selector %v: %v", svc.Spec.Selector, err)
		return pods, targetPort
	}
	targetPortFound := false
	if targetPortName.Type == intstr.Int {
		targetPortFound = true
		targetPort = int32(targetPortName.IntValue())
	}
	for _, pod := range podList {
		if !targetPortFound {
			for _, pc := range pod.Spec.Containers {
				for _, pp := range pc.Ports {
					if pp.Name == targetPortName.String() {
						targetPort = pp.ContainerPort
					}
				}
			}
		}
		pods = append(pods, utils.NamespaceName{Namespace: pod.Namespace, Name: pod.Name})
	}

	objects.SharedSvcToPodLister().Save(svcKey, PodsWithTargetPort{Pods: pods, TargetPort: targetPort})
	return pods, targetPort
}

func GetServicesForPod(pod *corev1.Pod) ([]string, []string) {
	var svcList, lbList []string
	services, err := utils.GetInformers().ServiceInformer.Lister().List(labels.Everything())
	if err != nil {
		utils.AviLog.Warnf("Got error while listing Services with NPL annotation: %v", err)
		return svcList, lbList
	}

	for _, svc := range services {
		if !matchSvcSelectorPodLabels(svc.Spec.Selector, pod.GetLabels()) {
			continue
		}
		svcKey := svc.Namespace + "/" + svc.Name
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			lbList = append(lbList, svcKey)
		}
		if svc.Spec.Type != corev1.ServiceTypeNodePort {
			svcList = append(svcList, svcKey)
		}
	}
	return svcList, lbList
}

func matchSvcSelectorPodLabels(svcSelector, podLabel map[string]string) bool {
	if len(svcSelector) == 0 {
		return false
	}

	for selectorKey, selectorVal := range svcSelector {
		if labelVal, ok := podLabel[selectorKey]; !ok || selectorVal != labelVal {
			return false
		}
	}
	return true
}

// Difference compares two slices a & b, returns the elements in `a` that aren't in `b`.
func Difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func CheckConstraintsForRFC(name string, pattern string, maxlength int) bool {
	if len(name) > maxlength {
		utils.AviLog.Warnf("Given string %s is longer than expected. Maximum allowed length is: %d", name, maxlength)
		return false
	}

	compliedRegex := regexp.MustCompile(pattern)
	match := compliedRegex.Match([]byte(name))

	return match
}

func CheckRFC1035(name string) bool {
	RFCpattern := "^[a-z]([-a-z0-9]*[a-z0-9])?$"
	maxlength := 63

	if CheckConstraintsForRFC(name, RFCpattern, maxlength) {
		return true
	}

	utils.AviLog.Warnf("Label provided %s does not follow RFC 1035 constraints", name)
	return false
}

func CorrectLabelToSatisfyRFC1035(name *string, prefix string) {

	if CheckRFC1035(prefix + *name) {
		previousLabel := *name
		*name = prefix + *name
		utils.AviLog.Warnf("Label %s has been changed to : %s", previousLabel, *name)
	}

}

func HasSharedVIPAnnotation(svcName, namespace string) bool {
	svcObj, err := utils.GetInformers().ServiceInformer.Lister().Services(namespace).Get(svcName)
	if err != nil {
		return false
	}
	_, ok := svcObj.Annotations[SharedVipSvcLBAnnotation]
	return ok
}

func CheckAndShortenLabelToFollowRFC1035(svcName string, svcNamespace string) (string, string) {
	// Limit the length of the label to 63 to follow RFC1035
	if len(svcName)+len(svcNamespace)+1 > DNS_LABEL_LENGTH {
		availableSpaceForName := DNS_LABEL_LENGTH - len(svcNamespace) - 1
		if availableSpaceForName <= 0 {
			// Length of the namespace is 63, Hence we need to recalculate the
			// space available for name by shortening the namespace length
			availableSpaceForNamespace := DNS_LABEL_LENGTH - len(svcName) - 1
			if availableSpaceForNamespace <= 0 {
				// length of the name is also 63, Hence we will take
				// 48 (75%) characters from namespace
				svcNamespace = svcNamespace[:48]
			} else {
				svcNamespace = svcNamespace[:availableSpaceForNamespace]
			}
			// A label must not end with hyphen.
			for svcNamespace[len(svcNamespace)-1] == '-' {
				svcNamespace = svcNamespace[:len(svcNamespace)-1]
			}
			availableSpaceForName = DNS_LABEL_LENGTH - len(svcNamespace) - 1
		}
		if len(svcName) > availableSpaceForName {
			svcName = svcName[:availableSpaceForName]
		}
	}
	return svcName, svcNamespace
}

func isInfraSettingUpdateRequired(infraSettingCR *akov1beta1.AviInfraSetting, network, t1lr, project string) bool {
	if infraSettingCR.Spec.NSXSettings.T1LR != nil && *infraSettingCR.Spec.NSXSettings.T1LR != t1lr {
		return true
	}
	infraProject := ""
	if infraSettingCR.Spec.NSXSettings.Project != nil {
		infraProject = *infraSettingCR.Spec.NSXSettings.Project
	}
	if project != infraProject {
		return true
	}
	infraNetwork := ""
	if len(infraSettingCR.Spec.Network.VipNetworks) > 0 {
		infraNetwork = infraSettingCR.Spec.Network.VipNetworks[0].NetworkName
	}
	if network != infraNetwork {
		return true
	}
	return false
}

func CreateOrUpdateAviInfraSetting(name, network, t1lr, project string) (*akov1beta1.AviInfraSetting, error) {
	infraSettingCR, err := AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().Get(name)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			utils.AviLog.Errorf("failed to get AviInfraSetting %s, error: %s", name, err.Error())
			return nil, err
		}
		infraSettingCR = nil
	}
	updateRequired := false
	if infraSettingCR != nil {
		updateRequired = isInfraSettingUpdateRequired(infraSettingCR, network, t1lr, project)
		if !updateRequired {
			return infraSettingCR, nil
		}
	}
	infraSettingCR = &akov1beta1.AviInfraSetting{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: akov1beta1.AviInfraSettingSpec{
			L7Settings: akov1beta1.AviInfraL7Settings{
				ShardSize: "DEDICATED",
			},
			NSXSettings: akov1beta1.AviInfraNSXSettings{
				T1LR: &t1lr,
			},
			SeGroup: akov1beta1.AviInfraSettingSeGroup{
				Name: GetClusterID(),
			},
		},
	}
	if network != "" {
		infraSettingCR.Spec.Network = akov1beta1.AviInfraSettingNetwork{
			VipNetworks: []akov1beta1.AviInfraSettingVipNetwork{
				{
					NetworkName: network,
				},
			},
		}
	}
	if project != "" {
		infraSettingCR.Spec.NSXSettings.Project = &project
	}

	if updateRequired {
		return AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Update(context.TODO(), infraSettingCR, metav1.UpdateOptions{})
	}
	return AKOControlConfig().V1beta1CRDClientset().AkoV1beta1().AviInfraSettings().Create(context.TODO(), infraSettingCR, metav1.CreateOptions{})
}

func RemoveInfraSettingAnnotationFromNamespaces(infraSettingCRs map[string]struct{}) error {
	if len(infraSettingCRs) == 0 {
		return nil
	}
	namespaces, err := utils.GetInformers().NSInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Warnf("failed to list all namespaces, error: %s", err.Error())
		return err
	}
	infraSettingToNamespacesMap := make(map[string][]string)
	for _, namespace := range namespaces {
		infraSettingName, ok := namespace.Annotations[InfraSettingNameAnnotation]
		if !ok {
			continue
		}
		infraSettingToNamespacesMap[infraSettingName] = append(infraSettingToNamespacesMap[infraSettingName], namespace.GetName())
	}
	for infraSettinName := range infraSettingCRs {
		namespaces := infraSettingToNamespacesMap[infraSettinName]
		for _, namespace := range namespaces {
			removeInfraSettingAnnotationFromNamespace(namespace, infraSettinName)
		}
	}
	return nil
}

func removeInfraSettingAnnotationFromNamespace(namespace string, infraSettingName ...string) error {
	nsObj, err := utils.GetInformers().NSInformer.Lister().Get(namespace)
	if err != nil {
		utils.AviLog.Warnf("Failed to GET the namespace details, namespace: %s, error :%s", namespace, err.Error())
		return err
	}
	if nsObj.Annotations == nil {
		return nil
	}
	if len(infraSettingName) > 0 && nsObj.Annotations[InfraSettingNameAnnotation] != infraSettingName[0] {
		utils.AviLog.Infof("AviInfraSetting %s is not annotated to the Namespace %s", infraSettingName[0], nsObj.GetName())
		return nil
	}
	delete(nsObj.Annotations, InfraSettingNameAnnotation)
	_, err = utils.GetInformers().ClientSet.CoreV1().Namespaces().Update(context.TODO(), nsObj, metav1.UpdateOptions{})
	if err != nil {
		utils.AviLog.Warnf("Error occurred while Updating namespace: %s", err.Error())
		return err
	}
	utils.AviLog.Infof("Removed AviInfraSetting %s annotation from Namespace %s", infraSettingName[0], namespace)
	return nil
}

func AnnotateNamespaceWithInfraSetting(namespace, infraSettingName string) error {
	nsObj, err := utils.GetInformers().NSInformer.Lister().Get(namespace)
	if err != nil {
		utils.AviLog.Warnf("Failed to GET the namespace details, namespace: %s, error :%s", namespace, err.Error())
		return err
	}
	if nsObj.Annotations == nil {
		nsObj.Annotations = make(map[string]string)
	}
	if nsObj.Annotations[InfraSettingNameAnnotation] == infraSettingName {
		return nil
	}
	nsObj.Annotations[InfraSettingNameAnnotation] = infraSettingName
	_, err = utils.GetInformers().ClientSet.CoreV1().Namespaces().Update(context.TODO(), nsObj, metav1.UpdateOptions{})
	if err != nil {
		utils.AviLog.Warnf("Error occurred while Updating namespace: %s", err.Error())
		return err
	}
	utils.AviLog.Infof("Annotated Namespace %s with AviInfraSetting %s", namespace, infraSettingName)
	return nil
}

func AnnotateSystemNamespaceWithInfraSetting() {
	namespaces, err := utils.GetInformers().ClientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}

	systemNamespaces := make([]corev1.Namespace, 0)
	systemNSVPC := ""
	for _, namespace := range namespaces.Items {
		for key, val := range namespace.GetAnnotations() {
			if key != "nsx.vmware.com/vpc_name" {
				continue
			}
			systemNSVPC = val
			systemNamespaces = append(systemNamespaces, namespace)
		}
	}

	if systemNSVPC == "" {
		return
	}
	arr := strings.Split(systemNSVPC, "/")
	namespace, vpcName := arr[0], arr[1]
	vpcCR, err := GetDynamicClientSet().Resource(VPCGVR).Namespace(namespace).Get(context.TODO(), vpcName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Errorf("failed to get VPC, name: %s, error: %s", vpcName, err.Error())
		return
	}
	status, ok := vpcCR.Object["status"].(map[string]interface{})
	if !ok {
		return
	}
	vpcPath, ok := status["nsxResourcePath"].(string)
	if !ok {
		return
	}
	arr = strings.Split(vpcPath, "/vpcs/")
	infraSettingName := arr[len(arr)-1]
	for _, namespace := range systemNamespaces {
		if val, ok := namespace.Annotations[InfraSettingNameAnnotation]; ok && val == infraSettingName {
			continue
		}
		namespace.Annotations[InfraSettingNameAnnotation] = infraSettingName
		_, err = utils.GetInformers().ClientSet.CoreV1().Namespaces().Update(context.TODO(), &namespace, metav1.UpdateOptions{})
		if err != nil {
			utils.AviLog.Warnf("Error occurred while Updating namespace: %s", err.Error())
			continue
		}
		utils.AviLog.Infof("Annotated Namespace %s with AviInfraSetting %s", namespace.GetName(), infraSettingName)
	}
}

func RunAviInfraSettingInformer(stopCh <-chan struct{}) {
	go AKOControlConfig().CRDInformers().AviInfraSettingInformer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, AKOControlConfig().CRDInformers().AviInfraSettingInformer.Informer().HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
	} else {
		utils.AviLog.Infof("Caches synced")
	}
}

func GetTenantFromInfraSettingForGwSvc(namespace, objName string) string {
	infraSetting := objects.InfraSettingL7Lister().GetGwSvcToInfraSetting(namespace + "/" + objName)
	if infraSetting == "" {
		return GetTenant()
	}
	tenant := objects.InfraSettingL7Lister().GetAviInfraSettingToTenant(infraSetting)
	if tenant == "" {
		tenant = GetTenant()
	}
	return tenant
}

func GetAllTenantsDefinedInAviInfraSettingCRs() (map[string]struct{}, error) {
	tenants := map[string]struct{}{
		GetTenant(): {},
	}
	if !AKOControlConfig().AviInfraSettingEnabled() {
		return tenants, nil
	}
	infraSettingCRs, err := AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().List(labels.Set(nil).AsSelector())
	if err != nil {
		utils.AviLog.Errorf("failed to list AviInfraSetting cr, disabling sync")
		return nil, err
	}

	for _, infraSetting := range infraSettingCRs {
		if infraSetting.Spec.NSXSettings.Project != nil {
			tenants[*infraSetting.Spec.NSXSettings.Project] = struct{}{}
		}
	}
	return tenants, nil
}
