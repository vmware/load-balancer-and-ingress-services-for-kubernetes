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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/avinetworks/sdk/go/clients"
	avimodels "github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
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

func AviRestOperate(c *clients.AviClient, rest_ops []*utils.RestOp) error {
	for i, op := range rest_ops {
		SetTenant := session.SetTenant(op.Tenant)
		SetTenant(c.AviSession)
		SetVersion := session.SetVersion(op.Version)
		SetVersion(c.AviSession)
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
			utils.AviLog.Warnf(`RestOp method %v path %v tenant %v Obj %s returned err %s with response %s`,
				op.Method, op.Path, op.Tenant, utils.Stringify(op.Obj), utils.Stringify(op.Err), utils.Stringify(op.Response))
			// Wrap the error into a websync error.
			err := &utils.WebSyncError{Err: op.Err, Operation: string(op.Method)}
			aviErr, ok := op.Err.(session.AviError)
			if !ok {
				utils.AviLog.Warnf("Error in rest operation is not of type AviError, err: %v, %T", op.Err, op.Err)
			} else if op.Model == "VsVip" && op.Method == utils.RestPut {
				utils.AviLog.Debugf("Error in rest operation for VsVip Put request.")
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
			utils.AviLog.Debugf(`RestOp method %v path %v tenant %v response %v`,
				op.Method, op.Path, op.Tenant, utils.Stringify(op.Response))
		}
	}
	return nil
}
