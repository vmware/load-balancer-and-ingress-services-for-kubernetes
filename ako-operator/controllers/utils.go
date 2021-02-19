package controllers

import (
	"reflect"
	"sort"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

// properties used for naming the dependent artifacts
const (
	StatefulSetName    = "ako"
	ServiceAccountName = "ako-sa"
	ServiceName        = "ako"
	ConfigMapName      = "avi-k8s-config"
	AviSystemNS        = "avi-system"
	AKOCR              = "ako-cr"
	CRBName            = "ako-crb"
	AKOServiceAccount  = "ako-sa"
	PSPName            = "ako"
	AviSecretName      = "avi-secret"
)

// below properties are applicable to a configmap object for AKO controller
const (
	ControllerIP           = "controllerIP"
	ControllerVersion      = "controllerVersion"
	CniPlugin              = "cniPlugin"
	ShardVSSize            = "shardVSSize"
	PassthroughShardSize   = "passhtroughShardSize"
	FullSyncFrequency      = "fullSyncFrequency"
	CloudName              = "cloudName"
	ClusterName            = "clusterName"
	EnableRHI              = "enableRHI"
	DefaultDomain          = "defaultDomain"
	DisableStaticRouteSync = "disableStaticRouteSync"
	DefaultIngController   = "defaultIngController"
	SubnetIP               = "subnetIP"
	SubnetPrefix           = "subnetPrefix"
	NetworkName            = "networkName"
	L7ShardingScheme       = "l7ShardingScheme"
	LogLevel               = "logLevel"
	DeleteConfig           = "deleteConfig"
	AdvancedL4             = "advancedL4"
	AutoFQDN               = "autoFQDN"
	SyncNamespace          = "syncNamespace"
	ServiceType            = "serviceType"
	NodeKey                = "nodeKey"
	NodeValue              = "nodeValue"
	ServiceEngineGroupName = "serviceEngineGroupName"
	NodeNetworkList        = "nodeNetworkList"
	APIServerPort          = "apiServerPort"
	NSSyncLabelKey         = "nsSyncLabelKey"
	NSSyncLabelValue       = "nsSyncLabelValue"
	TenantsPerCluster      = "tenantsPerCluster"
	TenantName             = "tenantName"
)

var SecretEnvVars = map[string]string{
	"CTRL_USERNAME": "username",
	"CTRL_PASSWORD": "password",
	"CTRL_CA_DATA":  "certificateAuthorityData",
}

var ConfigMapEnvVars = map[string]string{
	"CTRL_IPADDRESS":             ControllerIP,
	"CTRL_VERSION":               ControllerVersion,
	"CNI_PLUGIN":                 CniPlugin,
	"SHARD_VS_SIZE":              ShardVSSize,
	"PASSTHROUGH_SHARD_SIZE":     PassthroughShardSize,
	"FULL_SYNC_INTERVAL":         FullSyncFrequency,
	"CLOUD_NAME":                 CloudName,
	"CLUSTER_NAME":               ClusterName,
	"ENABLE_RHI":                 EnableRHI,
	"DEFAULT_DOMAIN":             DefaultDomain,
	"DISABLE_STATIC_ROUTE_SYNC":  DisableStaticRouteSync,
	"DEFAULT_ING_CONTROLLER":     DefaultIngController,
	"SUBNET_IP":                  SubnetIP,
	"SUBNET_PREFIX":              SubnetPrefix,
	"NETWORK_NAME":               NetworkName,
	"L7_SHARD_SCHEME":            L7ShardingScheme,
	"ADVANCED_L4":                AdvancedL4,
	"AUTO_FQDN":                  AutoFQDN,
	"SYNC_NAMESPACE":             SyncNamespace,
	"SERVICE_TYPE":               ServiceType,
	"NODE_KEY":                   NodeKey,
	"NODE_VALUE":                 NodeValue,
	"SEG_NAME":                   ServiceEngineGroupName,
	"NODE_NETWORK_LIST":          NodeNetworkList,
	"AKO_API_PORT":               APIServerPort,
	"TENANT_NAME":                TenantName,
	"TENANTS_PER_CLUSTER":        TenantsPerCluster,
	"NAMESPACE_SYNC_LABEL_KEY":   NSSyncLabelKey,
	"NAMESPACE_SYNC_LABEL_VALUE": NSSyncLabelValue,
}

func getSFNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: AviSystemNS,
		Name:      StatefulSetName,
	}
}

