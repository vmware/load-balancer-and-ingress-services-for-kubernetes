/*
Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.

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
	"context"
	"errors"
	"strconv"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1beta1"
)

func createOrUpdateStatefulSet(ctx context.Context, ako akov1beta1.AKOConfig, log logr.Logger, r *AKOConfigReconciler,
	aviSecret corev1.Secret) error {

	desiredSf, err := BuildStatefulSet(ako, aviSecret)
	if err != nil {
		log.Error(err, "error in building statefulset", "name", StatefulSetName)
		return err
	}

	var existingSf appsv1.StatefulSet
	err = r.Get(ctx, getSFNamespacedName(), &existingSf)

	if err != nil {
		if !k8serrors.IsNotFound(err) {
			log.Error(err, "failed to get StatefulSet")
			return err
		}
		// StatefulSet does not exist, create it.
		log.V(0).Info("no pre-existing statefulset, creating one")
		if err := ctrl.SetControllerReference(&ako, &desiredSf, r.Scheme); err != nil {
			log.Error(err, "error in setting controller reference to statefulset")
			return err
		}
		if err := r.Create(ctx, &desiredSf); err != nil {
			log.Error(err, "unable to create statefulset", "name", desiredSf.GetName())
			return err
		}
	} else {
		// StatefulSet exists, update it if necessary.
		if !rebootRequired && !isSfUpdateRequired(existingSf, desiredSf) {
			log.V(0).Info("no updates required for the statefulset")
			// Add to object list to ensure it's tracked for cleanup, then return.
			objList := getObjectList()
			objList[getSFNamespacedName()] = &existingSf
			return nil
		}

		log.V(0).Info("updating existing statefulset", "rebootRequired", rebootRequired)

		// To update an existing object, we must modify the object that was fetched from the
		// cluster, as it contains the required resourceVersion.
		desiredSf.Spec.DeepCopyInto(&existingSf.Spec)

		// Ensure the controller reference is set on the existing object before updating.
		if err := ctrl.SetControllerReference(&ako, &existingSf, r.Scheme); err != nil {
			log.Error(err, "error in setting controller reference to statefulset")
			return err
		}

		if err := r.Update(ctx, &existingSf); err != nil {
			log.Error(err, "unable to update statefulset", "name", existingSf.GetName())
			return err
		}

		if rebootRequired {
			rebootRequired = false
		}
	}

	// After creating or updating, fetch the latest state to update the global list.
	var currentSf appsv1.StatefulSet
	if err := r.Get(ctx, getSFNamespacedName(), &currentSf); err != nil {
		log.Error(err, "failed to get statefulset after create/update")
		return err
	}

	// update this object in the global list
	objList := getObjectList()
	objList[getSFNamespacedName()] = &currentSf
	log.V(0).Info("statefulset created/updated", "resource version", currentSf.GetResourceVersion())
	return nil
}

func getPullPolicy(pullPolicy string) (corev1.PullPolicy, error) {
	typedPullPolicy := corev1.PullPolicy(pullPolicy)
	switch typedPullPolicy {
	case corev1.PullAlways:
		return corev1.PullAlways, nil
	case corev1.PullIfNotPresent:
		return corev1.PullIfNotPresent, nil
	case corev1.PullNever:
		return corev1.PullNever, nil
	default:
		return corev1.PullPolicy(""), errors.New("invalid pull policy")
	}
}

func buildResources(ako akov1beta1.AKOConfig) (corev1.ResourceRequirements, error) {
	var rr corev1.ResourceRequirements

	limitCPU, err := resource.ParseQuantity(ako.Spec.Resources.Limits.CPU)
	if err != nil {
		return rr, err
	}
	limitMemory, err := resource.ParseQuantity(ako.Spec.Resources.Limits.Memory)
	if err != nil {
		return rr, err
	}

	requestedCPU, err := resource.ParseQuantity(ako.Spec.Resources.Requests.CPU)
	if err != nil {
		return rr, err
	}

	requestedMemory, err := resource.ParseQuantity(ako.Spec.Resources.Requests.Memory)
	if err != nil {
		return rr, err
	}

	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu":    limitCPU,
			"memory": limitMemory,
		},
		Requests: corev1.ResourceList{
			"cpu":    requestedCPU,
			"memory": requestedMemory,
		},
	}, nil
}

func BuildStatefulSet(ako akov1beta1.AKOConfig, aviSecret corev1.Secret) (appsv1.StatefulSet, error) {
	sf := appsv1.StatefulSet{}

	sf.ObjectMeta = metav1.ObjectMeta{
		Name:      StatefulSetName,
		Namespace: AviSystemNS,
	}

	image := ako.Spec.ImageRepository
	imagePullPolicy, err := getPullPolicy(ako.Spec.ImagePullPolicy)
	if err != nil {
		return sf, err
	}
	replicas := int32(ako.Spec.ReplicaCount)
	if replicas > 2 {
		return sf, errors.New("ReplicaCount greater than 2 is not supported for AKO StatefulSet")
	}
	sf.Spec.Replicas = &replicas
	sf.Spec.ServiceName = ServiceName
	akoLabels := map[string]string{
		"app":                    "ako",
		"app.kubernetes.io/name": "ako",
	}
	sf.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: akoLabels,
	}

	// build the env vars
	envVars := getEnvVars(ako, aviSecret)

	volumeMounts := []corev1.VolumeMount{}
	volumes := []corev1.Volume{}
	if ako.Spec.PersistentVolumeClaim != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "ako-pv-storage",
			MountPath: ako.Spec.MountPath,
		})
		volumes = append(volumes, corev1.Volume{
			Name: "ako-pv-storage",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: ako.Spec.PersistentVolumeClaim,
				},
			},
		})
	}

	apiServerPort := ako.Spec.AKOSettings.APIServerPort
	if apiServerPort == 0 {
		apiServerPort = 8080
	}

	ports := []corev1.ContainerPort{}

	resources, err := buildResources(ako)
	if err != nil {
		return sf, err
	}
	template := corev1.PodTemplateSpec{}
	template.SetLabels(akoLabels)
	if ako.Spec.AKOSettings.IstioEnabled {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "istio-certs",
			MountPath: "/etc/istio-output-certs/",
		})
		template.Annotations = map[string]string{
			"sidecar.istio.io/inject":                          "true",
			"traffic.sidecar.istio.io/includeInboundPorts":     "",
			"traffic.sidecar.istio.io/includeOutboundIPRanges": "",
			"proxy.istio.io/config": `proxyMetadata:
  OUTPUT_CERTS: /etc/istio-output-certs
`,
			"sidecar.istio.io/userVolume":      `[{"name": "istio-certs", "emptyDir": {"medium":"Memory"}}]`,
			"sidecar.istio.io/userVolumeMount": `[{"name": "istio-certs", "mountPath": "/etc/istio-output-certs"}]`,
		}
	}
	if ako.Spec.FeatureGates.EnablePrometheus {
		ports = append(ports, corev1.ContainerPort{
			Name:          "prometheus-port",
			ContainerPort: int32(apiServerPort),
			Protocol:      corev1.ProtocolTCP,
		})
		if ako.Spec.AKOSettings.IstioEnabled {
			template.Annotations[PrometheusScrapeAnnotation] = "true"
			template.Annotations[PrometheusPortAnnotation] = strconv.Itoa(apiServerPort)
			template.Annotations[PrometheusPathAnnotation] = "/metrics"
		} else {
			template.Annotations = map[string]string{
				PrometheusScrapeAnnotation: "true",
				PrometheusPortAnnotation:   strconv.Itoa(apiServerPort),
				PrometheusPathAnnotation:   "/metrics",
			}
		}
	}
	template.Spec = corev1.PodSpec{
		ServiceAccountName: ServiceAccountName,
		Volumes:            volumes,
		Containers: []corev1.Container{
			{
				Name:            "ako",
				VolumeMounts:    volumeMounts,
				Image:           image,
				ImagePullPolicy: imagePullPolicy,
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{
							Command: []string{"/bin/sh", "/var/pre_stop_hook.sh"},
						},
					},
				},
				Ports:     ports,
				Resources: resources,
				LivenessProbe: &corev1.Probe{
					InitialDelaySeconds: 5,
					PeriodSeconds:       10,
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/api/status",
							Port: intstr.FromInt(apiServerPort),
						},
					},
				},
				Env: envVars,
			},
		},
		Affinity: &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key: "app.kubernetes.io/name", Operator: metav1.LabelSelectorOpIn, Values: []string{"ako"},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}
	if len(ako.Spec.ImagePullSecrets) != 0 {
		var imagePullSecrets []corev1.LocalObjectReference
		for _, secret := range ako.Spec.ImagePullSecrets {
			imagePullSecrets = append(imagePullSecrets, corev1.LocalObjectReference{Name: secret.Name})
		}
		template.Spec.ImagePullSecrets = imagePullSecrets
	}
	if ako.Spec.FeatureGates.GatewayAPI {
		gatewayImagePullPolicy, err := getPullPolicy(ako.Spec.GatewayAPI.Image.PullPolicy)
		if err != nil {
			return sf, err
		}
		envVarsGateway := getEnvVarsForGateway(ako)
		gatewayContainer := corev1.Container{
			Name:            "ako-gateway-api",
			VolumeMounts:    volumeMounts,
			Image:           ako.Spec.GatewayAPI.Image.Repository,
			ImagePullPolicy: gatewayImagePullPolicy,
			Resources:       resources,
			Env:             envVarsGateway,
		}
		template.Spec.Containers = append(template.Spec.Containers, gatewayContainer)
	}
	sf.Spec.Template = template
	return sf, nil
}
