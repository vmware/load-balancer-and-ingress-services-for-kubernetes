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

package utils

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/util/workqueue"
)

var queuewrapper sync.Once
var queueInstance *WorkQueueWrapper
var fixedQueues = [...]*WorkerQueue{{NumWorkers: NumWorkersIngestion, WorkqueueName: ObjectIngestionLayer}, {NumWorkers: NumWorkersGraph, WorkqueueName: GraphLayer}}

type WorkQueueWrapper struct {
	// This struct should manage a set of WorkerQueues for the various layers
	queueCollection map[string]*WorkerQueue
}

func (w *WorkQueueWrapper) GetQueueByName(queueName string) *WorkerQueue {
	workqueue, _ := w.queueCollection[queueName]
	return workqueue
}

func SharedWorkQueue(queueParams ...*WorkerQueue) *WorkQueueWrapper {
	queuewrapper.Do(func() {
		queueInstance = &WorkQueueWrapper{}
		queueInstance.queueCollection = make(map[string]*WorkerQueue)
		if len(queueParams) != 0 {
			for _, queue := range queueParams {
				workqueue := NewWorkQueue(queue.NumWorkers, queue.WorkqueueName, queue.SlowSyncTime)
				queueInstance.queueCollection[queue.WorkqueueName] = workqueue
			}
		} else {
			for _, queue := range fixedQueues {
				workqueue := NewWorkQueue(queue.NumWorkers, queue.WorkqueueName)
				queueInstance.queueCollection[queue.WorkqueueName] = workqueue
			}
		}
	})
	return queueInstance
}

// Common utils like processing worker queue, that is common for all objects.
type WorkerQueue struct {
	NumWorkers    uint32
	Workqueue     []workqueue.RateLimitingInterface
	WorkqueueName string
	workerIdMutex sync.Mutex
	workerId      uint32
	SyncFunc      func(interface{}, *sync.WaitGroup) error
	SlowSyncTime  int
}

func NewWorkQueue(num_workers uint32, workerQueueName string, slowSyncTime ...int) *WorkerQueue {
	queue := &WorkerQueue{}
	queue.Workqueue = make([]workqueue.RateLimitingInterface, num_workers)
	queue.workerId = (uint32(1) << num_workers) - 1
	queue.NumWorkers = num_workers
	queue.WorkqueueName = workerQueueName
	if len(slowSyncTime) > 0 {
		queue.SlowSyncTime = slowSyncTime[0]
	}
	for i := uint32(0); i < num_workers; i++ {
		queue.Workqueue[i] = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), fmt.Sprintf("avi-%s", workerQueueName))
	}
	return queue
}

func (c *WorkerQueue) Run(stopCh <-chan struct{}, wg *sync.WaitGroup) error {
	AviLog.Infof("Starting workers to drain the %s layer queues", c.WorkqueueName)
	if c.SyncFunc == nil {
		// This is a bad situation, the sync function is required.
		AviLog.Fatalf("Sync function is not set for workqueue: %s", c.WorkqueueName)
		return nil
	}
	for i := uint32(0); i < c.NumWorkers; i++ {
		wg.Add(1)
		go c.runWorker(wg)
	}
	AviLog.Infof("Started the workers for: %s", c.WorkqueueName)
	return nil
}
func (c *WorkerQueue) StopWorkers(stopCh <-chan struct{}) {
	for i := uint32(0); i < c.NumWorkers; i++ {
		c.Workqueue[i].ShutDown()
	}
	AviLog.Infof("Shutting down the workers for %s", c.WorkqueueName)
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue. Pick a worker_id from worker_id mask
func (c *WorkerQueue) runWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	workerId := uint32(0xffffffff)
	c.workerIdMutex.Lock()
	for i := uint32(0); i < c.NumWorkers; i++ {
		if ((uint32(1) << i) & c.workerId) != 0 {
			workerId = i
			c.workerId = c.workerId & ^(uint32(1) << i)
			break
		}
	}
	c.workerIdMutex.Unlock()
	AviLog.Infof("Worker id %d", workerId)
	for c.processNextWorkItem(workerId, wg) {
	}
	c.workerIdMutex.Lock()
	c.workerId = c.workerId | (uint32(1) << workerId)
	c.workerIdMutex.Unlock()
}

func (c *WorkerQueue) processNextWorkItem(worker_id uint32, wg *sync.WaitGroup) bool {
	if c.SlowSyncTime != 0 {
		timer := time.NewTimer(time.Duration(c.SlowSyncTime) * time.Second)
		<-timer.C
		return c.processBatchedItems(worker_id, wg)
	}
	return c.processSingleWorkItem(worker_id, wg)
}

func (c *WorkerQueue) processSingleWorkItem(worker_id uint32, wg *sync.WaitGroup) bool {
	obj, shutdown := c.Workqueue[worker_id].Get()
	if shutdown {
		return false
	}
	var ev string
	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.Workqueue[worker_id].Done(obj)
		// Run the syncToAvi, passing it the ev resource to be synced.
		err := c.SyncFunc(obj, wg)
		if err != nil {
			AviLog.Errorf("There was an error while syncing the key: %s", ev)
		}
		c.Workqueue[worker_id].Forget(obj)

		return nil
	}(obj)
	if err != nil {
		runtime.HandleError(err)
		return false
	}
	return true
}

func (c *WorkerQueue) processBatchedItems(worker_id uint32, wg *sync.WaitGroup) bool {
	length := c.Workqueue[worker_id].Len()
	var overallStatus bool
	for i := 0; i < length; i++ {
		overallStatus = c.processSingleWorkItem(worker_id, wg)
		// Break if there's a problem in processing.
		if !overallStatus {
			return overallStatus
		}
	}
	return true
}