func getCRName() types.NamespacedName {
	return types.NamespacedName{
		Name: AKOCR,
	}
}

func getCRBName() types.NamespacedName {
	return types.NamespacedName{
		Name: CRBName,
	}
}
func getSAName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: AviSystemNS,
		Name:      AKOServiceAccount,
	}
}

func getPSPName() types.NamespacedName {
	return types.NamespacedName{
		Name: PSPName,
	}
}

func getConfigMapName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: AviSystemNS,
		Name:      ConfigMapName,
	}
}

func getChecksum(cm v1.ConfigMap, skipList []string) uint32 {
	cmValues := []string{}

	for k, v := range cm.Data {
		if utils.HasElem(skipList, k) {
			continue
		}
		cmValues = append(cmValues, v)
	}
	sort.Strings(cmValues)
	return utils.Hash(utils.Stringify(cmValues))
}

func getListOfEnvVars(container v1.Container) map[string]v1.EnvVar {
	envList := make(map[string]v1.EnvVar)

	for _, env := range container.Env {
		envList[env.Name] = env
	}
	return envList
}

func isEnvListEqual(aEnvList, bEnvList map[string]v1.EnvVar) bool {
	if len(aEnvList) != len(bEnvList) {
		return false
	}
	for k, v := range aEnvList {
		bEnv, ok := bEnvList[k]
		if !ok {
			return false
		}
		if v.String() != bEnv.String() {
			return false
		}
	}
	return true
}

func isSfUpdateRequired(existingSf appsv1.StatefulSet, newSf appsv1.StatefulSet) bool {
	newContainer := newSf.Spec.Template.Spec.Containers[0]

	// update to the statefulset required?
	if existingSf.Spec.Replicas != nil && *existingSf.Spec.Replicas == 1 {
		if len(existingSf.Spec.Template.Spec.Containers) != 1 {
			return true
		}
		akoContainer := existingSf.Spec.Template.Spec.Containers[0]
		if newContainer.Image != akoContainer.Image {
			return true
		}
		if newContainer.ImagePullPolicy != akoContainer.ImagePullPolicy {
			return true
		}
		existingEnv := getListOfEnvVars(akoContainer)
		newEnv := getListOfEnvVars(newContainer)
		if !isEnvListEqual(existingEnv, newEnv) {
			return true
		}
		if !reflect.DeepEqual(akoContainer.Ports, newContainer.Ports) {
			return true
		}
		if !reflect.DeepEqual(akoContainer.Resources, newContainer.Resources) {
			return true
		}
		if newContainer.LivenessProbe.HTTPGet.Port != akoContainer.LivenessProbe.HTTPGet.Port {
			return true
		}
	} else {
		return true
	}
	return false
}

func getEnvVars(ako akov1alpha1.AKOConfig, aviSecret v1.Secret) []v1.EnvVar {

	envVars := []v1.EnvVar{}
	for k, v := range ConfigMapEnvVars {
		if k == "NODE_KEY" || k == "NODE_VALUE" {
			// see if this is present in the ako spec
			if string(ako.Spec.L7Settings.ServiceType) != "NodePort" {
				continue
			}
		}
		envVar := v1.EnvVar{
			Name: k,
			ValueFrom: &v1.EnvVarSource{
				ConfigMapKeyRef: &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: ConfigMapName,
					},
					Key: v,
				},
			},
		}
		envVars = append(envVars, envVar)
	}

	cacertRef, ok := aviSecret.Data["certificateAuthorityData"]
	for k, v := range SecretEnvVars {
		if k == "CTRL_CA_DATA" && (!ok || string(cacertRef) == "") {
			continue
		}
		envVar := v1.EnvVar{
			Name: k,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "avi-secret",
					},
					Key: v,
				},
			},
		}
		envVars = append(envVars, envVar)
	}

	envVars = append(envVars, v1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath:  "metadata.name",
				APIVersion: "v1",
			},
		},
	})

	// add logfile and pvc
	if ako.Spec.PersistentVolumeClaim != "" {
		envVars = append(envVars, v1.EnvVar{
			Name:  "USE_PVC",
			Value: "true",
		})
	}

	envVars = append(envVars, v1.EnvVar{
		Name:  "LOG_FILE_PATH",
		Value: ako.Spec.MountPath,
	})

	envVars = append(envVars, v1.EnvVar{
		Name:  "LOG_FILE_NAME",
		Value: ako.Spec.LogFile,
	})
	return envVars
}
