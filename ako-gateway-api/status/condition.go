/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package status

import (
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type Condition interface {
	Type(string) Condition
	Reason(string) Condition
	Status(metav1.ConditionStatus) Condition
	Message(string) Condition
	ObservedGeneration(int64) Condition
	SetIn(*[]metav1.Condition)
}

type condition struct {
	conditionType      string
	reason             string
	status             metav1.ConditionStatus
	message            string
	observedGeneration int64
}

func NewCondition() Condition {
	return &condition{observedGeneration: -1}
}

func (c *condition) Type(t string) Condition {
	c.conditionType = t
	return c
}

func (c *condition) Reason(r string) Condition {
	c.reason = r
	return c
}

func (c *condition) Status(s metav1.ConditionStatus) Condition {
	c.status = s
	return c
}

func (c *condition) Message(m string) Condition {
	c.message = m
	return c
}

func (c *condition) ObservedGeneration(o int64) Condition {
	c.observedGeneration = o
	return c
}

func (c *condition) SetIn(conditions *[]metav1.Condition) {

	if c.conditionType == "" ||
		c.status == "" ||
		c.reason == "" ||
		c.observedGeneration == -1 {
		utils.AviLog.Errorf("condition has empty values, not updating the status %v", c)
		return
	}

	apimeta.SetStatusCondition(
		conditions,
		metav1.Condition{
			Type:               c.conditionType,
			Status:             c.status,
			Reason:             c.reason,
			Message:            c.message,
			ObservedGeneration: c.observedGeneration,
		},
	)
}
