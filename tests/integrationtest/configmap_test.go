package integrationtest

import (
	"testing"
	"time"

	"github.com/onsi/gomega"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
)

func TestOtherConfigmapDeletion(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	AddConfigMapWithName(KubeClient, "test-configmap")
	time.Sleep(10 * time.Second)
	DeleteConfigMapWithName(KubeClient, t, "test-configmap")
	time.Sleep(5 * time.Second)
	g.Eventually(func() bool {
		return lib.OtherCMDeleteFlag
	}, 30*time.Second).Should(gomega.BeTrue())

}
