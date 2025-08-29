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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"
	avisession "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	controllerutils "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/utils"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// PKIProfileReconciler reconciles a PKIProfile object
type PKIProfileReconciler struct {
	client.Client
	AviClient     avisession.AviClientInterface
	Scheme        *runtime.Scheme
	Cache         cache.CacheOperation
	Logger        *utils.AviLogger
	EventRecorder record.EventRecorder
	ClusterName   string
}

// GetLogger returns the logger for the reconciler to implement NamespaceHandler interface
func (r *PKIProfileReconciler) GetLogger() *utils.AviLogger {
	return r.Logger
}

type PKIProfileRequest struct {
	Name string `json:"name"`
	akov1alpha1.PKIProfileSpec
	Markers []*models.RoleFilterMatchLabel `json:"markers,omitempty"`
}

// +kubebuilder:rbac:groups=ako.vmware.com,resources=pkiprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ako.vmware.com,resources=pkiprofiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ako.vmware.com,resources=pkiprofiles/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PKIProfile object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *PKIProfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Logger.WithValues("name", req.Name, "namespace", req.Namespace, "traceID", uuid.New().String())
	ctx = utils.LoggerWithContext(ctx, log)
	log.Debug("Reconciling PKIProfile")
	defer log.Debug("Reconciled PKIProfile")
	pki := &akov1alpha1.PKIProfile{}
	err := r.Client.Get(ctx, req.NamespacedName, pki)
	if err != nil {
		if k8serror.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error("Failed to get PKIProfile")
		return ctrl.Result{}, err
	}
	if pki.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(pki, constants.PKIProfileFinalizer) {
			controllerutil.AddFinalizer(pki, constants.PKIProfileFinalizer)
			if err := r.Update(ctx, pki); err != nil {
				log.Error("Failed to add finalizer to PKIProfile")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	} else {
		// The object is being deleted
		removeFinalizer := false
		if err, removeFinalizer = r.DeleteObject(ctx, pki); err != nil {
			return ctrl.Result{}, err
		}
		if removeFinalizer {
			controllerutil.RemoveFinalizer(pki, constants.PKIProfileFinalizer)
		} else {
			if err := r.Status().Update(ctx, pki); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: constants.RequeueInterval}, nil
		}
		if err := r.Update(ctx, pki); err != nil {
			return ctrl.Result{}, err
		}
		r.EventRecorder.Event(pki, corev1.EventTypeNormal, "Deleted", "PKIProfile deleted successfully from Avi Controller")
		log.Info("successfully deleted PKIProfile")
		return ctrl.Result{}, nil
	}
	if err := r.ReconcileIfRequired(ctx, pki); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PKIProfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&akov1alpha1.PKIProfile{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named("pkiprofile").
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(controllerutils.CreateGenericNamespaceHandler(
				r,
				"pkiprofile",
				func() *akov1alpha1.PKIProfileList { return &akov1alpha1.PKIProfileList{} },
				func(list *akov1alpha1.PKIProfileList) []client.Object {
					objects := make([]client.Object, len(list.Items))
					for i, item := range list.Items {
						objects[i] = &item
					}
					return objects
				},
			)),
			builder.WithPredicates(controllerutils.TenantAnnotationNamespacePredicate()),
		).
		Complete(r)
}

// DeleteObject deletes the PKIProfile from Avi Controller and returns (error, bool)
// The boolean indicates whether the finalizer should be removed (true) or kept (false)
func (r *PKIProfileReconciler) DeleteObject(ctx context.Context, pki *akov1alpha1.PKIProfile) (error, bool) {
	log := utils.LoggerFromContext(ctx)
	if pki.Status.UUID != "" && pki.Status.Tenant != "" {
		if err := r.AviClient.AviSessionDelete(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.PKIProfileURL, pki.Status.UUID)), nil, nil, session.SetOptTenant(pki.Status.Tenant)); err != nil {
			// Handle 404 as success case - object doesn't exist, which is the desired state for delete
			if aviError, ok := err.(session.AviError); ok {
				switch aviError.HttpStatusCode {
				case 404:
					log.Info("PKIProfile not found on Avi Controller (404), treating as successful deletion")
					return nil, true
				case 403:
					log.Errorf("PKIProfile is being referred by other objects, cannot be deleted. %s", aviError.Error())
					r.EventRecorder.Event(pki, corev1.EventTypeWarning, "DeletionSkipped", aviError.Error())
					pki.Status.Conditions = controllerutils.SetCondition(pki.Status.Conditions, metav1.Condition{
						Type:               "Deleted",
						Status:             metav1.ConditionFalse,
						LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
						Reason:             "DeletionSkipped",
						Message:            controllerutils.ParseAviErrorMessage(*aviError.Message),
					})
					return nil, false
				}
			}
			log.Errorf("error deleting pki profile: %s", err.Error())
			r.EventRecorder.Event(pki, corev1.EventTypeWarning, "DeletionFailed", fmt.Sprintf("Failed to delete PKIProfile from Avi Controller: %v", err))
			return err, false
		}
	} else {
		r.EventRecorder.Event(pki, corev1.EventTypeWarning, "DeletionSkipped", "UUID not present, PKIProfile may not have been created on Avi Controller")
		log.Warn("error deleting pki profile. uuid not present. possibly avi pki profile object not created")
	}
	return nil, true
}

