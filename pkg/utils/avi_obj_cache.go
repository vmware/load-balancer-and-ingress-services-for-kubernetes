/*
 * Copyright 2019-2020 VMware, Inc.
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
	"sync"
)

type CtrlPropCache struct {
	*AviCache
}

var ctrlPropOnce sync.Once
var ctrlPropCacheInstance *AviCache

func SharedCtrlProp() *CtrlPropCache {
	ctrlPropOnce.Do(func() {
		ctrlPropCacheInstance = NewAviCache()
	})
	return &CtrlPropCache{ctrlPropCacheInstance}
}

func (o *CtrlPropCache) PopulateCtrlProp(cp map[string]string) {
	for k := range cp {
		o.AviCacheAdd(k, cp[k])
	}
}

func (o *CtrlPropCache) PopulateCtrlAPIUserHeaders(userHeader map[string]string) {
	o.AviCacheAdd(ControllerAPIHeader, userHeader)
}

func (o *CtrlPropCache) PopulateCtrlAPIScheme(scheme string) {
	o.AviCacheAdd(ControllerAPIScheme, scheme)
}

func (o *CtrlPropCache) GetAllCtrlProp() map[string]string {
	ctrlProps := make(map[string]string)
	ctrlUsername, ok := o.AviCacheGet(ENV_CTRL_USERNAME)
	if !ok || ctrlUsername == nil {
		ctrlProps[ENV_CTRL_USERNAME] = ""
	} else {
		ctrlProps[ENV_CTRL_USERNAME] = ctrlUsername.(string)
	}
	ctrlPassword, ok := o.AviCacheGet(ENV_CTRL_PASSWORD)
	if !ok || ctrlPassword == nil {
		ctrlProps[ENV_CTRL_PASSWORD] = ""
	} else {
		ctrlProps[ENV_CTRL_PASSWORD] = ctrlPassword.(string)
	}
	ctrlAuthToken, ok := o.AviCacheGet(ENV_CTRL_AUTHTOKEN)
	if !ok || ctrlAuthToken == nil {
		ctrlProps[ENV_CTRL_AUTHTOKEN] = ""
	} else {
		ctrlProps[ENV_CTRL_AUTHTOKEN] = ctrlAuthToken.(string)
	}
	ctrlCAData, ok := o.AviCacheGet(ENV_CTRL_CADATA)
	if !ok || ctrlCAData == nil {
		ctrlProps[ENV_CTRL_CADATA] = ""
	} else {
		ctrlProps[ENV_CTRL_CADATA] = ctrlCAData.(string)
	}
	return ctrlProps
}

func (o *CtrlPropCache) GetCtrlUserHeader() map[string]string {
	headerData, ok := o.AviCacheGet(ControllerAPIHeader)
	if !ok || headerData == nil {
		return map[string]string{}
	}
	header, ok := headerData.(map[string]string)
	if !ok {
		return map[string]string{}
	}
	return header
}

func (o *CtrlPropCache) GetCtrlAPIScheme() string {
	apiScheme, ok := o.AviCacheGet(ControllerAPIScheme)
	if !ok || apiScheme == nil {
		return ""
	}
	scheme, ok := apiScheme.(string)
	if !ok {
		return ""
	}
	return scheme
}
