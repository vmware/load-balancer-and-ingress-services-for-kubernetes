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
	GWClassName        = "avi-lb"
	GWClassController  = "ako.vmware.com/avi-lb"
)

// below properties are applicable to a configmap object for AKO controller
const (
	ControllerIP           = "controllerIP"
	ControllerVersion      = "controllerVersion"
	CniPlugin              = "cniPlugin"
	EnableEVH              = "enableEVH"
	Layer7Only             = "layer7Only"
	ServicesAPI            = "servicesAPI"
	VipPerNamespace        = "vipPerNamespace"
	ShardVSSize            = "shardVSSize"
	PassthroughShardSize   = "passhtroughShardSize"
	FullSyncFrequency      = "fullSyncFrequency"
	CloudName              = "cloudName"
	ClusterName            = "clusterName"
	EnableRHI              = "enableRHI"
	DefaultDomain          = "defaultDomain"
	DisableStaticRouteSync = "disableStaticRouteSync"
	DefaultIngController   = "defaultIngController"
	VipNetworkList         = "vipNetworkList"
	BgpPeerLabels          = "bgpPeerLabels"
	EnableEvents           = "enableEvents"
	LogLevel               = "logLevel"
	DeleteConfig           = "deleteConfig"
	AutoFQDN               = "autoFQDN"
	ServiceType            = "serviceType"
	NodeKey                = "nodeKey"
	NodeValue              = "nodeValue"
	ServiceEngineGroupName = "serviceEngineGroupName"
	NodeNetworkList        = "nodeNetworkList"
	APIServerPort          = "apiServerPort"
	NSSyncLabelKey         = "nsSyncLabelKey"
	NSSyncLabelValue       = "nsSyncLabelValue"
	TenantName             = "tenantName"
	NoPGForSni             = "noPGForSni"
	NsxtT1LR               = "nsxtT1LR"
	PrimaryInstance        = "primaryInstance"
	IstioEnabled           = "istioEnabled"
	BlockedNamespaceList   = "blockedNamespaceList"
	IPFamily               = "ipFamily"
	EnableMCI              = "enableMCI"
	UseDefaultSecretsOnly  = "useDefaultSecretsOnly"
	DefaultLBController    = "defaultLBController"
	VRFName                = "vrfName"
	EnablePrometheus       = "enablePrometheus"
)

var ConfigMapEnvVars = map[string]string{
	"CTRL_IPADDRESS":             ControllerIP,
	"CTRL_VERSION":               ControllerVersion,
	"CNI_PLUGIN":                 CniPlugin,
	"ENABLE_EVH":                 EnableEVH,
	"SERVICES_API":               ServicesAPI,
	"SHARD_VS_SIZE":              ShardVSSize,
	"PASSTHROUGH_SHARD_SIZE":     PassthroughShardSize,
	"FULL_SYNC_INTERVAL":         FullSyncFrequency,
	"CLOUD_NAME":                 CloudName,
	"CLUSTER_NAME":               ClusterName,
	"ENABLE_RHI":                 EnableRHI,
	"BGP_PEER_LABELS":            BgpPeerLabels,
	"DEFAULT_DOMAIN":             DefaultDomain,
	"DISABLE_STATIC_ROUTE_SYNC":  DisableStaticRouteSync,
	"DEFAULT_ING_CONTROLLER":     DefaultIngController,
	"VIP_NETWORK_LIST":           VipNetworkList,
	"AUTO_L4_FQDN":               AutoFQDN,
	"SERVICE_TYPE":               ServiceType,
	"NODE_KEY":                   NodeKey,
	"NODE_VALUE":                 NodeValue,
	"SEG_NAME":                   ServiceEngineGroupName,
	"NODE_NETWORK_LIST":          NodeNetworkList,
	"AKO_API_PORT":               APIServerPort,
	"TENANT_NAME":                TenantName,
	"NAMESPACE_SYNC_LABEL_KEY":   NSSyncLabelKey,
	"NAMESPACE_SYNC_LABEL_VALUE": NSSyncLabelValue,
	"NSXT_T1_LR":                 NsxtT1LR,
	"PRIMARY_AKO_FLAG":           PrimaryInstance,
	"ISTIO_ENABLED":              IstioEnabled,
	"IP_FAMILY":                  IPFamily,
	"MCI_ENABLED":                EnableMCI,
	"BLOCKED_NS_LIST":            BlockedNamespaceList,
	"VIP_PER_NAMESPACE":          VipPerNamespace,
	"USE_DEFAULT_SECRETS_ONLY":   UseDefaultSecretsOnly,
	"VRF_NAME":                   VRFName,
	"DEFAULT_LB_CONTROLLER":      DefaultLBController,
	"PROMETHEUS_ENABLED":         EnablePrometheus,
}

