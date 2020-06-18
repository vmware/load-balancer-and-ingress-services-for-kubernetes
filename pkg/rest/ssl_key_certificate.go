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

package rest

import (
	"errors"
	"fmt"

	avicache "ako/pkg/cache"
	"ako/pkg/lib"
	"ako/pkg/nodes"

	"github.com/avinetworks/container-lib/utils"
	avimodels "github.com/avinetworks/sdk/go/models"
	"github.com/davecgh/go-spew/spew"
)

func (rest *RestOperations) AviSSLBuild(ssl_node *nodes.AviTLSKeyCertNode, cache_obj *avicache.AviSSLCache) *utils.RestOp {
	name := ssl_node.Name
	tenant := fmt.Sprintf("/api/tenant/?name=%s", ssl_node.Tenant)
	certificate := string(ssl_node.Cert)
	key := string(ssl_node.Key)
	cr := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	sslkeycert := avimodels.SSLKeyAndCertificate{Name: &name,
		CreatedBy: &cr, TenantRef: &tenant, Certificate: &avimodels.SSLCertificate{Certificate: &certificate}, Key: &key}
	// TODO other fields like cloud_ref and lb algo

	macro := utils.AviRestObjMacro{ModelName: "SSLKeyAndCertificate", Data: sslkeycert}

	var path string
	var rest_op utils.RestOp
	if cache_obj != nil {
		path = "/api/sslkeyandcertificate/" + cache_obj.Uuid
		rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: sslkeycert,
			Tenant: ssl_node.Tenant, Model: "SSLKeyAndCertificate", Version: utils.CtrlVersion}
	} else {
		path = "/api/macro"
		rest_op = utils.RestOp{Path: path, Method: utils.RestPost, Obj: macro,
			Tenant: ssl_node.Tenant, Model: "SSLKeyAndCertificate", Version: utils.CtrlVersion}
	}
	return &rest_op
}

func (rest *RestOperations) AviSSLKeyCertDel(uuid string, tenant string) *utils.RestOp {
	path := "/api/sslkeyandcertificate/" + uuid
	rest_op := utils.RestOp{Path: path, Method: "DELETE",
		Tenant: tenant, Model: "SSLKeyAndCertificate", Version: utils.CtrlVersion}
	utils.AviLog.Info(spew.Sprintf("SSLCertKey DELETE Restop %v \n",
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviSSLKeyCertAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no reponse for sslkeycert, err: %s, response: %s", key, rest_op.Err, rest_op.Response)
		return errors.New("Errored rest_op")
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "sslkeyandcertificate", key)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warnf("Unable to find SSLKeyCert obj in resp %v", rest_op.Response)
		return errors.New("SSLKeyCert not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("Name not present in response %v", resp)
			continue
		}
		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("Uuid not present in response %v", resp)
			continue
		}
		_, ok = resp["certificate"].(map[string]interface{})
		if !ok {
			utils.AviLog.Warnf("Certificate not present in response %v", resp)
			continue
		}
		var SSLKeyAndCertificate string
		switch rest_op.Obj.(type) {
		case utils.AviRestObjMacro:
			SSLKeyAndCertificate = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.SSLKeyAndCertificate).Certificate.Certificate
		case avimodels.SSLKeyAndCertificate:
			SSLKeyAndCertificate = *rest_op.Obj.(avimodels.SSLKeyAndCertificate).Certificate.Certificate
		}
		checksum := lib.SSLKeyCertChecksum(name, SSLKeyAndCertificate)
		ssl_cache_obj := avicache.AviSSLCache{Name: name, Tenant: rest_op.Tenant,
			Uuid: uuid, CloudConfigCksum: checksum}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.SSLKeyCache.AviCacheAdd(k, &ssl_cache_obj)
		// Update the VS object
		if vsKey != (avicache.NamespaceName{}) {
			vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
			if ok {
				vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
				if found {
					utils.AviLog.Debugf("The VS cache before modification by SSLKeyCert is :%v", utils.Stringify(vs_cache_obj))
					vs_cache_obj.AddToSSLKeyCertCollection(k)
					utils.AviLog.Infof("Modified the VS cache object for SSLKeyCert Collection. The cache now is :%v", utils.Stringify(vs_cache_obj))
				}

			} else {
				vs_cache_obj := rest.cache.VsCacheMeta.AviCacheAddVS(vsKey)
				vs_cache_obj.AddToSSLKeyCertCollection(k)
				utils.AviLog.Info(spew.Sprintf("Added VS cache key during SSLKeyCert update %v val %v\n", vsKey,
					vs_cache_obj))
			}
			utils.AviLog.Info(spew.Sprintf("Added SSLKeyCert cache k %v val %v\n", k,
				ssl_cache_obj))
		}
	}

	return nil
}

func (rest *RestOperations) AviSSLCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	sslkey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	rest.cache.SSLKeyCache.AviCacheDelete(sslkey)
	if vsKey != (avicache.NamespaceName{}) {
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.RemoveFromSSLKeyCertCollection(sslkey)
			}
		}
	}

	return nil

}
