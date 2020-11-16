package controllers

import (
	"context"

	"github.com/go-logr/logr"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func createOrUpdateServiceAccount(ctx context.Context, ako akov1alpha1.AKOConfig, log logr.Logger, r *AKOConfigReconciler) error {
	var oldSA v1.ServiceAccount

	if err := r.Get(ctx, getSAName(), &oldSA); err != nil {
		log.V(0).Info("no existing serviceaccount with name", "name", AKOServiceAccount)
	} else {
		if oldSA.ObjectMeta.GetName() != "" {
			log.V(0).Info("a serviceaccount with name already exists, won't update", "name",
				oldSA.ObjectMeta.GetName())
			return nil
		}
	}

	sa := BuildServiceAccount(ako, r, log)
	err := r.Create(ctx, &sa)
	if err != nil {
		log.Error(err, "unable to create serviceaccount", "namespace", sa.ObjectMeta.GetNamespace(),
			"name", sa.ObjectMeta.GetName())
		return err
	}
	return nil
}

// BuildServiceAccount builds a serviceaccount object from the akoconfig resource
func BuildServiceAccount(ako akov1alpha1.AKOConfig, r *AKOConfigReconciler, log logr.Logger) v1.ServiceAccount {
	sa := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AKOServiceAccount,
			Namespace: AviSystemNS,
		},
	}

	err := ctrl.SetControllerReference(&ako, &sa, r.Scheme)
	if err != nil {
		log.Error(err, "error in setting controller reference to serviceaccount, serviceaccount changes will be ignored")
	}
	return sa
}