var ConfigMapEnvVarsGateway = map[string]string{
	"CTRL_IPADDRESS":     ControllerIP,
	"CTRL_VERSION":       ControllerVersion,
	"FULL_SYNC_INTERVAL": FullSyncFrequency,
	"CLOUD_NAME":         CloudName,
	"CLUSTER_NAME":       ClusterName,
	"SEG_NAME":           ServiceEngineGroupName,
	"TENANT_NAME":        TenantName,
	"PRIMARY_AKO_FLAG":   PrimaryInstance,
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

func getGWClassName() types.NamespacedName {
	return types.NamespacedName{
		Name: GWClassName,
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
	// update to the statefulset required?
	if existingSf.Spec.Replicas != nil {
		if newSf.Spec.Replicas != nil && *newSf.Spec.Replicas != *existingSf.Spec.Replicas {
			return true
		}
		if len(existingSf.Spec.Template.Spec.Containers) != len(newSf.Spec.Template.Spec.Containers) {
			return true
		}
		newAKOContainer := newSf.Spec.Template.Spec.Containers[0]
		akoContainer := existingSf.Spec.Template.Spec.Containers[0]
		if newAKOContainer.Image != akoContainer.Image {
			return true
		}
		if newAKOContainer.ImagePullPolicy != akoContainer.ImagePullPolicy {
			return true
		}
		existingEnv := getListOfEnvVars(akoContainer)
		newEnv := getListOfEnvVars(newAKOContainer)
		if !isEnvListEqual(existingEnv, newEnv) {
			return true
		}
		if !reflect.DeepEqual(akoContainer.Ports, newAKOContainer.Ports) {
			return true
		}
		if !reflect.DeepEqual(akoContainer.Resources, newAKOContainer.Resources) {
			return true
		}
		if newAKOContainer.LivenessProbe.HTTPGet.Port != akoContainer.LivenessProbe.HTTPGet.Port {
			return true
		}
		if len(akoContainer.VolumeMounts) != 0 && !reflect.DeepEqual(akoContainer.VolumeMounts, newAKOContainer.VolumeMounts) {
			return true
		}
		if !reflect.DeepEqual(existingSf.Spec.Template.Spec.ImagePullSecrets, newSf.Spec.Template.Spec.ImagePullSecrets) {
			return true
		}

		if len(existingSf.Spec.Template.Spec.Containers) == 2 {
			newGatewayContainer := newSf.Spec.Template.Spec.Containers[1]
			gatewayContainer := existingSf.Spec.Template.Spec.Containers[1]
			if newGatewayContainer.Image != gatewayContainer.Image {
				return true
			}
			if newGatewayContainer.ImagePullPolicy != gatewayContainer.ImagePullPolicy {
				return true
			}
			existingGatewayEnv := getListOfEnvVars(gatewayContainer)
			newGatewayEnv := getListOfEnvVars(newGatewayContainer)
			if !isEnvListEqual(existingGatewayEnv, newGatewayEnv) {
				return true
			}
			if !reflect.DeepEqual(gatewayContainer.Resources, newGatewayContainer.Resources) {
				return true
			}
			if len(gatewayContainer.VolumeMounts) != 0 && !reflect.DeepEqual(gatewayContainer.VolumeMounts, newGatewayContainer.VolumeMounts) {
				return true
			}
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
	envVars = append(envVars, v1.EnvVar{
		Name: "POD_NAMESPACE",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath:  "metadata.namespace",
				APIVersion: "v1",
			},
		},
	})
	return envVars
}

func getEnvVarsForGateway(ako akov1alpha1.AKOConfig) []v1.EnvVar {
	envVars := []v1.EnvVar{}
	for k, v := range ConfigMapEnvVarsGateway {
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
		Value: ako.Spec.AKOGatewayLogFile,
	})
	envVars = append(envVars, v1.EnvVar{
		Name: "POD_NAMESPACE",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath:  "metadata.namespace",
				APIVersion: "v1",
			},
		},
	})
	return envVars
}
