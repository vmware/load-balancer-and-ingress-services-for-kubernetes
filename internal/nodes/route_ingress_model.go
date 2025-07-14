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

package nodes

import (
	"errors"
	"fmt"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	routev1 "github.com/openshift/api/route/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// RouteIngressModel : High Level interfaces that should be implemenetd by
// all l7 route objects, e.g: k8s ingress, openshift route
type RouteIngressModel interface {
	GetName() string
	GetNamespace() string
	GetType() string
	GetSvcLister() *objects.SvcLister
	GetSpec() interface{}
	GetAnnotations() map[string]string
	ParseHostPath() IngressConfig
	Exists() bool
	// this is required due to different naming convention used in ingress where we dont use service name
	// later if we decide to have common naming for ingress and route, then we can hav a common method
	GetDiffPathSvc(map[string][]string, []IngressHostPathSvc, bool) map[string][]string

	GetAviInfraSetting() *akov1beta1.AviInfraSetting
}

// OshiftRouteModel : Model for openshift routes with it's own service lister
type OshiftRouteModel struct {
	key          string
	name         string
	namespace    string
	spec         routev1.RouteSpec
	infrasetting *akov1beta1.AviInfraSetting
	annotations  map[string]string
}

// K8sIngressModel : Model for kubernetes ingresses with default service lister
type K8sIngressModel struct {
	key          string
	name         string
	namespace    string
	spec         networkingv1.IngressSpec
	infrasetting *akov1beta1.AviInfraSetting
	annotations  map[string]string
}

// multiClusterIngressModel : Model for multi-cluster ingresses with default service lister
type multiClusterIngressModel struct {
	key         string
	name        string
	namespace   string
	spec        *akov1alpha1.MultiClusterIngressSpec
	annotations map[string]string
}

func GetOshiftRouteModel(name, namespace, key string) (*OshiftRouteModel, error, bool) {
	routeModel := OshiftRouteModel{
		key:       key,
		name:      name,
		namespace: namespace,
	}
	processObj := true
	processObj = utils.CheckIfNamespaceAccepted(namespace)

	routeObj, err := utils.GetInformers().RouteInformer.Lister().Routes(namespace).Get(name)
	if err != nil {
		return &routeModel, err, processObj
	}
	routeModel.spec = routeObj.Spec
	routeModel.annotations = routeObj.GetAnnotations()
	if !lib.HasValidBackends(routeObj.Spec, name, namespace, key) {
		err := errors.New("validation failed for alternate backends for route: " + name)
		return &routeModel, err, false
	}
	routeModel.infrasetting, err = getL7RouteInfraSetting(key, routeObj.GetAnnotations(), routeObj.GetNamespace())
	return &routeModel, err, processObj
}

func (m *OshiftRouteModel) GetName() string {
	return m.name
}

func (m *OshiftRouteModel) GetNamespace() string {
	return m.namespace
}

func (m *OshiftRouteModel) GetAnnotations() map[string]string {
	return m.annotations
}

func (m *OshiftRouteModel) GetType() string {
	return utils.OshiftRoute
}

func (m *OshiftRouteModel) GetSvcLister() *objects.SvcLister {
	return objects.OshiftRouteSvcLister()
}

func (m *OshiftRouteModel) GetSpec() interface{} {
	return m.spec
}

func (or *OshiftRouteModel) ParseHostPath() IngressConfig {
	o := NewNodesValidator()
	return o.ParseHostPathForRoute(or.namespace, or.name, or.spec, or.key)
}

func (m *OshiftRouteModel) Exists() bool {
	if m.GetSpec() != nil {
		return true
	}
	return false
}

func (m *OshiftRouteModel) GetDiffPathSvc(storedPathSvc map[string][]string, currentPathSvc []IngressHostPathSvc, checkSvc bool) map[string][]string {
	pathSvcCopy := make(map[string][]string)
	for k, v := range storedPathSvc {
		pathSvcCopy[k] = v
	}
	currPathSvcMap := make(map[string][]string)
	for _, val := range currentPathSvc {
		currPathSvcMap[val.Path] = append(currPathSvcMap[val.Path], val.ServiceName)
	}
	for path, services := range currPathSvcMap {
		// for OshiftRouteModel service diff is always checked
		storedServices, ok := pathSvcCopy[path]
		if ok {
			pathSvcCopy[path] = lib.Difference(storedServices, services)
			if len(pathSvcCopy[path]) == 0 {
				delete(pathSvcCopy, path)
			}
		}
	}
	return pathSvcCopy
}

func (m *OshiftRouteModel) GetAviInfraSetting() *akov1beta1.AviInfraSetting {
	return m.infrasetting.DeepCopy()
}

func GetK8sIngressModel(name, namespace, key string) (*K8sIngressModel, error, bool) {
	ingrModel := K8sIngressModel{
		key:       key,
		name:      name,
		namespace: namespace,
	}
	processObj := true
	ingObj, err := utils.GetInformers().IngressInformer.Lister().Ingresses(namespace).Get(name)
	if err != nil {
		return &ingrModel, err, processObj
	}
	if ingObj.GetDeletionTimestamp() != nil {
		return &ingrModel, err, processObj
	}
	processObj = lib.ValidateIngressForClass(key, ingObj) && utils.CheckIfNamespaceAccepted(namespace)
	ingrModel.spec = ingObj.Spec
	ingrModel.annotations = ingObj.GetAnnotations()
	ingrModel.infrasetting, err = getL7IngressInfraSetting(key, utils.String(ingObj.Spec.IngressClassName), namespace)
	return &ingrModel, err, processObj
}

func (m *K8sIngressModel) GetName() string {
	return m.name
}

func (m *K8sIngressModel) GetNamespace() string {
	return m.namespace
}

func (m *K8sIngressModel) GetAnnotations() map[string]string {
	return m.annotations
}

func (m *K8sIngressModel) GetType() string {
	return utils.Ingress
}

func (m *K8sIngressModel) GetSvcLister() *objects.SvcLister {
	return objects.SharedSvcLister()
}

func (m *K8sIngressModel) GetSpec() interface{} {
	return m.spec
}

func (m *K8sIngressModel) ParseHostPath() IngressConfig {
	o := NewNodesValidator()
	return o.ParseHostPathForIngress(m.namespace, m.name, m.spec, m.annotations, m.key)
}

func (m *K8sIngressModel) Exists() bool {
	if m.GetSpec() != nil {
		return true
	}
	return false
}

func (m *K8sIngressModel) GetDiffPathSvc(storedPathSvc map[string][]string, currentPathSvc []IngressHostPathSvc, checkSvc bool) map[string][]string {
	pathSvcCopy := make(map[string][]string)
	for k, v := range storedPathSvc {
		pathSvcCopy[k] = v
	}
	currPathSvcMap := make(map[string][]string)
	for _, val := range currentPathSvc {
		currPathSvcMap[val.Path] = append(currPathSvcMap[val.Path], val.ServiceName)
	}
	for path, services := range currPathSvcMap {
		storedServices, ok := pathSvcCopy[path]
		if ok {
			if checkSvc {
				pathSvcCopy[path] = lib.Difference(storedServices, services)
				if len(pathSvcCopy[path]) == 0 {
					delete(pathSvcCopy, path)
				}
			} else {
				delete(pathSvcCopy, path)
			}
		}
	}
	return pathSvcCopy
}

func (m *K8sIngressModel) GetAviInfraSetting() *akov1beta1.AviInfraSetting {
	return m.infrasetting.DeepCopy()
}

func getL7IngressInfraSetting(key string, ingClassName string, namespace string) (*akov1beta1.AviInfraSetting, error) {
	var infraSetting *akov1beta1.AviInfraSetting

	if ingClassName == "" {
		if defaultIngressClass, found := lib.IsAviLBDefaultIngressClass(); !found {
			//No ingress class is found, return namespace specific infra setting CR
			return lib.GetNamespacedAviInfraSetting(key, namespace, lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer)
		} else {
			ingClassName = defaultIngressClass
		}
	}

	ingClass, err := utils.GetInformers().IngressClassInformer.Lister().Get(ingClassName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Unable to get corresponding IngressClass %s", key, err.Error())
		return nil, err
	} else {
		if ingClass.Spec.Parameters != nil && *ingClass.Spec.Parameters.APIGroup == lib.AkoGroup && ingClass.Spec.Parameters.Kind == lib.AviInfraSetting {
			infraSetting, err = lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().Get(ingClass.Spec.Parameters.Name)
			if err != nil {
				utils.AviLog.Warnf("key: %s, msg: Unable to get corresponding AviInfraSetting via IngressClass %s", key, err.Error())
				return nil, err
			}
			if infraSetting.Status.Status != lib.StatusAccepted {
				utils.AviLog.Warnf("key: %s, msg: Referred AviInfraSetting %s is invalid", key, infraSetting.Name)
				return nil, fmt.Errorf("Referred AviInfraSetting %s is invalid", infraSetting.Name)
			}
			return infraSetting, nil
		}
	}

	return lib.GetNamespacedAviInfraSetting(key, namespace, lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer)
}

func getL7RouteInfraSetting(key string, routeAnnotations map[string]string, namespace string) (*akov1beta1.AviInfraSetting, error) {
	var err error
	var infraSetting *akov1beta1.AviInfraSetting

	if infraSettingAnnotation, ok := routeAnnotations[lib.InfraSettingNameAnnotation]; ok && infraSettingAnnotation != "" {
		infraSetting, err = lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().Get(infraSettingAnnotation)
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: Unable to get corresponding AviInfraSetting via annotation %s", key, err.Error())
			return nil, err
		}
		if infraSetting.Status.Status != lib.StatusAccepted {
			utils.AviLog.Warnf("key: %s, msg: Referred AviInfraSetting %s is invalid", key, infraSetting.Name)
			return nil, fmt.Errorf("Referred AviInfraSetting %s is invalid", infraSetting.Name)
		}
	}

	if infraSetting == nil {
		return lib.GetNamespacedAviInfraSetting(key, namespace, lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer)
	}

	return infraSetting, nil
}

