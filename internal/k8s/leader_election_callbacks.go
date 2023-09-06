/*
 * Copyright 2022-2023 VMware, Inc.
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

package k8s

import (
	"fmt"

	v1 "k8s.io/api/core/v1"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/rest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

// Leader election callback functions
func (c *AviController) OnStartedLeading() {
	event, ok := c.NewFSM().Event[fmt.Sprintf("%d-%d", c.State.current, LEADER)]
	if !ok {
		utils.AviLog.Fatalf("Invalid transition. Allowed Current state: %d, Next state: %d", UNKNOWN, LEADER)
		return
	}
	event.Handle()
}

func (c *AviController) OnNewLeader() {
	event, ok := c.NewFSM().Event[fmt.Sprintf("%d-%d", c.State.current, FOLLOWER)]
	if !ok {
		utils.AviLog.Fatalf("Invalid transition. Allowed Current state: %d, Next state: %d", UNKNOWN, FOLLOWER)
		return
	}
	event.Handle()
}

func (c *AviController) OnStoppedLeading() {
	event, ok := c.NewFSM().Event[fmt.Sprintf("%d-%d", c.State.current, FOLLOWER)]
	if !ok {
		utils.AviLog.Fatalf("Invalid transition. Allowed Current state: %d, Next state: %d", LEADER, FOLLOWER)
		return
	}
	event.Handle()
}

// Event Handler functions
func (c *AviController) OnStartedLeadingDuringBootup() {
	c.publishAllParentVSKeysToRestLayer()
	c.CleanupStaleVSes()
	// once the l3 cache is populated, we can call the updatestatus functions from here
	restlayer := rest.NewRestOperations(avicache.SharedAviObjCache())
	restlayer.SyncObjectStatuses()
}

func (c *AviController) OnStartedLeadingAfterFailover() {
	c.publishAllParentVSKeysToRestLayer()
	c.SyncCRDObjects()
	// once the l3 cache is populated, we can call the updatestatus functions from here
	restlayer := rest.NewRestOperations(avicache.SharedAviObjCache())
	restlayer.SyncObjectStatuses()
}

func (c *AviController) OnNewLeaderDuringBootup() {
	c.publishAllParentVSKeysToRestLayer()
	c.CleanupStaleVSes()
}

func (c *AviController) OnLostLeadership() {
	lib.AKOControlConfig().SetIsLeaderFlag(false)
	c.DisableSync = true
	lib.SetDisableSync(true)
	utils.AviLog.Fatal("AKO lost the leadership")
}

// Transition functions
func (c *AviController) TransitionToLeader() {

	lib.AKOControlConfig().SetIsLeaderFlag(true)

	// Transition the state to leader
	c.State.previous = c.State.current
	c.State.current = LEADER

	// inform the user about state transition
	utils.AviLog.Debugf("AKO became a leader")
	lib.AKOControlConfig().PodEventf(v1.EventTypeNormal, "LeaderElection", "AKO became a leader")
}

func (c *AviController) TransitionToFollower() {

	lib.AKOControlConfig().SetIsLeaderFlag(false)

	// Transition the state to follower
	c.State.previous = c.State.current
	c.State.current = FOLLOWER

	// inform the user about state transition
	utils.AviLog.Debugf("AKO became a follower")
	lib.AKOControlConfig().PodEventf(v1.EventTypeNormal, "LeaderElection", "AKO became a follower")
}

func (c *AviController) NoOperation() {}

const (

	// All possible states
	UNKNOWN uint8 = iota
	LEADER
	FOLLOWER

	// Name of the events
	BOOT_UP_AS_LEADER   = "BOOT_UP_AS_LEADER"
	BOOT_UP_AS_FOLLOWER = "BOOT_UP_AS_FOLLOWER"
	FOLLOWER_TO_LEADER  = "FOLLOWER_TO_LEADER"
	LOST_LEADERSHIP     = "LOST_LEADERSHIP"
)

type State struct {
	previous uint8
	current  uint8
}

type FSM struct {
	// key -> current state-next state
	Event map[string]*Event
}

type Event struct {
	Name       string
	Handler    func()
	Transition func()
}

func (c *AviController) NewFSM() *FSM {
	fsm := &FSM{}
	fsm.Event = make(map[string]*Event)

	// Booting up as a leader
	fsm.Event[fmt.Sprintf("%d-%d", UNKNOWN, LEADER)] = &Event{
		Name:       BOOT_UP_AS_LEADER,
		Transition: c.TransitionToLeader,
		Handler:    c.OnStartedLeadingDuringBootup,
	}

	// Booting up as a follower
	fsm.Event[fmt.Sprintf("%d-%d", UNKNOWN, FOLLOWER)] = &Event{
		Name:       BOOT_UP_AS_FOLLOWER,
		Transition: c.TransitionToFollower,
		Handler:    c.OnNewLeaderDuringBootup,
	}

	// failover from follower to leader
	fsm.Event[fmt.Sprintf("%d-%d", FOLLOWER, LEADER)] = &Event{
		Name:       FOLLOWER_TO_LEADER,
		Transition: c.TransitionToLeader,
		Handler:    c.OnStartedLeadingAfterFailover,
	}

	// leader lost leadership
	fsm.Event[fmt.Sprintf("%d-%d", LEADER, FOLLOWER)] = &Event{
		Name:       LOST_LEADERSHIP,
		Transition: c.NoOperation,
		Handler:    c.OnLostLeadership,
	}
	return fsm
}

func (e *Event) Handle() {
	utils.AviLog.Debugf("Triggering the event %s", e.Name)
	e.Transition()
	e.Handler()
	utils.AviLog.Debugf("Finished the event %s", e.Name)
}
