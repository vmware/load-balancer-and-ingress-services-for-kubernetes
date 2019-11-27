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

// VIMgrSEVMRuntimeClient is a client for avi VIMgrSEVMRuntime resource
type VIMgrSEVMRuntimeClient struct {
	aviSession *session.AviSession
}

// NewVIMgrSEVMRuntimeClient creates a new client for VIMgrSEVMRuntime resource
func NewVIMgrSEVMRuntimeClient(aviSession *session.AviSession) *VIMgrSEVMRuntimeClient {
	return &VIMgrSEVMRuntimeClient{aviSession: aviSession}
}

func (client *VIMgrSEVMRuntimeClient) getAPIPath(uuid string) string {
	path := "api/vimgrsevmruntime"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VIMgrSEVMRuntime objects
func (client *VIMgrSEVMRuntimeClient) GetAll() ([]*models.VIMgrSEVMRuntime, error) {
	var plist []*models.VIMgrSEVMRuntime
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing VIMgrSEVMRuntime by uuid
func (client *VIMgrSEVMRuntimeClient) Get(uuid string) (*models.VIMgrSEVMRuntime, error) {
	var obj *models.VIMgrSEVMRuntime
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing VIMgrSEVMRuntime by name
func (client *VIMgrSEVMRuntimeClient) GetByName(name string) (*models.VIMgrSEVMRuntime, error) {
	var obj *models.VIMgrSEVMRuntime
	err := client.aviSession.GetObjectByName("vimgrsevmruntime", name, &obj)
	return obj, err
}

// GetObject - Get an existing VIMgrSEVMRuntime by filters like name, cloud, tenant
// Api creates VIMgrSEVMRuntime object with every call.
func (client *VIMgrSEVMRuntimeClient) GetObject(options ...session.ApiOptionsParams) (*models.VIMgrSEVMRuntime, error) {
	var obj *models.VIMgrSEVMRuntime
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vimgrsevmruntime", newOptions...)
	return obj, err
}

// Create a new VIMgrSEVMRuntime object
func (client *VIMgrSEVMRuntimeClient) Create(obj *models.VIMgrSEVMRuntime) (*models.VIMgrSEVMRuntime, error) {
	var robj *models.VIMgrSEVMRuntime
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing VIMgrSEVMRuntime object
func (client *VIMgrSEVMRuntimeClient) Update(obj *models.VIMgrSEVMRuntime) (*models.VIMgrSEVMRuntime, error) {
	var robj *models.VIMgrSEVMRuntime
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing VIMgrSEVMRuntime object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VIMgrSEVMRuntime
// or it should be json compatible of form map[string]interface{}
func (client *VIMgrSEVMRuntimeClient) Patch(uuid string, patch interface{}, patchOp string) (*models.VIMgrSEVMRuntime, error) {
	var robj *models.VIMgrSEVMRuntime
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing VIMgrSEVMRuntime object with a given UUID
func (client *VIMgrSEVMRuntimeClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing VIMgrSEVMRuntime object with a given name
func (client *VIMgrSEVMRuntimeClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *VIMgrSEVMRuntimeClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
