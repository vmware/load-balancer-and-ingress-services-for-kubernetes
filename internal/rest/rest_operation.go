/*
 * Copyright 2020-2021 VMware, Inc.
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
	"time"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/api/models"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"
	"k8s.io/apimachinery/pkg/util/runtime"

	avimodels "github.com/vmware/alb-sdk/go/models"
)

// modelSchema defines an interface to handle rest operations for an object type.
// For each object type (e.g. VirtualService), a new Model has to be implemented for this interface,
// and the instance should be added in supportedModelTypes.
// Usually the Model should have a refMap which has a mapping between reference object type and
// a function which removes the reference. The function of the interface - RemoveRef should return
// different functions based on the key of refMap.
type modelSchema interface {
	GetType() string
	RemoveRef(reftype string) func(*utils.RestOp, string) bool
}

var (
	virtualServiceModel = initVSModel()
	poolGroupModel      = initPGModel()
	L4PolicySetModel    = initL4PolSetModel()

	// List of Models for which we would try to remove object refs in case of error
	supportedModelTypes = map[string]modelSchema{
		"VirtualService": virtualServiceModel,
		"PoolGroup":      poolGroupModel,
		"L4PolicySet":    L4PolicySetModel,
	}

	// List of Objects for which we would handle error
	supportedObjForError = map[string]struct{}{
		"Pool":                 {},
		"VsVip":                {},
		"SSLKeyAndCertificate": {},
	}
)

type poolGroupSchema struct {
	Type   string
	refMap map[string]func(*utils.RestOp, string) bool
}

func (v *poolGroupSchema) GetType() string {
	return v.Type
}
func (v *poolGroupSchema) RemoveRef(refType string) func(*utils.RestOp, string) bool {
	return v.refMap[refType]
}

func initPGModel() *poolGroupSchema {
	pg := poolGroupSchema{}
	pg.Type = "PoolGroup"
	pg.refMap = map[string]func(*utils.RestOp, string) bool{
		"Pool": pg.removePoolRef,
	}
	return &pg
}

func (v *poolGroupSchema) removePoolRef(op *utils.RestOp, objRef string) bool {
	pg, ok := op.Obj.(*avimodels.PoolGroup)
	if !ok {
		utils.AviLog.Infof("Failed to remove Pool ref, object is not of type PoolGroup")
		return false
	}

	for i := range pg.Members {
		if strings.EqualFold(*pg.Members[i].PoolRef, objRef) {
			pg.Members = append(pg.Members[:i], pg.Members[i+1:]...)
			utils.AviLog.Infof("Successfully removed pool ref %s from PoolGroup: %s", objRef, *pg.Name)
			break
		}
	}
	op.Obj = pg
	return true
}

type virtualserviceSchema struct {
	Type   string
	refMap map[string]func(*utils.RestOp, string) bool
}

func (v *virtualserviceSchema) GetType() string {
	return v.Type
}

func initVSModel() *virtualserviceSchema {
	vs := virtualserviceSchema{}
	vs.Type = "VirtualService"
	vs.refMap = map[string]func(*utils.RestOp, string) bool{
		"SSLKeyAndCertificate": vs.removeCertRef,
		"Pool":                 vs.removePoolRef,
		"VsVip":                vs.removeVsVipRef,
	}
	return &vs
}

func (v *virtualserviceSchema) RemoveRef(refType string) func(*utils.RestOp, string) bool {
	return v.refMap[refType]
}

func (v *virtualserviceSchema) removeCertRef(op *utils.RestOp, objRef string) bool {
	vs, ok := op.Obj.(*avimodels.VirtualService)
	if !ok {
		utils.AviLog.Infof("Failed to remove SSL Cert ref, object is not of type Virtualservice")
		return false
	}
	for i, v := range vs.SslKeyAndCertificateRefs {
		if strings.EqualFold(v, objRef) {
			vs.SslKeyAndCertificateRefs = append(vs.SslKeyAndCertificateRefs[:i], vs.SslKeyAndCertificateRefs[i+1:]...)
			utils.AviLog.Infof("Successfully removed SSl Cert ref %s from VS: %s", objRef, *vs.Name)
		}
	}
	op.Obj = vs
	return true
}

func (v *virtualserviceSchema) removePoolRef(op *utils.RestOp, objRef string) bool {
	vs, ok := op.Obj.(*avimodels.VirtualService)
	if !ok {
		utils.AviLog.Infof("Failed to remove Pool ref, object is not of type Virtualservice")
		return false
	}
	if strings.EqualFold(*vs.PoolRef, objRef) {
		vs.PoolRef = nil
	}
	for i := range vs.ServicePoolSelect {
		//check if normal equal can be made to work
		if strings.EqualFold(*vs.ServicePoolSelect[i].ServicePoolRef, objRef) {
			vs.ServicePoolSelect[i].ServicePoolRef = nil
			utils.AviLog.Infof("Successfully removed Pool ref %s from VS: %s", objRef, *vs.Name)
		}
	}
	op.Obj = vs
	return true
}

func (v *virtualserviceSchema) removeVsVipRef(op *utils.RestOp, objRef string) bool {
	vs, ok := op.Obj.(*avimodels.VirtualService)
	if !ok {
		utils.AviLog.Infof("Failed to remove VsVip ref, object is not of type Virtualservice")
		return false
	}
	if strings.EqualFold(*vs.VsvipRef, objRef) {
		// If VsVip creation failed, then the VS Operations should be aborted
		utils.AviLog.Infof("VSVip creation failed, object ref won't be removed from Virtualservice")
		return false
	}
	return true
}

type L4PolicySetSchema struct {
	Type   string
	refMap map[string]func(*utils.RestOp, string) bool
}

func (v *L4PolicySetSchema) GetType() string {
	return v.Type
}

func initL4PolSetModel() *L4PolicySetSchema {
	l4PolSet := L4PolicySetSchema{}
	l4PolSet.Type = "L4PolicySet"
	l4PolSet.refMap = map[string]func(*utils.RestOp, string) bool{
		"Pool": l4PolSet.removePoolRef,
	}
	return &l4PolSet
}

func (v *L4PolicySetSchema) RemoveRef(refType string) func(*utils.RestOp, string) bool {
	return v.refMap[refType]
}

func (v *L4PolicySetSchema) removePoolRef(op *utils.RestOp, objRef string) bool {
	l4PolSet, ok := op.Obj.(*avimodels.L4PolicySet)
	if !ok {
		utils.AviLog.Infof("Failed to remove Pool ref, object is not of type L4PolicySet")
		return false
	}

	for i, rule := range l4PolSet.L4ConnectionPolicy.Rules {
		if strings.EqualFold(*rule.Action.SelectPool.PoolRef, objRef) {
			l4PolSet.L4ConnectionPolicy.Rules = append(l4PolSet.L4ConnectionPolicy.Rules[:i], l4PolSet.L4ConnectionPolicy.Rules[i+1:]...)
		}
	}
	op.Obj = l4PolSet
	return true
}

func removeObjRefFromRestOps(restOps []*utils.RestOp, objName, objType string) bool {
	if _, ok := supportedObjForError[objType]; !ok {
		utils.AviLog.Debugf("Ignoring error for unsupported type: %v", objType)
		return false
	}
	objRef := "/api/" + objType + "/?name=" + objName
	for i, op := range restOps {
		if m, ok := supportedModelTypes[op.Model]; ok {
			if removeFunc := m.RemoveRef(objType); removeFunc != nil {
				if !removeFunc(restOps[i], objRef) {
					return false
				}
			}
		}
	}
	return true
}

func isErrorRetryable(statusCode int, errMsg string) bool {
	// List of status codes for which we support retry
	if (statusCode >= 500 && statusCode < 599) || statusCode == 404 || statusCode == 401 || statusCode == 408 || statusCode == 409 {
		return true
	}
	if statusCode == 400 && strings.Contains(errMsg, lib.NoFreeIPError) {
		return true
	}
	if statusCode == 403 && strings.Contains(errMsg, lib.ConfigDisallowedDuringUpgradeError) {
		return true
	}
	return false
}

type AviRestClientPool struct {
	AviClient []*clients.AviClient
}

type RestOperator interface {
	AviRestOperateWrapper(aviClient *clients.AviClient, rest_ops []*utils.RestOp, key string) error
	AviRestOperate(c *clients.AviClient, rest_ops []*utils.RestOp, key string) error
	ExecuteRestAndPopulateCache(rest_ops []*utils.RestOp, aviObjKey avicache.NamespaceName, avimodel *nodes.AviObjectGraph, key string, isEvh bool, sslKey ...utils.NamespaceName) (bool, bool)
	SyncObjectStatuses()
	RestRespArrToObjByType(rest_op *utils.RestOp, obj_type string, key string) []map[string]interface{}
}

func NewRestOperator(restOp *RestOperations) RestOperator {
	if lib.AKOControlConfig().IsLeader() {
		return &leader{restOp: restOp}
	}
	return &follower{restOp: restOp}
}

type leader struct {
	restOp *RestOperations
}

func (l *leader) ExecuteRestAndPopulateCache(rest_ops []*utils.RestOp, aviObjKey avicache.NamespaceName, avimodel *nodes.AviObjectGraph, key string, isEvh bool, sslKey ...utils.NamespaceName) (bool, bool) {
	// Choose a avi client based on the model name hash. This would ensure that the same worker queue processes updates for a given VS all the time.
	shardSize := lib.GetshardSize()
	if shardSize == 0 {
		// Dedicated VS case
		shardSize = 8
	}
	var retry, fastRetry, processNextObj bool
	bkt := utils.Bkt(key, shardSize)
	if len(l.restOp.aviRestPoolClient.AviClient) > 0 && len(rest_ops) > 0 {
		utils.AviLog.Infof("key: %s, msg: processing in rest queue number: %v, caller %v", key, bkt, runtime.GetCaller())
		aviclient := l.restOp.aviRestPoolClient.AviClient[bkt]
		err := l.AviRestOperateWrapper(aviclient, rest_ops, key)
		if err == nil {
			models.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
			utils.AviLog.Debugf("key: %s, msg: rest call executed successfully, will update cache", key)

			// Add to local obj caches
			for _, rest_op := range rest_ops {
				l.restOp.PopulateOneCache(rest_op, aviObjKey, key)
			}

		} else if aviObjKey.Name == lib.DummyVSForStaleData {
			utils.AviLog.Warnf("key: %s, msg: error in rest request %v, for %s, won't retry", key, err.Error(), aviObjKey.Name)
			return false, processNextObj
		} else {
			var publishKey string
			if avimodel != nil && isEvh && len(avimodel.GetAviEvhVS()) > 0 {
				publishKey = avimodel.GetAviEvhVS()[0].Name
			} else if avimodel != nil && !isEvh && len(avimodel.GetAviVS()) > 0 {
				publishKey = avimodel.GetAviVS()[0].Name
			}

			if publishKey == "" {
				// This is a delete case for the virtualservice. Derive the virtualservice from the 'key'
				splitKeys := strings.Split(key, "/")
				if len(splitKeys) == 2 {
					publishKey = splitKeys[1]
				}
			}

			if l.restOp.CheckAndPublishForRetry(err, publishKey, key, avimodel) {
				return false, processNextObj
			}
			utils.AviLog.Warnf("key: %s, msg: there was an error sending the macro %v", key, err.Error())
			models.RestStatus.UpdateAviApiRestStatus("", err)
			for i := len(rest_ops) - 1; i >= 0; i-- {
				// Go over each of the failed requests and enqueue them to the worker queue for retry.
				if rest_ops[i].Err != nil {
					// check for VSVIP errors for blocked IP address updates
					if checkVsVipUpdateErrors(key, rest_ops[i]) {
						l.restOp.PopulateOneCache(rest_ops[i], aviObjKey, key)
						continue
					}

					// If it's for a SNI child, publish the parent VS's key
					refreshCacheForRetry := false
					if avimodel != nil && isEvh && len(avimodel.GetAviEvhVS()) > 0 {
						refreshCacheForRetry = true
					} else if avimodel != nil && !isEvh && len(avimodel.GetAviVS()) > 0 {
						refreshCacheForRetry = true
					}
					if refreshCacheForRetry {
						utils.AviLog.Warnf("key: %s, msg: Retrieved key for Retry:%s, object: %s", key, publishKey, rest_ops[i].ObjName)
						aviError, ok := rest_ops[i].Err.(session.AviError)
						if !ok {
							utils.AviLog.Infof("key: %s, msg: Error is not of type AviError, err: %v, %T", key, rest_ops[i].Err, rest_ops[i].Err)
							continue
						}
						retryable, fastRetryable, nextObj := l.restOp.RefreshCacheForRetryLayer(publishKey, aviObjKey, rest_ops[i], aviError, aviclient, avimodel, key, isEvh)
						retry = retry || retryable
						processNextObj = processNextObj || nextObj
						if avimodel.GetRetryCounter() != 0 {
							fastRetry = fastRetry || fastRetryable
						} else {
							fastRetry = false
							utils.AviLog.Warnf("key: %s, msg: retry count exhausted, would be added to slow retry queue", key)
						}
					} else {
						utils.AviLog.Warnf("key: %s, msg: Avi model not set, possibly a DELETE call", key)
						aviError, ok := rest_ops[i].Err.(session.AviError)
						// If it's 404, don't retry
						if ok {
							statuscode := aviError.HttpStatusCode
							if statuscode != 404 {
								l.restOp.PublishKeyToSlowRetryLayer(publishKey, key)
								//Here as it is 404 for specific object in a current child, AKO can go ahead with next child
								return false, true
							} else {
								l.restOp.AviVsCacheDel(rest_ops[i], aviObjKey, key)
							}
						}
					}
				} else {
					l.restOp.PopulateOneCache(rest_ops[i], aviObjKey, key)
				}
			}

			if retry {
				if fastRetry {
					l.restOp.PublishKeyToRetryLayer(publishKey, key)
				} else {
					l.restOp.PublishKeyToSlowRetryLayer(publishKey, key)
				}
			}
			return false, processNextObj
		}
	}
	return true, true
}

func (l *leader) AviRestOperateWrapper(aviClient *clients.AviClient, rest_ops []*utils.RestOp, key string) error {
	restTimeoutChan := make(chan error, 1)
	go func() {
		err := l.AviRestOperate(aviClient, rest_ops, key)
		restTimeoutChan <- err
	}()
	select {
	case err := <-restTimeoutChan:
		return err
	case <-time.After(lib.ControllerReqWaitTime * time.Second):
		utils.AviLog.Warnf("key: %s, msg: timed out waiting for rest response after %d seconds", key, lib.ControllerReqWaitTime)
		return errors.New("timed out waiting for rest response")
	}
}

func (l *leader) AviRestOperate(c *clients.AviClient, rest_ops []*utils.RestOp, key string) error {
	var failure bool
	for i, op := range rest_ops {
		SetTenant := session.SetTenant(op.Tenant)
		SetTenant(c.AviSession)
		if op.Version != "" {
			SetVersion := session.SetVersion(op.Version)
			SetVersion(c.AviSession)
		}
		switch op.Method {
		case utils.RestPost:
			op.Err = c.AviSession.Post(op.Path, op.Obj, &op.Response)
		case utils.RestPut:
			op.Err = c.AviSession.Put(op.Path, op.Obj, &op.Response)
		case utils.RestGet:
			op.Err = c.AviSession.Get(op.Path, &op.Response)
		case utils.RestPatch:
			op.Err = c.AviSession.Patch(op.Path, op.Obj, op.PatchOp,
				&op.Response)
		case utils.RestDelete:
			op.Err = c.AviSession.Delete(op.Path)
		default:
			utils.AviLog.Errorf("Unknown RestOp %v", op.Method)
			op.Err = fmt.Errorf("Unknown RestOp %v", op.Method)
		}
		if op.Err != nil {
			utils.AviLog.Warnf("key: %s, msg: RestOp method %v path %v tenant %v Obj %s returned err %s with response %s",
				key, op.Method, op.Path, op.Tenant, utils.Stringify(op.Obj), utils.Stringify(op.Err), utils.Stringify(op.Response))
			// Wrap the error into a websync error.
			err := &utils.WebSyncError{Err: op.Err, Operation: string(op.Method)}
			aviErr, ok := op.Err.(session.AviError)
			if !ok {
				utils.AviLog.Warnf("key: %s, msg: Error in rest operation is not of type AviError, err: %v, %T", key, op.Err, op.Err)
			} else if op.Model == "VsVip" && op.Method == utils.RestPut {
				utils.AviLog.Debugf("key: %s, msg: Error in rest operation for VsVip Put request.", key)
			} else if aviErr.HttpStatusCode == 404 && op.Method == utils.RestDelete {
				utils.AviLog.Warnf("key: %s, msg: Error during rest operation: %v, object of type %s not found in the controller. Ignoring err: %v", key, op.Method, op.Model, op.Err)
				continue
			} else if aviErr.HttpStatusCode == 409 && op.Method == utils.RestPost {
				utils.AviLog.Warnf("key: %s, msg: Error during rest operation: %v, object of type %s found in the controller. Ignoring err: %v", key, op.Method, op.Model, op.Err)
				failure = true
				continue
			} else if !isErrorRetryable(aviErr.HttpStatusCode, *aviErr.Message) {
				if op.Method != utils.RestPost {
					continue
				}
				if removeObjRefFromRestOps(rest_ops, op.ObjName, op.Model) {
					continue
				}
			}

			for j := i + 1; j < len(rest_ops); j++ {
				rest_ops[j].Err = errors.New("Aborted due to prev error")
			}
			return err
		} else {
			utils.AviLog.Debugf("key: %s, msg: RestOp method %v path %v tenant %v response %v objName %v",
				key, op.Method, op.Path, op.Tenant, utils.Stringify(op.Response), op.ObjName)
		}
	}
	if failure {
		return errors.New("required to populate cache and then retry")
	}
	return nil
}

type follower struct {
	restOp *RestOperations
}

func (f *follower) ExecuteRestAndPopulateCache(rest_ops []*utils.RestOp, aviObjKey avicache.NamespaceName, avimodel *nodes.AviObjectGraph, key string, isEvh bool, sslKey ...utils.NamespaceName) (bool, bool) {

	// Delay the REST calls in the follower.
	<-time.After(500 * time.Millisecond)

	// Choose a avi client based on the model name hash. This would ensure that the same worker queue processes updates for a given VS all the time.
	shardSize := lib.GetshardSize()
	if shardSize == 0 {
		// Dedicated VS case
		shardSize = 8
	}
	var retry, fastRetry, processNextObj bool
	bkt := utils.Bkt(key, shardSize)
	if len(f.restOp.aviRestPoolClient.AviClient) > 0 && len(rest_ops) > 0 {
		utils.AviLog.Infof("key: %s, msg: processing in rest queue number: %v, caller %v", key, bkt, runtime.GetCaller())
		aviclient := f.restOp.aviRestPoolClient.AviClient[bkt]
		err := f.AviRestOperateWrapper(aviclient, rest_ops, key)
		if err == nil {
			models.RestStatus.UpdateAviApiRestStatus(utils.AVIAPI_CONNECTED, nil)
			utils.AviLog.Debugf("key: %s, msg: rest call executed successfully, will update cache", key)

			// Add to local obj caches
			for _, rest_op := range rest_ops {
				f.restOp.PopulateOneCache(rest_op, aviObjKey, key)
			}

		} else if aviObjKey.Name == lib.DummyVSForStaleData {
			utils.AviLog.Warnf("key: %s, msg: error in rest request %v, for %s, won't retry", key, err.Error(), aviObjKey.Name)
			return false, processNextObj
		} else {
			var publishKey string
			if avimodel != nil && isEvh && len(avimodel.GetAviEvhVS()) > 0 {
				publishKey = avimodel.GetAviEvhVS()[0].Name
			} else if avimodel != nil && !isEvh && len(avimodel.GetAviVS()) > 0 {
				publishKey = avimodel.GetAviVS()[0].Name
			}

			if publishKey == "" {
				// This is a delete case for the virtualservice. Derive the virtualservice from the 'key'
				splitKeys := strings.Split(key, "/")
				if len(splitKeys) == 2 {
					publishKey = splitKeys[1]
				}
			}

			if err.Error() == "Got empty response for non-delete operation" ||
				err.Error() == "Got non-empty response for delete operation" {
				utils.AviLog.Warnf("key: %s, aborted the rest operation due to an error. err %s", key, err.Error())
				f.restOp.PublishKeyToRetryLayer(publishKey, key)
				return false, processNextObj
			}

			if f.restOp.CheckAndPublishForRetry(err, publishKey, key, avimodel) {
				return false, processNextObj
			}
			utils.AviLog.Warnf("key: %s, msg: there was an error sending the macro %v", key, err.Error())
			models.RestStatus.UpdateAviApiRestStatus("", err)
			for i := len(rest_ops) - 1; i >= 0; i-- {
				// Go over each of the failed requests and enqueue them to the worker queue for retry.
				if rest_ops[i].Err != nil {
					// check for VSVIP errors for blocked IP address updates
					if checkVsVipUpdateErrors(key, rest_ops[i]) {
						f.restOp.PopulateOneCache(rest_ops[i], aviObjKey, key)
						continue
					}

					// If it's for a SNI child, publish the parent VS's key
					refreshCacheForRetry := false
					if avimodel != nil && isEvh && len(avimodel.GetAviEvhVS()) > 0 {
						refreshCacheForRetry = true
					} else if avimodel != nil && !isEvh && len(avimodel.GetAviVS()) > 0 {
						refreshCacheForRetry = true
					}
					if refreshCacheForRetry {
						utils.AviLog.Warnf("key: %s, msg: Retrieved key for Retry:%s, object: %s", key, publishKey, rest_ops[i].ObjName)
						aviError, ok := rest_ops[i].Err.(session.AviError)
						if !ok {
							utils.AviLog.Infof("key: %s, msg: Error is not of type AviError, err: %v, %T", key, rest_ops[i].Err, rest_ops[i].Err)
							continue
						}
						retryable, fastRetryable, nextObj := f.restOp.RefreshCacheForRetryLayer(publishKey, aviObjKey, rest_ops[i], aviError, aviclient, avimodel, key, isEvh)
						retry = retry || retryable
						processNextObj = processNextObj || nextObj
						if avimodel.GetRetryCounter() != 0 {
							fastRetry = fastRetry || fastRetryable
						} else {
							fastRetry = false
							utils.AviLog.Warnf("key: %s, msg: retry count exhausted, would be added to slow retry queue", key)
						}
					} else {
						utils.AviLog.Warnf("key: %s, msg: Avi model not set, possibly a DELETE call", key)
						aviError, ok := rest_ops[i].Err.(session.AviError)
						// If it's 404, don't retry
						if ok {
							statuscode := aviError.HttpStatusCode
							if statuscode != 404 {
								f.restOp.PublishKeyToSlowRetryLayer(publishKey, key)
								//Here as it is 404 for specific object in a current child, AKO can go ahead with next child
								return false, true
							} else {
								f.restOp.AviVsCacheDel(rest_ops[i], aviObjKey, key)
							}
						}
					}
				} else {
					f.restOp.PopulateOneCache(rest_ops[i], aviObjKey, key)
				}
			}

			if retry {
				if fastRetry {
					f.restOp.PublishKeyToRetryLayer(publishKey, key)
				} else {
					f.restOp.PublishKeyToSlowRetryLayer(publishKey, key)
				}
			}
			return false, processNextObj
		}
	}
	return true, true
}

func (f *follower) AviRestOperateWrapper(aviClient *clients.AviClient, rest_ops []*utils.RestOp, key string) error {
	restTimeoutChan := make(chan error, 1)
	go func() {
		err := f.AviRestOperate(aviClient, rest_ops, key)
		restTimeoutChan <- err
	}()
	select {
	case err := <-restTimeoutChan:
		return err
	case <-time.After(lib.ControllerReqWaitTime * time.Second):
		utils.AviLog.Warnf("key: %s, msg: timed out waiting for rest response after %d seconds", key, lib.ControllerReqWaitTime)
		return errors.New("timed out waiting for rest response")
	}
}

func (f *follower) AviRestOperate(c *clients.AviClient, rest_ops []*utils.RestOp, key string) error {
	for i, op := range rest_ops {
		SetTenant := session.SetTenant(op.Tenant)
		SetTenant(c.AviSession)
		if op.Version != "" {
			SetVersion := session.SetVersion(op.Version)
			SetVersion(c.AviSession)
		}
		op.Path += "?name=" + op.ObjName
		utils.AviLog.Debugf("key: %s, msg: Got a REST operation: %s, %s", op.ObjName, op.Path)
		op.Err = c.AviSession.Get(op.Path, &op.Response)
		if op.Err != nil {
			utils.AviLog.Warnf("key: %s msg: RestOp method %v path %v tenant %v Obj %s returned err %s with response %s",
				key, op.Method, op.Path, op.Tenant, utils.Stringify(op.Obj), utils.Stringify(op.Err), utils.Stringify(op.Response))
			// Wrap the error into a websync error.
			err := &utils.WebSyncError{Err: op.Err, Operation: string(op.Method)}
			aviErr, ok := op.Err.(session.AviError)
			if !ok {
				utils.AviLog.Warnf("key: %s msg: Error in rest operation is not of type AviError, err: %v, %T", key, op.Err, op.Err)
			} else if op.Model == "VsVip" && op.Method == utils.RestPut {
				utils.AviLog.Debugf("key: %s msg: Error in rest operation for VsVip Put request.", key)
			} else if aviErr.HttpStatusCode == 404 && op.Method == utils.RestDelete {
				utils.AviLog.Warnf("key: %s msg: Error during rest operation: %v, object of type %s not found in the controller. Ignoring err: %v", key, op.Method, op.Model, op.Err)
				continue
			} else if !isErrorRetryable(aviErr.HttpStatusCode, *aviErr.Message) {
				if op.Method != utils.RestPost {
					continue
				}
				if removeObjRefFromRestOps(rest_ops, op.ObjName, op.Model) {
					continue
				}
			}

			for j := i + 1; j < len(rest_ops); j++ {
				rest_ops[j].Err = errors.New("Aborted due to prev error")
			}
			return err
		} else {
			utils.AviLog.Debugf("key: %s msg: RestOp method %v path %v tenant %v response %v objName %v",
				key, op.Method, op.Path, op.Tenant, utils.Stringify(op.Response), op.ObjName)
			if op.Method == utils.RestDelete && op.Response != nil {
				return errors.New("Got non-empty response for delete operation")
			}
			if resp, ok := op.Response.(map[string]interface{}); ok {
				if count, ok := resp["count"].(float64); ok {
					if op.Method != utils.RestDelete && count == 0 {
						return errors.New("Got empty response for non-delete operation")
					}
				}
			}
		}
	}
	return nil
}