// TODO: Make this function generic
func (r *PKIProfileReconciler) ReconcileIfRequired(ctx context.Context, pki *akov1alpha1.PKIProfile) error {
	log := utils.LoggerFromContext(ctx)
	namespaceTenant, err := controllerutils.GetTenantInNamespace(ctx, r.Client, pki.Namespace)
	if err != nil {
		log.Errorf("error getting tenant in namespace: %s", err.Error())
		return err
	}

	// Check if tenant in status differs from tenant in namespace annotation
	// Only trigger tenant mismatch if status has a tenant set (not for new resources)
	if pki.Status.Tenant != "" && pki.Status.Tenant != namespaceTenant {
		log.Infof("Tenant update detected. Status tenant: %s, Namespace tenant: %s. Deleting PKIProfile from AVI.", pki.Status.Tenant, namespaceTenant)
		err, _ := r.DeleteObject(ctx, pki)
		if err != nil {
			log.Errorf("Failed to delete PKIProfile due to error: %s", err.Error())
			r.EventRecorder.Event(pki, corev1.EventTypeWarning, "DeletionFailed", fmt.Sprintf("Failed to delete PKIProfile due to error: %v", err))
			return err
		}
		// Clear the status to force recreation with correct tenant
		pki.Status = akov1alpha1.PKIProfileStatus{}
		log.Info("PKIProfile deleted from AVI due to tenant update, status cleared for recreation")
	}

	// Process certificate references and create certificates in AVI controller
	if len(pki.Spec.CACertificateRefs) > 0 {
		log.Infof("Processing %d certificate references in PKIProfile %s/%s",
			len(pki.Spec.CACertificateRefs), pki.Namespace, pki.Name)

		err = r.processCertificateReferences(ctx, pki, namespaceTenant)
		if err != nil {
			log.Errorf("Failed to process certificate references: %v", err)
			r.EventRecorder.Event(pki, corev1.EventTypeWarning, "CertificateProcessingFailed",
				fmt.Sprintf("Failed to process certificate references: %v", err))
			return err
		}
	}

	pkiReq := &PKIProfileRequest{
		Name:           fmt.Sprintf("%s-%s-%s", r.ClusterName, pki.Namespace, pki.Name),
		PKIProfileSpec: pki.Spec,
		Markers:        controllerutils.CreateMarkers(r.ClusterName, pki.Namespace),
	}
	// this is a POST Call
	if pki.Status.UUID == "" {
		resp, err := r.createPKIProfile(ctx, pkiReq, pki, namespaceTenant)
		if err != nil {
			r.EventRecorder.Event(pki, corev1.EventTypeWarning, "CreationFailed", fmt.Sprintf("Failed to create PKIProfile on Avi Controller: %v", err))
			log.Errorf("error creating pki profile: %s", err.Error())
			return err
		}
		uuid, err := extractUUID(resp)
		if err != nil {
			r.EventRecorder.Event(pki, corev1.EventTypeWarning, "UUIDExtractionFailed", fmt.Sprintf("Failed to extract UUID: %v", err))
			log.Errorf("error extracting UUID from pki profile: %s", err.Error())
			return err
		}
		pki.Status.UUID = uuid
		r.EventRecorder.Event(pki, corev1.EventTypeNormal, "Created", "PKIProfile created successfully on Avi Controller")
	} else {
		// this is a PUT Call
		// check if no op by checking generation
		if pki.GetGeneration() == pki.Status.ObservedGeneration {
			// if no op from kubernetes side, check if op required from OOB changes by checking lastModified timestamp
			if pki.Status.LastUpdated != nil {
				dataMap, ok := r.Cache.GetObjectByUUID(ctx, pki.Status.UUID)
				if ok {
					if dataMap.GetLastModifiedTimeStamp().Before(pki.Status.LastUpdated.Time) {
						log.Debug("no op for pki profile")
						return nil
					}
				}
			}
			log.Debugf("overwriting pki profile")
		}
		resp := map[string]interface{}{}
		if err := r.AviClient.AviSessionPut(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.PKIProfileURL, pki.Status.UUID)), pkiReq, &resp, session.SetOptTenant(namespaceTenant)); err != nil {
			log.Errorf("error updating pki profile %s", err.Error())
			r.EventRecorder.Event(pki, corev1.EventTypeWarning, "UpdateFailed", fmt.Sprintf("Failed to update PKIProfile on Avi Controller: %v", err))
			return err
		}
		pki.Status.Conditions = controllerutils.SetCondition(pki.Status.Conditions, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
			Reason:             "Updated",
			Message:            "PKIProfile updated successfully on Avi Controller",
		})
		r.EventRecorder.Event(pki, corev1.EventTypeNormal, "Updated", "PKIProfile updated successfully on Avi Controller")
		log.Info("successfully updated pki profile")
	}
	pki.Status.BackendObjectName = pkiReq.Name
	pki.Status.Tenant = namespaceTenant
	lastUpdated := metav1.Time{Time: time.Now().UTC()}
	pki.Status.LastUpdated = &lastUpdated
	pki.Status.ObservedGeneration = pki.Generation
	if err := r.Status().Update(ctx, pki); err != nil {
		r.EventRecorder.Event(pki, corev1.EventTypeWarning, "StatusUpdateFailed", fmt.Sprintf("Failed to update PKIProfile status: %v", err))
		log.Errorf("unable to update pki profile status %s", err.Error())
		return err
	}
	return nil
}

