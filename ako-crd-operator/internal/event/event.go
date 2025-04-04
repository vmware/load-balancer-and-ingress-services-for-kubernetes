package event

import (
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type EventManager struct {
	eventRecorder *utils.EventRecorder
	podMeta       *v1.Pod
}

func NewEventManager(eventRecorder *utils.EventRecorder, podMeta *v1.Pod) *EventManager {
	return &EventManager{
		eventRecorder: eventRecorder,
		podMeta:       podMeta,
	}
}

func (em *EventManager) PodEventf(eventType, reason, message string, formatArgs ...string) {
	em.eventRecorder.Eventf(em.podMeta, eventType, reason, message, formatArgs)
}

func (em *EventManager) Eventf(runtimeObjectMeta runtime.Object, eventType, reason, message string, formatArgs ...string) {
	em.eventRecorder.Eventf(runtimeObjectMeta, eventType, reason, message, formatArgs)
}