func GetMultiClusterIngressModel(name, namespace, key string) (RouteIngressModel, error, bool) {
	mciModel := &multiClusterIngressModel{
		key:       key,
		name:      name,
		namespace: namespace,
	}
	processObj := utils.CheckIfNamespaceAccepted(namespace)

	ingObj, err := utils.GetInformers().MultiClusterIngressInformer.Lister().MultiClusterIngresses(namespace).Get(name)
	if err != nil {
		return mciModel, err, processObj
	}
	mciModel.spec = &ingObj.Spec
	mciModel.annotations = ingObj.GetAnnotations()
	return mciModel, err, processObj
}

func (mciModel *multiClusterIngressModel) GetName() string {
	return mciModel.name
}

func (mciModel *multiClusterIngressModel) GetNamespace() string {
	return mciModel.namespace
}

func (mciModel *multiClusterIngressModel) GetAnnotations() map[string]string {
	return mciModel.annotations
}

func (mciModel *multiClusterIngressModel) GetType() string {
	return lib.MultiClusterIngress
}

func (mciModel *multiClusterIngressModel) GetSvcLister() *objects.SvcLister {
	return objects.SharedMultiClusterIngressSvcLister()
}

func (mciModel *multiClusterIngressModel) GetSpec() interface{} {
	return mciModel.spec
}

