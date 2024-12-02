package status

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"
)

type nplservice struct {
	publisher status.StatusPublisher
}

func (n *nplservice) Update(key string, option status.StatusOptions) {
	n.publisher.UpdateNPLAnnotation(key, option.Namespace, option.ObjName)
}

func (n *nplservice) BulkUpdate(key string, options []status.StatusOptions) {}
func (n *nplservice) Patch(key string, obj runtime.Object, status *status.Status, retryNum ...int) error {
	return nil
}
func (n *nplservice) Delete(key string, option status.StatusOptions) {
	n.publisher.DeleteNPLAnnotation(key, option.Namespace, option.ObjName)
}