// createPKIProfile will attempt to create a PKI profile, if it already exists, it will return an object which contains the uuid
func (r *PKIProfileReconciler) createPKIProfile(ctx context.Context, pkiReq *PKIProfileRequest, pki *akov1alpha1.PKIProfile, tenant string) (map[string]interface{}, error) {
	log := utils.LoggerFromContext(ctx)
	resp := map[string]interface{}{}
	if err := r.AviClient.AviSessionPost(utils.GetUriEncoded(constants.PKIProfileURL), pkiReq, &resp, session.SetOptTenant(tenant)); err != nil {
		log.Errorf("error posting pki profile: %s", err.Error())
		if aviError, ok := err.(session.AviError); ok {
			if aviError.HttpStatusCode == http.StatusConflict && strings.Contains(aviError.Error(), "already exists") {
				log.Info("pki profile already exists. trying to get uuid")
				err := r.AviClient.AviSessionGet(utils.GetUriEncoded(fmt.Sprintf("%s?name=%s", constants.PKIProfileURL, pkiReq.Name)), &resp)
				if err != nil {
					log.Errorf("error getting pki profile %s", err.Error())
					return nil, err
				}
				uuid, err := extractUUID(resp)
				if err != nil {
					log.Errorf("error extracting UUID from pki profile: %s", err.Error())
					return nil, err
				}
				log.Info("updating pki profile")
				if err := r.AviClient.AviSessionPut(utils.GetUriEncoded(fmt.Sprintf("%s/%s", constants.PKIProfileURL, uuid)), pkiReq, &resp, session.SetOptTenant(tenant)); err != nil {
					log.Errorf("error updating pki profile", err.Error())
					return nil, err
				}
				pki.Status.Conditions = controllerutils.SetCondition(pki.Status.Conditions, metav1.Condition{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
					Reason:             "Updated",
					Message:            "PKIProfile updated successfully on Avi Controller",
				})
				return resp, nil
			}
		}
		return nil, err
	}
	pki.Status.Conditions = controllerutils.SetCondition(pki.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Time{Time: time.Now().UTC()},
		Reason:             "Created",
		Message:            "PKIProfile created successfully on Avi Controller",
	})
	log.Info("PKI profile successfully created")
	return resp, nil
}

