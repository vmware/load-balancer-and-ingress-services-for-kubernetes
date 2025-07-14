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
	"time"
)

type FullSyncThread struct {
	Stopped           bool
	ShutdownChan      chan string
	QuickSyncChan     chan struct{}
	Interval          time.Duration
	SyncFunction      func()
	QuickSyncFunction func(bool) error
}

func NewFullSyncThread(interval time.Duration) *FullSyncThread {
	return &FullSyncThread{
		Stopped:       false,
		ShutdownChan:  make(chan string),
		QuickSyncChan: make(chan struct{}, 1),
		Interval:      interval,
	}
}

func (w *FullSyncThread) Run() {
	defer close(w.ShutdownChan)
	AviLog.Infof("Started the Full Sync Worker")
	for {
		select {
		case <-w.ShutdownChan:
			AviLog.Infof("Shutting down full sync go routine")
			return
		case <-w.QuickSyncChan:
			// First the regular sync function - that syncs the cache
			w.SyncFunction()
			// Second the function that syncs the k8s objects.
			w.QuickSyncFunction(true)
			break
		case <-time.After(w.Interval):
			// Just the cache sync functions.
			w.SyncFunction()
			break
		}
	}
}

func (w *FullSyncThread) Shutdown() {
	w.Stopped = true
	w.ShutdownChan <- "shutdown"
}

func (w *FullSyncThread) QuickSync() {
	select {
	case w.QuickSyncChan <- struct{}{}:
		AviLog.Debugf("Scheduled QuickSync on Worker %v", w)
		return
	default:
	}

}
