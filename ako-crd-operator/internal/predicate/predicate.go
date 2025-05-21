package predicate

import (
	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"
)

func NewCustomHMPredicate() predicate.Predicate {
	return predicate.Funcs{
		// Only allow updates when there is generation change or the last updated time is outdated after T time
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldObj := e.ObjectOld.(*akov1alpha1.HealthMonitor)
			newObj := e.ObjectNew.(*akov1alpha1.HealthMonitor)
			// TODO: make time check configurable
			return oldObj.GetGeneration() != newObj.GetGeneration() || (newObj.Status.LastUpdated != nil &&
				time.Now().Sub(newObj.Status.LastUpdated.Time) >= 5*time.Hour)

		},

		// Allow create events only if LastUpdated is nil
		// this will basically stop reconciliation on pod restart
		CreateFunc: func(e event.CreateEvent) bool {
			obj := e.Object.(*akov1alpha1.HealthMonitor)
			return obj.Status.UUID == ""
		},

		// Allow delete events
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},

		// Allow generic events (e.g., external triggers)
		GenericFunc: func(e event.GenericEvent) bool {
			return true
		},
	}

}
