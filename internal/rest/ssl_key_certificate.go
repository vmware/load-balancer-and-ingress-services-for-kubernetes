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
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/avinetworks/sdk/go/models"
	"github.com/davecgh/go-spew/spew"
)

func (rest *RestOperations) AviSSLBuild(ssl_node *nodes.AviTLSKeyCertNode, cache_obj *avicache.AviSSLCache) *utils.RestOp {
	name := ssl_node.Name
	tenant := fmt.Sprintf("/api/tenant/?name=%s", ssl_node.Tenant)
	certificate := string(ssl_node.Cert)
	key := string(ssl_node.Key)
	cr := lib.AKOUser
	certType := ssl_node.Type
	sslkeycert := avimodels.SSLKeyAndCertificate{Name: &name,
		CreatedBy: &cr, TenantRef: &tenant, Certificate: &avimodels.SSLCertificate{Certificate: &certificate},
		Key: &key, Type: &certType}

	if ssl_node.CACert != "" {
		cacertRef := "/api/sslkeyandcertificate/?name=" + ssl_node.CACert
		caName := ssl_node.CACert
		sslkeycert.CaCerts = []*avimodels.CertificateAuthority{{
			CaRef: &cacertRef,
			Name:  &caName,
		}}
	}

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
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for sslkeycert, err: %s, response: %s", key, rest_op.Err, rest_op.Response)
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
		var SSLKeyAndCertificate avimodels.SSLKeyAndCertificate
		var cert, cacert string
		switch rest_op.Obj.(type) {
		case utils.AviRestObjMacro:
			SSLKeyAndCertificate = rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.SSLKeyAndCertificate)
		case avimodels.SSLKeyAndCertificate:
			SSLKeyAndCertificate = rest_op.Obj.(avimodels.SSLKeyAndCertificate)
		}
		cert = *SSLKeyAndCertificate.Certificate.Certificate
		hasCA := false
		if len(SSLKeyAndCertificate.CaCerts) > 0 {
			if SSLKeyAndCertificate.CaCerts[0].CaRef != nil {
				cacert = strings.TrimPrefix(*SSLKeyAndCertificate.CaCerts[0].CaRef, "/api/sslkeyandcertificate/?name=")
				hasCA = true
			}
		}
		checksum := lib.SSLKeyCertChecksum(name, cert, cacert)
		ssl_cache_obj := avicache.AviSSLCache{Name: name, Tenant: rest_op.Tenant,
			Uuid: uuid, CloudConfigCksum: checksum, HasCARef: hasCA}

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

func (rest *RestOperations) AviPkiProfileBuild(pki_node *nodes.AviPkiProfileNode, cache_obj *avicache.AviPkiProfileCache) *utils.RestOp {
	caCert := string(pki_node.CACert)
	tenant := fmt.Sprintf("/api/tenant/?name=%s", pki_node.Tenant)
	name := pki_node.Name
	var caCerts []*avimodels.SSLCertificate
	cr := lib.AKOUser
	crlcheck := false

	pkiobject := avimodels.PKIprofile{
		Name:      &name,
		CreatedBy: &cr,
		TenantRef: &tenant,
		CrlCheck:  &crlcheck,
		CaCerts: append(caCerts, &avimodels.SSLCertificate{
			Certificate: &caCert,
		}),
	}

	macro := utils.AviRestObjMacro{ModelName: "PKIprofile", Data: pkiobject}

	var path string
	var rest_op utils.RestOp
	if cache_obj != nil {
		path = "/api/pkiprofile/" + cache_obj.Uuid
		rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: pkiobject,
			Tenant: pki_node.Tenant, Model: "PKIprofile", Version: utils.CtrlVersion}
	} else {
		path = "/api/macro"
		rest_op = utils.RestOp{Path: path, Method: utils.RestPost, Obj: macro,
			Tenant: pki_node.Tenant, Model: "PKIprofile", Version: utils.CtrlVersion}
	}
	return &rest_op
}

func (rest *RestOperations) AviPkiProfileDel(uuid string, tenant string) *utils.RestOp {
	path := "/api/pkiprofile/" + uuid
	rest_op := utils.RestOp{Path: path, Method: "DELETE",
		Tenant: tenant, Model: "PKIprofile", Version: utils.CtrlVersion}
	utils.AviLog.Info(spew.Sprintf("PKIprofile DELETE Restop %v \n",
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviPkiProfileAdd(rest_op *utils.RestOp, poolKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("rest_op has err or no response for PkiProfileObj")
		return errors.New("Errored rest_op")
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "pkiprofile", key)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warnf("Unable to find PkiProfile obj in resp %v", rest_op.Response)
		return errors.New("PkiProfile not found")
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

		var pkiCertificate string
		switch rest_op.Obj.(type) {
		case utils.AviRestObjMacro:
			pkiCertificate = *rest_op.Obj.(utils.AviRestObjMacro).Data.(avimodels.PKIprofile).CaCerts[0].Certificate
		case avimodels.PKIprofile:
			pkiCertificate = *rest_op.Obj.(avimodels.PKIprofile).CaCerts[0].Certificate
		}

		checksum := lib.SSLKeyCertChecksum(name, pkiCertificate, "")
		pki_cache_obj := avicache.AviPkiProfileCache{
			Name:             name,
			Tenant:           rest_op.Tenant,
			Uuid:             uuid,
			CloudConfigCksum: checksum,
		}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.PKIProfileCache.AviCacheAdd(k, &pki_cache_obj)

		// Update the Pool object
		if poolKey != (avicache.NamespaceName{}) {
			pool_cache, ok := rest.cache.PoolCache.AviCacheGet(poolKey)
			if ok {
				pool_cache_obj, found := pool_cache.(*avicache.AviPoolCache)
				if found {
					utils.AviLog.Debugf("The Pool cache before modification by PkiProfile is :%v", utils.Stringify(pool_cache_obj))
					pool_cache_obj.PkiProfileCollection = k
					utils.AviLog.Infof("Modified the Pool cache object for PkiProfile Collection. The cache now is :%v", utils.Stringify(pool_cache_obj))
				}

			} else {
				pool_cache_obj := rest.cache.PoolCache.AviCacheAddPool(poolKey)
				pool_cache_obj.PkiProfileCollection = k
				utils.AviLog.Info(spew.Sprintf("Added Pool cache key during PkiProfile update %v val %v\n", poolKey,
					pool_cache_obj))
			}
			utils.AviLog.Info(spew.Sprintf("Added PkiProfile cache k %v val %v\n", k,
				pki_cache_obj))
		}
	}

	return nil
}

func (rest *RestOperations) AviPkiProfileCacheDel(rest_op *utils.RestOp, poolKey avicache.NamespaceName, key string) error {
	pkikey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	rest.cache.PKIProfileCache.AviCacheDelete(pkikey)

	if poolKey != (avicache.NamespaceName{}) {
		poolCache, ok := rest.cache.PoolCache.AviCacheGet(poolKey)
		if ok {
			poolCacheObj, found := poolCache.(*avicache.AviPoolCache)
			if found {
				poolCacheObj.PkiProfileCollection = avicache.NamespaceName{}
			}
		}
	}

	return nil
}
