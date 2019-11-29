/***************************************************************************
 *
 * AVI CONFIDENTIAL
 * __________________
 *
 * [2013] - [2018] Avi Networks Incorporated
 * All Rights Reserved.
 *
 * NOTICE: All information contained herein is, and remains the property
 * of Avi Networks Incorporated and its suppliers, if any. The intellectual
 * and technical concepts contained herein are proprietary to Avi Networks
 * Incorporated, and its suppliers and are covered by U.S. and Foreign
 * Patents, patents in process, and are protected by trade secret or
 * copyright law, and other laws. Dissemination of this information or
 * reproduction of this material is strictly forbidden unless prior written
 * permission is obtained from Avi Networks Incorporated.
 */

package clients

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

import (
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
)

// VIMgrVMRuntimeClient is a client for avi VIMgrVMRuntime resource
type VIMgrVMRuntimeClient struct {
	aviSession *session.AviSession
}

// NewVIMgrVMRuntimeClient creates a new client for VIMgrVMRuntime resource
func NewVIMgrVMRuntimeClient(aviSession *session.AviSession) *VIMgrVMRuntimeClient {
	return &VIMgrVMRuntimeClient{aviSession: aviSession}
}

func (client *VIMgrVMRuntimeClient) getAPIPath(uuid string) string {
	path := "api/vimgrvmruntime"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VIMgrVMRuntime objects
func (client *VIMgrVMRuntimeClient) GetAll() ([]*models.VIMgrVMRuntime, error) {
	var plist []*models.VIMgrVMRuntime
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing VIMgrVMRuntime by uuid
func (client *VIMgrVMRuntimeClient) Get(uuid string) (*models.VIMgrVMRuntime, error) {
	var obj *models.VIMgrVMRuntime
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing VIMgrVMRuntime by name
func (client *VIMgrVMRuntimeClient) GetByName(name string) (*models.VIMgrVMRuntime, error) {
	var obj *models.VIMgrVMRuntime
	err := client.aviSession.GetObjectByName("vimgrvmruntime", name, &obj)
	return obj, err
}

// GetObject - Get an existing VIMgrVMRuntime by filters like name, cloud, tenant
// Api creates VIMgrVMRuntime object with every call.
func (client *VIMgrVMRuntimeClient) GetObject(options ...session.ApiOptionsParams) (*models.VIMgrVMRuntime, error) {
	var obj *models.VIMgrVMRuntime
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vimgrvmruntime", newOptions...)
	return obj, err
}

// Create a new VIMgrVMRuntime object
func (client *VIMgrVMRuntimeClient) Create(obj *models.VIMgrVMRuntime) (*models.VIMgrVMRuntime, error) {
	var robj *models.VIMgrVMRuntime
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing VIMgrVMRuntime object
func (client *VIMgrVMRuntimeClient) Update(obj *models.VIMgrVMRuntime) (*models.VIMgrVMRuntime, error) {
	var robj *models.VIMgrVMRuntime
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing VIMgrVMRuntime object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VIMgrVMRuntime
// or it should be json compatible of form map[string]interface{}
func (client *VIMgrVMRuntimeClient) Patch(uuid string, patch interface{}, patchOp string) (*models.VIMgrVMRuntime, error) {
	var robj *models.VIMgrVMRuntime
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing VIMgrVMRuntime object with a given UUID
func (client *VIMgrVMRuntimeClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing VIMgrVMRuntime object with a given name
func (client *VIMgrVMRuntimeClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *VIMgrVMRuntimeClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
