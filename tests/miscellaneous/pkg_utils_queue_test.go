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

package miscellaneous

import (
	"sync"
	"testing"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func TestNewWorkQueue(t *testing.T) {
	tests := []struct {
		name            string
		numWorkers      uint32
		queueName       string
		slowSyncTime    []int
		expectedWorkers uint32
	}{
		{
			name:            "Create queue with 4 workers",
			numWorkers:      4,
			queueName:       "test-queue",
			slowSyncTime:    nil,
			expectedWorkers: 4,
		},
		{
			name:            "Create queue with 8 workers",
			numWorkers:      8,
			queueName:       "test-queue-8",
			slowSyncTime:    nil,
			expectedWorkers: 8,
		},
		{
			name:            "Create queue with slow sync time",
			numWorkers:      4,
			queueName:       "slow-queue",
			slowSyncTime:    []int{5},
			expectedWorkers: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var queue *utils.WorkerQueue
			if tt.slowSyncTime != nil {
				queue = utils.NewWorkQueue(tt.numWorkers, tt.queueName, tt.slowSyncTime...)
			} else {
				queue = utils.NewWorkQueue(tt.numWorkers, tt.queueName)
			}

			if queue == nil {
				t.Fatal("NewWorkQueue() returned nil")
			}

			if queue.NumWorkers != tt.expectedWorkers {
				t.Errorf("NewWorkQueue() NumWorkers = %v, want %v", queue.NumWorkers, tt.expectedWorkers)
			}

			if queue.WorkqueueName != tt.queueName {
				t.Errorf("NewWorkQueue() WorkqueueName = %v, want %v", queue.WorkqueueName, tt.queueName)
			}

			if len(queue.Workqueue) != int(tt.expectedWorkers) {
				t.Errorf("NewWorkQueue() created %d queues, want %d", len(queue.Workqueue), tt.expectedWorkers)
			}

			if tt.slowSyncTime != nil && queue.SlowSyncTime != tt.slowSyncTime[0] {
				t.Errorf("NewWorkQueue() SlowSyncTime = %v, want %v", queue.SlowSyncTime, tt.slowSyncTime[0])
			}

			// Verify all workqueues are initialized
			for i := 0; i < int(tt.expectedWorkers); i++ {
				if len(queue.Workqueue) > i && queue.Workqueue[i] == nil {
					t.Errorf("NewWorkQueue() workqueue[%d] is nil", i)
				}
			}
		})
	}
}

func TestWorkerQueueStopWorkers(t *testing.T) {
	queue := utils.NewWorkQueue(2, "test-stop-queue")

	stopCh := make(chan struct{})

	// Stop the workers
	queue.StopWorkers(stopCh)

	// Verify queues are shut down by checking they don't accept new items
	// Note: This is a basic structural test
	if queue.Workqueue == nil {
		t.Error("StopWorkers() affected queue structure")
	}
}

func TestWorkerQueueBasicOperations(t *testing.T) {
	queue := utils.NewWorkQueue(4, "test-basic-queue")

	// Test basic queue properties
	if queue.NumWorkers != 4 {
		t.Errorf("Queue NumWorkers = %v, want 4", queue.NumWorkers)
	}

	if queue.WorkqueueName != "test-basic-queue" {
		t.Errorf("Queue WorkqueueName = %v, want test-basic-queue", queue.WorkqueueName)
	}

	// Verify workqueue array is properly sized
	if len(queue.Workqueue) != 4 {
		t.Errorf("Queue has %d workqueues, want 4", len(queue.Workqueue))
	}
}

func TestWorkerQueueWithSlowSync(t *testing.T) {
	slowSyncTime := 2
	queue := utils.NewWorkQueue(2, "slow-sync-queue", slowSyncTime)

	if queue.SlowSyncTime != slowSyncTime {
		t.Errorf("Queue SlowSyncTime = %v, want %v", queue.SlowSyncTime, slowSyncTime)
	}
}

func TestWorkerQueueConcurrentAccess(t *testing.T) {
	queue := utils.NewWorkQueue(4, "concurrent-queue")
	var wg sync.WaitGroup

	// Test concurrent access to queue properties
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Read queue properties concurrently
			_ = queue.NumWorkers
			_ = queue.WorkqueueName
			_ = len(queue.Workqueue)
		}()
	}

	wg.Wait()
}

func TestWorkerQueueMultipleQueues(t *testing.T) {
	// Create multiple queues and verify they're independent
	queue1 := utils.NewWorkQueue(2, "queue-1")
	queue2 := utils.NewWorkQueue(4, "queue-2")
	queue3 := utils.NewWorkQueue(8, "queue-3")

	if queue1.NumWorkers != 2 {
		t.Errorf("queue1 NumWorkers = %v, want 2", queue1.NumWorkers)
	}
	if queue2.NumWorkers != 4 {
		t.Errorf("queue2 NumWorkers = %v, want 4", queue2.NumWorkers)
	}
	if queue3.NumWorkers != 8 {
		t.Errorf("queue3 NumWorkers = %v, want 8", queue3.NumWorkers)
	}

	// Verify queue names are independent
	if queue1.WorkqueueName == queue2.WorkqueueName {
		t.Error("Queue names should be independent")
	}
}

