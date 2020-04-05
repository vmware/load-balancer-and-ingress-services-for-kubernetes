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

// VIMgrHostRuntimeClient is a client for avi VIMgrHostRuntime resource
type VIMgrHostRuntimeClient struct {
	aviSession *session.AviSession
}

// NewVIMgrHostRuntimeClient creates a new client for VIMgrHostRuntime resource
func NewVIMgrHostRuntimeClient(aviSession *session.AviSession) *VIMgrHostRuntimeClient {
	return &VIMgrHostRuntimeClient{aviSession: aviSession}
}

func (client *VIMgrHostRuntimeClient) getAPIPath(uuid string) string {
	path := "api/vimgrhostruntime"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VIMgrHostRuntime objects
func (client *VIMgrHostRuntimeClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VIMgrHostRuntime, error) {
	var plist []*models.VIMgrHostRuntime
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VIMgrHostRuntime by uuid
func (client *VIMgrHostRuntimeClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VIMgrHostRuntime, error) {
	var obj *models.VIMgrHostRuntime
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VIMgrHostRuntime by name
func (client *VIMgrHostRuntimeClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VIMgrHostRuntime, error) {
	var obj *models.VIMgrHostRuntime
	err := client.aviSession.GetObjectByName("vimgrhostruntime", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VIMgrHostRuntime by filters like name, cloud, tenant
// Api creates VIMgrHostRuntime object with every call.
func (client *VIMgrHostRuntimeClient) GetObject(options ...session.ApiOptionsParams) (*models.VIMgrHostRuntime, error) {
	var obj *models.VIMgrHostRuntime
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vimgrhostruntime", newOptions...)
	return obj, err
}

// Create a new VIMgrHostRuntime object
func (client *VIMgrHostRuntimeClient) Create(obj *models.VIMgrHostRuntime, options ...session.ApiOptionsParams) (*models.VIMgrHostRuntime, error) {
	var robj *models.VIMgrHostRuntime
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VIMgrHostRuntime object
func (client *VIMgrHostRuntimeClient) Update(obj *models.VIMgrHostRuntime, options ...session.ApiOptionsParams) (*models.VIMgrHostRuntime, error) {
	var robj *models.VIMgrHostRuntime
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VIMgrHostRuntime object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VIMgrHostRuntime
// or it should be json compatible of form map[string]interface{}
func (client *VIMgrHostRuntimeClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VIMgrHostRuntime, error) {
	var robj *models.VIMgrHostRuntime
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VIMgrHostRuntime object with a given UUID
func (client *VIMgrHostRuntimeClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VIMgrHostRuntime object with a given name
func (client *VIMgrHostRuntimeClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VIMgrHostRuntimeClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