// processCertificateReferences processes all certificate references in the PKI profile
func (r *PKIProfileReconciler) processCertificateReferences(ctx context.Context, pki *akov1alpha1.PKIProfile, tenant string) error {
	log := utils.LoggerFromContext(ctx)

	for _, certRef := range pki.Spec.CACertificateRefs {
		namespace := ""
		if certRef.Namespace != nil {
			namespace = string(*certRef.Namespace)
		} else {
			namespace = pki.Namespace
		}

		log.Infof("Processing certificate reference: %s/%s", namespace, certRef.Name)

		cert, err := r.createCertificateInAvi(ctx, certRef, pki, tenant)
		if err != nil {
			log.Errorf("Failed to create certificate for reference %s/%s: %v", namespace, certRef.Name, err)
			r.EventRecorder.Event(pki, corev1.EventTypeWarning, "CertificateCreationFailed",
				fmt.Sprintf("Failed to create certificate %s/%s in AVI controller: %v", namespace, certRef.Name, err))
			return err
		}

		if cert != nil {
			log.Infof("Certificate %s created/found in AVI controller with UUID: %s", *cert.Name, *cert.UUID)
			r.EventRecorder.Event(pki, corev1.EventTypeNormal, "CertificateCreated",
				fmt.Sprintf("Certificate %s created/found in AVI controller", *cert.Name))
		} else {
			log.Debugf("Certificate processing skipped for reference %s/%s (no changes detected)", namespace, certRef.Name)
		}
	}

	return nil
}

// createCertificateInAvi creates a certificate object in AVI controller if it is referenced in PKI profile
func (r *PKIProfileReconciler) createCertificateInAvi(ctx context.Context, certRef v1.ObjectReference, pki *akov1alpha1.PKIProfile, tenant string) (*models.SSLKeyAndCertificate, error) {
	log := utils.LoggerFromContext(ctx)

	// Get the secret containing the certificate
	secret := &corev1.Secret{}
	secretKey := types.NamespacedName{
		Name:      string(certRef.Name),
		Namespace: "",
	}

	// Handle namespace from ObjectReference
	if certRef.Namespace != nil {
		secretKey.Namespace = string(*certRef.Namespace)
	} else {
		// If namespace is not specified in the reference, use the PKI profile's namespace
		secretKey.Namespace = pki.Namespace
	}

	err := r.Client.Get(ctx, secretKey, secret)
	if err != nil {
		if k8serror.IsNotFound(err) {
			log.Warnf("Certificate secret %s/%s not found", secretKey.Namespace, secretKey.Name)
			return nil, fmt.Errorf("certificate secret %s/%s not found", secretKey.Namespace, secretKey.Name)
		}
		log.Errorf("Failed to get certificate secret %s/%s: %v", secretKey.Namespace, secretKey.Name, err)
		return nil, fmt.Errorf("failed to get certificate secret: %w", err)
	}

	// Check if secret needs processing based on observed resource version
	if !r.shouldProcessSecret(ctx, secret) {
		log.Infof("Secret %s/%s has not changed since last processing (observedResourceVersion: %s, currentResourceVersion: %s), skipping certificate creation",
			secret.Namespace, secret.Name,
			secret.Annotations[constants.ObservedResourceVersionAnnotation],
			secret.ResourceVersion)
		return nil, nil
	}

	// Extract certificate and key from secret
	certData, certExists := secret.Data["tls.crt"]
	if !certExists {
		// Try ca.crt for CA certificates
		certData, certExists = secret.Data["ca.crt"]
	}

	if !certExists {
		log.Errorf("Certificate secret %s/%s does not contain tls.crt or ca.crt field", secretKey.Namespace, secretKey.Name)
		return nil, fmt.Errorf("certificate secret does not contain tls.crt or ca.crt field")
	}

	// Create the SSL certificate object for AVI
	certString := string(certData)
	sslCert := &models.SSLCertificate{
		Certificate: &certString,
	}

	// Create the SSL key and certificate object
	certName := fmt.Sprintf("%s-%s-%s-%s-cert", r.ClusterName, pki.Namespace, pki.Name, secretKey.Name)
	certType := "SSL_CERTIFICATE_TYPE_CA"
	certFormat := "SSL_PEM"

	sslKeyAndCert := &models.SSLKeyAndCertificate{
		Name:        &certName,
		Certificate: sslCert,
		Type:        &certType,
		Format:      &certFormat,
	}

	// Add key if present (for full certificates with private keys)
	if keyData, keyExists := secret.Data["tls.key"]; keyExists {
		keyString := string(keyData)
		sslKeyAndCert.Key = &keyString
		certType = "SSL_CERTIFICATE_TYPE_VIRTUALSERVICE"
		sslKeyAndCert.Type = &certType
	}

	// Check if certificate already exists in AVI
	existingCert, err := r.getCertificateByName(ctx, certName)
	if err == nil && existingCert != nil {
		log.Infof("Certificate %s already exists in AVI controller", certName)

		// Update secret annotation even for existing certificates
		if err := r.updateSecretObservedResourceVersion(ctx, secret); err != nil {
			log.Warnf("Failed to update secret annotation for %s/%s: %v", secret.Namespace, secret.Name, err)
		}

		return existingCert, nil
	}

	// Create certificate in AVI controller using REST API
	url := "api/sslkeyandcertificate"
	var response models.SSLKeyAndCertificate

	err = r.AviClient.AviSessionPost(url, sslKeyAndCert, &response, session.SetOptTenant(tenant))
	if err != nil {
		log.Errorf("Failed to create certificate %s in AVI controller: %v", certName, err)
		return nil, fmt.Errorf("failed to create certificate in AVI controller: %w", err)
	}

	log.Infof("Successfully created certificate %s in AVI controller with UUID: %s", certName, *response.UUID)

	// Update secret annotation with observed resource version
	if err := r.updateSecretObservedResourceVersion(ctx, secret); err != nil {
		log.Warnf("Failed to update secret annotation for %s/%s: %v", secret.Namespace, secret.Name, err)
		// Don't fail the entire operation for annotation update failure
	}

	return &response, nil
}

