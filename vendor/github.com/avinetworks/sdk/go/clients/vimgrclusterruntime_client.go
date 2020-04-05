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

// VIMgrClusterRuntimeClient is a client for avi VIMgrClusterRuntime resource
type VIMgrClusterRuntimeClient struct {
	aviSession *session.AviSession
}

// NewVIMgrClusterRuntimeClient creates a new client for VIMgrClusterRuntime resource
func NewVIMgrClusterRuntimeClient(aviSession *session.AviSession) *VIMgrClusterRuntimeClient {
	return &VIMgrClusterRuntimeClient{aviSession: aviSession}
}

func (client *VIMgrClusterRuntimeClient) getAPIPath(uuid string) string {
	path := "api/vimgrclusterruntime"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VIMgrClusterRuntime objects
func (client *VIMgrClusterRuntimeClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VIMgrClusterRuntime, error) {
	var plist []*models.VIMgrClusterRuntime
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VIMgrClusterRuntime by uuid
func (client *VIMgrClusterRuntimeClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VIMgrClusterRuntime, error) {
	var obj *models.VIMgrClusterRuntime
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VIMgrClusterRuntime by name
func (client *VIMgrClusterRuntimeClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VIMgrClusterRuntime, error) {
	var obj *models.VIMgrClusterRuntime
	err := client.aviSession.GetObjectByName("vimgrclusterruntime", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VIMgrClusterRuntime by filters like name, cloud, tenant
// Api creates VIMgrClusterRuntime object with every call.
func (client *VIMgrClusterRuntimeClient) GetObject(options ...session.ApiOptionsParams) (*models.VIMgrClusterRuntime, error) {
	var obj *models.VIMgrClusterRuntime
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vimgrclusterruntime", newOptions...)
	return obj, err
}

// Create a new VIMgrClusterRuntime object
func (client *VIMgrClusterRuntimeClient) Create(obj *models.VIMgrClusterRuntime, options ...session.ApiOptionsParams) (*models.VIMgrClusterRuntime, error) {
	var robj *models.VIMgrClusterRuntime
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VIMgrClusterRuntime object
func (client *VIMgrClusterRuntimeClient) Update(obj *models.VIMgrClusterRuntime, options ...session.ApiOptionsParams) (*models.VIMgrClusterRuntime, error) {
	var robj *models.VIMgrClusterRuntime
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VIMgrClusterRuntime object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VIMgrClusterRuntime
// or it should be json compatible of form map[string]interface{}
func (client *VIMgrClusterRuntimeClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VIMgrClusterRuntime, error) {
	var robj *models.VIMgrClusterRuntime
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VIMgrClusterRuntime object with a given UUID
func (client *VIMgrClusterRuntimeClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VIMgrClusterRuntime object with a given name
func (client *VIMgrClusterRuntimeClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VIMgrClusterRuntimeClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
