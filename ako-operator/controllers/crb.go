package controllers

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-operator/api/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func createOrUpdateClusterroleBinding(ctx context.Context, ako akov1alpha1.AKOConfig, log logr.Logger, r *AKOConfigReconciler) error {
	var oldCRB rbacv1.ClusterRoleBinding

	if err := r.Get(ctx, getCRBName(), &oldCRB); err != nil {
		log.V(0).Info("no existing clusterrolebinding with name", "name", CRBName)
	} else {
		if oldCRB.ObjectMeta.GetName() != "" {
			log.V(0).Info("a clusterrolebinding with name already exists, will update", "name",
				oldCRB.ObjectMeta.GetName())
		}
	}

	crb := BuildClusterroleBinding(ako, r, log)
	if oldCRB.ObjectMeta.GetName() != "" {
		if reflect.DeepEqual(oldCRB.Subjects, crb.Subjects) {
			log.Info("no updates required for clusterrolebinding")
			return nil
		}
		err := r.Update(ctx, &crb)
		if err != nil {
			log.Error(err, "unable to update clusterrolebinding", "namespace", crb.ObjectMeta.GetNamespace(),
				"name", crb.ObjectMeta.GetName())
			return err
		}
	} else {
		err := r.Create(ctx, &crb)
		if err != nil {
			log.Error(err, "unable to create clusterrolebinding", "namespace", crb.ObjectMeta.GetNamespace(),
				"name", crb.ObjectMeta.GetName())
			return err
		}
	}
	return nil
}

func BuildClusterroleBinding(ako akov1alpha1.AKOConfig, r *AKOConfigReconciler, log logr.Logger) rbacv1.ClusterRoleBinding {
	crb := rbacv1.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      CRBName,
			Namespace: AviSystemNS,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     AKOCR,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      AKOServiceAccount,
				Namespace: AviSystemNS,
			},
		},
	}

	err := ctrl.SetControllerReference(&ako, &crb, r.Scheme)
	if err != nil {
		log.Error(err, "error in setting controller reference, clusterrolebinding changes would be ignored")
	}
	return crb
}