// getCertificateByName retrieves a certificate from AVI controller by name
func (r *PKIProfileReconciler) getCertificateByName(ctx context.Context, name string) (*models.SSLKeyAndCertificate, error) {
	log := utils.LoggerFromContext(ctx)

	url := fmt.Sprintf("api/sslkeyandcertificate?name=%s", name)

	// Use the generic collection result to handle the response
	result, err := r.AviClient.AviSessionGetCollectionRaw(url)
	if err != nil {
		log.Debugf("Certificate %s not found in AVI controller: %v", name, err)
		return nil, err
	}

	if result.Count > 0 && len(result.Results) > 0 {
		// Parse the JSON results
		var certificates []map[string]interface{}
		if err := json.Unmarshal(result.Results, &certificates); err != nil {
			log.Errorf("Failed to unmarshal certificate results: %v", err)
			return nil, fmt.Errorf("failed to parse certificate results: %w", err)
		}

		if len(certificates) > 0 {
			certData := certificates[0]
			// Create a basic certificate object with just the UUID and name
			cert := &models.SSLKeyAndCertificate{}
			if uuid, exists := certData["uuid"]; exists {
				if uuidStr, ok := uuid.(string); ok {
					cert.UUID = &uuidStr
				}
			}
			if nameField, exists := certData["name"]; exists {
				if nameStr, ok := nameField.(string); ok {
					cert.Name = &nameStr
				}
			}
			return cert, nil
		}
	}

	return nil, fmt.Errorf("certificate %s not found", name)
}

// updateSecretObservedResourceVersion updates the secret annotation with the observed resource version
func (r *PKIProfileReconciler) updateSecretObservedResourceVersion(ctx context.Context, secret *corev1.Secret) error {
	log := utils.LoggerFromContext(ctx)

	// Initialize annotations map if it doesn't exist
	if secret.Annotations == nil {
		secret.Annotations = make(map[string]string)
	}

	// Set the observed resource version annotation
	secret.Annotations[constants.ObservedResourceVersionAnnotation] = secret.ResourceVersion

	// Update the secret
	if err := r.Client.Update(ctx, secret); err != nil {
		log.Errorf("Failed to update secret %s/%s with observed resource version: %v", secret.Namespace, secret.Name, err)
		return fmt.Errorf("failed to update secret annotation: %w", err)
	}

	log.Infof("Updated secret %s/%s with observed resource version: %s", secret.Namespace, secret.Name, secret.ResourceVersion)
	return nil
}

// shouldProcessSecret determines if a secret needs processing based on observed resource version
func (r *PKIProfileReconciler) shouldProcessSecret(ctx context.Context, secret *corev1.Secret) bool {
	log := utils.LoggerFromContext(ctx)

	// If no annotations exist, this is the first time processing
	if secret.Annotations == nil {
		log.Debugf("Secret %s/%s has no annotations, needs processing", secret.Namespace, secret.Name)
		return true
	}

	// Get the observed resource version from annotations
	observedResourceVersion, exists := secret.Annotations[constants.ObservedResourceVersionAnnotation]
	if !exists {
		log.Debugf("Secret %s/%s has no observed resource version annotation, needs processing", secret.Namespace, secret.Name)
		return true
	}

	// Compare observed resource version with current resource version
	if observedResourceVersion != secret.ResourceVersion {
		log.Debugf("Secret %s/%s resource version changed (observed: %s, current: %s), needs processing",
			secret.Namespace, secret.Name, observedResourceVersion, secret.ResourceVersion)
		return true
	}

	// Resource versions match, no processing needed
	log.Debugf("Secret %s/%s resource version unchanged (observed: %s, current: %s), skipping processing",
		secret.Namespace, secret.Name, observedResourceVersion, secret.ResourceVersion)
	return false
}
