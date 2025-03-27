package nodes

import (
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"
)

func (o *AviObjectGraph) BuildHealthMonitorGraph(namespace, name, key string, hm *v1alpha1.HealthMonitor) {
	o.Lock.Lock()
	defer o.Lock.Unlock()

	hmNode := AviHealthMonitor{}
	hmNode.Name = name
	hmNode.Tenant = lib.GetTenant()
	hmNode.HealthMonitorSpec = hm.Spec.DeepCopy()

	o.modelNodes = append(o.modelNodes, &hmNode)
}
