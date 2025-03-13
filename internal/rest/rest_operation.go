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

	"github.com/vmware/alb-sdk/go/clients"
	avimodels "github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
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
	op.Obj = vs
	return false
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
	if statusCode == 400 && (strings.Contains(errMsg, lib.VrfContextNotFoundError) || strings.Contains(errMsg, lib.VrfContextObjectNotFoundError)) {
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
	AviRestOperate(c *clients.AviClient, rest_ops []*utils.RestOp, key string) error
	isRetryRequired(key string, err error) bool
	SyncObjectStatuses()
	RestRespArrToObjByType(rest_op *utils.RestOp, obj_type string, key string) []map[string]interface{}
}

type (
	leader struct {
		restOp *RestOperations
	}
	follower struct {
		restOp *RestOperations
	}
)

func NewRestOperator(restOp *RestOperations, overrideLeaderFlag ...bool) RestOperator {
	if lib.AKOControlConfig().IsLeader() ||
		len(overrideLeaderFlag) > 0 && overrideLeaderFlag[0] {
		return &leader{restOp: restOp}
	}
	return &follower{restOp: restOp}
}

func (l *leader) isRetryRequired(key string, err error) bool {
	return false
}

func (l *leader) AviRestOperate(c *clients.AviClient, rest_ops []*utils.RestOp, key string) error {
	for i, op := range rest_ops {
		// This condition check is introduced to prevent any keys which is already present in the Graph
		// Queue from doing any POST/PUT/PATCH/GET operations at the controller when the `deleteConfig` is set.
		if lib.DisableSync &&
			op.Method != utils.RestDelete &&
			op.Model != "VrfContext" {
			utils.AviLog.Warnf("key: %s, msg: Sync is disabled, Only DELETE operation is allowed for models other than VRF model", key)
			continue
		}
		lib.IncrementRestOpCouter(utils.Stringify(op.Method), op.ObjName)
		if op.Version != "" {
			SetVersion := session.SetVersion(op.Version)
			SetVersion(c.AviSession)
		}
		switch op.Method {
		case utils.RestPost:
			op.Err = c.AviSession.Post(utils.GetUriEncoded(op.Path), op.Obj, &op.Response)
		case utils.RestPut:
			op.Err = c.AviSession.Put(utils.GetUriEncoded(op.Path), op.Obj, &op.Response)
		case utils.RestGet:
			op.Err = c.AviSession.Get(utils.GetUriEncoded(op.Path), &op.Response)
		case utils.RestPatch:
			op.Err = c.AviSession.Patch(utils.GetUriEncoded(op.Path), op.Obj, op.PatchOp,
				&op.Response)
		case utils.RestDelete:
			op.Err = c.AviSession.Delete(utils.GetUriEncoded(op.Path))
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
			} else if op.Model == "VrfContext" && aviErr.HttpStatusCode == 412 {
				utils.AviLog.Debugf("key: %s, msg: Error in rest operation for VrfContext Put request.", key)
			} else if op.Model == "VrfContext" && aviErr.HttpStatusCode == 400 && strings.Contains(*aviErr.Message, lib.VrfContextNoPermission) {
				// retry operation  by switching to admin tenant
				err = nil
				utils.AviLog.Warnf("key:%s, msg: Switching to Admin tenant from %s for %s method", key, lib.GetTenant(), op.Method)
				shardSize := lib.GetshardSize()
				if shardSize == 0 {
					// Dedicated VS case
					shardSize = 8
				}
				bkt := utils.Bkt(key, shardSize)
				client := cache.SharedAVIClients(lib.GetAdminTenant()).AviClient[bkt]
				op.Err = client.AviSession.Put(utils.GetUriEncoded(op.Path), op.Obj, &op.Response)
				if op.Err == nil {
					utils.AviLog.Debugf("key: %s, msg: RestOp method %v path %v tenant %v response %v objName %v",
						key, op.Method, op.Path, lib.GetAdminTenant(), utils.Stringify(op.Response), op.ObjName)
					continue
				}
				utils.AviLog.Warnf("key: %s, msg: RestOp method %v path %v tenant %v Obj %s returned err %s with response %s",
					key, op.Method, op.Path, lib.GetAdminTenant(), utils.Stringify(op.Obj), utils.Stringify(op.Err), utils.Stringify(op.Response))
				err = &utils.WebSyncError{Err: op.Err, Operation: string(op.Method)}
			} else if !isErrorRetryable(aviErr.HttpStatusCode, *aviErr.Message) {
				if op.Method != utils.RestPost {
					utils.AviLog.Warnf("key: %s, msg: RestOp method %v error %s is not retryable", key, op.Method, utils.Stringify(op.Err))
					continue
				}
				if removeObjRefFromRestOps(rest_ops, op.ObjName, op.Model) {
					utils.AviLog.Warnf("key: %s, msg: RestOp method %v error %s is not retryable. Removed object %s from rest op", key, op.Method, utils.Stringify(op.Err), op.ObjName)
					continue
				}
			}

			for j := i + 1; j < len(rest_ops); j++ {
				rest_ops[j].Err = errors.New("Aborted due to prev error")
			}
			return err
		} else {
			utils.AviLog.Infof("key: %s, msg: RestOp method %v path %v tenant %v response %v objName %v",
				key, op.Method, op.Path, op.Tenant, utils.Stringify(op.Response), op.ObjName)
		}
	}
	return nil
}