func (mciModel *multiClusterIngressModel) ParseHostPath() IngressConfig {
	o := NewNodesValidator()
	return o.ParseHostPathForMultiClusterIngress(mciModel.namespace, mciModel.name, mciModel.spec, mciModel.key)
}

func (mciModel *multiClusterIngressModel) Exists() bool {
	return mciModel.spec != nil
}

func (mciModel *multiClusterIngressModel) GetDiffPathSvc(storedPathSvc map[string][]string, currentPathSvc []IngressHostPathSvc, checkSvc bool) map[string][]string {
	pathSvcCopy := make(map[string][]string)
	for k, v := range storedPathSvc {
		pathSvcCopy[k] = v
	}
	currPathSvcMap := make(map[string][]string)
	for _, val := range currentPathSvc {
		currPathSvcMap[val.Path] = append(currPathSvcMap[val.Path], val.ServiceName)
	}
	for path, services := range currPathSvcMap {
		storedServices, ok := pathSvcCopy[path]
		if ok {
			if checkSvc {
				pathSvcCopy[path] = lib.Difference(storedServices, services)
				if len(pathSvcCopy[path]) == 0 {
					delete(pathSvcCopy, path)
				}
			} else {
				delete(pathSvcCopy, path)
			}
		}
	}
	return pathSvcCopy
}

func (mciModel *multiClusterIngressModel) GetAviInfraSetting() *akov1beta1.AviInfraSetting {
	enablePublicIP := true
	return &akov1beta1.AviInfraSetting{
		Spec: akov1beta1.AviInfraSettingSpec{
			Network: akov1beta1.AviInfraSettingNetwork{
				EnablePublicIP: &enablePublicIP,
			},
		},
		Status: akov1beta1.AviInfraSettingStatus{
			Status: lib.StatusAccepted,
		},
	}
}