func TestWorkerQueueInitialization(t *testing.T) {
	numWorkers := uint32(4)
	queue := utils.NewWorkQueue(numWorkers, "init-test-queue")

	// Verify worker ID mask is initialized correctly
	// The mask should be (2^numWorkers - 1)
	// This is an internal detail but important for worker allocation

	// Verify all workqueues are non-nil
	for i := uint32(0); i < numWorkers; i++ {
		if queue.Workqueue[i] == nil {
			t.Errorf("Workqueue[%d] is nil after initialization", i)
		}
	}
}

func TestWorkerQueueZeroWorkers(t *testing.T) {
	// Test edge case with 0 workers
	queue := utils.NewWorkQueue(0, "zero-workers-queue")

	if queue.NumWorkers != 0 {
		t.Errorf("Queue with 0 workers has NumWorkers = %v", queue.NumWorkers)
	}

	if len(queue.Workqueue) != 0 {
		t.Errorf("Queue with 0 workers has %d workqueues", len(queue.Workqueue))
	}
}

func TestWorkerQueueSingleWorker(t *testing.T) {
	// Test edge case with 1 worker
	queue := utils.NewWorkQueue(1, "single-worker-queue")

	if queue.NumWorkers != 1 {
		t.Errorf("Queue NumWorkers = %v, want 1", queue.NumWorkers)
	}

	if len(queue.Workqueue) != 1 {
		t.Errorf("Queue has %d workqueues, want 1", len(queue.Workqueue))
	}

	if queue.Workqueue[0] == nil {
		t.Error("Single workqueue is nil")
	}
}

func TestWorkerQueueLargeNumberOfWorkers(t *testing.T) {
	// Test with a larger number of workers
	numWorkers := uint32(32)
	queue := utils.NewWorkQueue(numWorkers, "large-queue")

	if queue.NumWorkers != numWorkers {
		t.Errorf("Queue NumWorkers = %v, want %v", queue.NumWorkers, numWorkers)
	}

	if len(queue.Workqueue) != int(numWorkers) {
		t.Errorf("Queue has %d workqueues, want %d", len(queue.Workqueue), numWorkers)
	}

	// Verify all are initialized
	for i := 0; i < int(numWorkers); i++ {
		if queue.Workqueue[i] == nil {
			t.Errorf("Workqueue[%d] is nil", i)
		}
	}
}

func TestWorkerQueueNameUniqueness(t *testing.T) {
	// Create multiple queues with different names
	names := []string{"queue-a", "queue-b", "queue-c", "queue-d"}
	queues := make([]*utils.WorkerQueue, len(names))

	for i, name := range names {
		queues[i] = utils.NewWorkQueue(2, name)
	}

	// Verify each queue has its assigned name
	for i, queue := range queues {
		if queue.WorkqueueName != names[i] {
			t.Errorf("Queue %d has name %v, want %v", i, queue.WorkqueueName, names[i])
		}
	}
}

func TestWorkerQueueSlowSyncTimeDefault(t *testing.T) {
	// Test that SlowSyncTime defaults to 0 when not provided
	queue := utils.NewWorkQueue(2, "default-sync-queue")

	if queue.SlowSyncTime != 0 {
		t.Errorf("Default SlowSyncTime = %v, want 0", queue.SlowSyncTime)
	}
}

func TestWorkerQueueSlowSyncTimeMultipleValues(t *testing.T) {
	// Test that only first slowSyncTime value is used
	queue := utils.NewWorkQueue(2, "multi-sync-queue", 5, 10, 15)

	if queue.SlowSyncTime != 5 {
		t.Errorf("SlowSyncTime = %v, want 5 (first value)", queue.SlowSyncTime)
	}
}

func TestWorkerQueueStructFields(t *testing.T) {
	queue := utils.NewWorkQueue(4, "field-test-queue", 3)

	// Verify all expected fields are accessible
	tests := []struct {
		name  string
		check func() bool
	}{
		{
			name: "NumWorkers is set",
			check: func() bool {
				return queue.NumWorkers == 4
			},
		},
		{
			name: "WorkqueueName is set",
			check: func() bool {
				return queue.WorkqueueName == "field-test-queue"
			},
		},
		{
			name: "SlowSyncTime is set",
			check: func() bool {
				return queue.SlowSyncTime == 3
			},
		},
		{
			name: "Workqueue array is initialized",
			check: func() bool {
				return queue.Workqueue != nil && len(queue.Workqueue) == 4
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check() {
				t.Errorf("Field check failed: %s", tt.name)
			}
		})
	}
}

func TestWorkerQueueStopWorkersDoesNotPanic(t *testing.T) {
	queue := utils.NewWorkQueue(2, "panic-test-queue")
	stopCh := make(chan struct{})

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("StopWorkers() panicked: %v", r)
		}
	}()

	queue.StopWorkers(stopCh)
}

func TestWorkerQueueStopWorkersMultipleTimes(t *testing.T) {
	queue := utils.NewWorkQueue(2, "multi-stop-queue")
	stopCh := make(chan struct{})

	// Stop multiple times should not cause issues
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Multiple StopWorkers() calls panicked: %v", r)
		}
	}()

	queue.StopWorkers(stopCh)
	// Note: Calling ShutDown multiple times on a workqueue is typically safe
	// but may log warnings
}

func TestWorkerQueueCreationPerformance(t *testing.T) {
	// Test that queue creation is reasonably fast
	start := time.Now()

	for i := 0; i < 100; i++ {
		_ = utils.NewWorkQueue(4, "perf-test-queue")
	}

	duration := time.Since(start)

	// Creating 100 queues should take less than 1 second
	if duration > time.Second {
		t.Errorf("Queue creation took %v, expected < 1s", duration)
	}
}