func (f *follower) isRetryRequired(key string, err error) bool {
	if err == nil {
		return false
	}
	switch err.Error() {
	case "Got empty response for non-delete operation",
		"Got non-empty response for delete operation":
		utils.AviLog.Debugf("key: %s, msg: aborted the rest operation due the reason: %s", key, err.Error())
		return true
	}
	return false
}

func (f *follower) AviRestOperate(c *clients.AviClient, rest_ops []*utils.RestOp, key string) error {

	// Adding a delay of 500ms before doing a GET operation in the follower AKO.
	// 500ms is selected to give leader AKO enough time to create the objects in the controller,
	// before the follower AKO does a GET operation. In case of a 404(object not found) error, the
	// follower AKO pushes the key to retry layer for retry.
	<-time.After(500 * time.Millisecond)

	for i, op := range rest_ops {
		if op.Version != "" {
			SetVersion := session.SetVersion(op.Version)
			SetVersion(c.AviSession)
		}

		// Path for GET operation is appended with include_name and created_by to make
		// the result unique. But in case of VS VIP objects, the created_by is not required
		// because of this reason, it is omitted.
		op.Path += "?name=" + op.ObjName + "&include_name=true"
		if op.Model == "VsVip" {
			op.Path += "&cloud_ref.name=" + utils.CloudName
		} else {
			op.Path += "&created_by=" + lib.AKOUser
		}

		utils.AviLog.Debugf("key: %s, msg: Got a REST operation: %s, %s", key, op.ObjName, op.Path)
		op.Err = c.AviSession.Get(utils.GetUriEncoded(op.Path), &op.Response)
		if op.Err != nil {
			utils.AviLog.Warnf("key: %s, msg: RestOp method %v path %v tenant %v Obj %s returned err %s with response %s",
				key, op.Method, op.Path, op.Tenant, utils.Stringify(op.Obj), utils.Stringify(op.Err), utils.Stringify(op.Response))
			// Wrap the error into a websync error.
			err := &utils.WebSyncError{Err: op.Err, Operation: string(op.Method)}
			aviErr, ok := op.Err.(session.AviError)
			if !ok {
				utils.AviLog.Warnf("key: %s, msg: Error in rest operation is not of type AviError, err: %v, %T", key, op.Err, op.Err)
			} else if aviErr.HttpStatusCode == 404 && op.Method == utils.RestDelete {
				utils.AviLog.Warnf("key: %s, msg: Error during rest operation: %v, object of type %s not found in the controller. Ignoring err: %v", key, op.Method, op.Model, op.Err)
				continue
			} else if !isErrorRetryable(aviErr.HttpStatusCode, *aviErr.Message) {
				if op.Method != utils.RestPost {
					utils.AviLog.Warnf("key: %s, msg: RestOp method %v error %s is not retryable", key, op.Method, utils.Stringify(op.Err))
					continue
				}
				if removeObjRefFromRestOps(rest_ops, op.ObjName, op.Model) {
					utils.AviLog.Warnf("key: %s, msg: RestOp method %v error %s is not retryable. Removed object %s from rest op", key, op.Method, utils.Stringify(op.Err), op.ObjName)
					continue
				}
			}

			for j := i + 1; j < len(rest_ops); j++ {
				rest_ops[j].Err = errors.New("Aborted due to prev error")
			}
			return err
		} else {
			utils.AviLog.Infof("key: %s, msg: RestOp method %v path %v tenant %v response %v objName %v",
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
