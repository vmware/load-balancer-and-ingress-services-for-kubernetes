/*
 * [2013] - [2019] Avi Networks Incorporated
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
	Stopped      bool
	ShutdownChan chan string
	Interval     time.Duration
	SyncFunction func()
}

func NewFullSyncThread(interval time.Duration) *FullSyncThread {
	return &FullSyncThread{
		Stopped:      false,
		ShutdownChan: make(chan string),
		Interval:     interval,
	}
}

func (w *FullSyncThread) Run() {
	defer close(w.ShutdownChan)
	time.Sleep(w.Interval)
	AviLog.Info.Printf("Started the Full Sync Worker")
	for {
		select {
		case <-w.ShutdownChan:
			AviLog.Info.Printf("Shutting down full sync go routine")
			return
		case <-time.After(w.Interval):
			break
		}
		w.SyncFunction()
	}
}

func (w *FullSyncThread) Shutdown() {
	w.Stopped = true
	w.ShutdownChan <- "shutdown"
}
