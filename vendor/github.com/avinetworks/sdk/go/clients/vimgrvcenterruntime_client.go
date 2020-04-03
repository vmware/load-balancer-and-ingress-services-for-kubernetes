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

// VIMgrVcenterRuntimeClient is a client for avi VIMgrVcenterRuntime resource
type VIMgrVcenterRuntimeClient struct {
	aviSession *session.AviSession
}

// NewVIMgrVcenterRuntimeClient creates a new client for VIMgrVcenterRuntime resource
func NewVIMgrVcenterRuntimeClient(aviSession *session.AviSession) *VIMgrVcenterRuntimeClient {
	return &VIMgrVcenterRuntimeClient{aviSession: aviSession}
}

func (client *VIMgrVcenterRuntimeClient) getAPIPath(uuid string) string {
	path := "api/vimgrvcenterruntime"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VIMgrVcenterRuntime objects
func (client *VIMgrVcenterRuntimeClient) GetAll(options ...session.ApiOptionsParams) ([]*models.VIMgrVcenterRuntime, error) {
	var plist []*models.VIMgrVcenterRuntime
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing VIMgrVcenterRuntime by uuid
func (client *VIMgrVcenterRuntimeClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.VIMgrVcenterRuntime, error) {
	var obj *models.VIMgrVcenterRuntime
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing VIMgrVcenterRuntime by name
func (client *VIMgrVcenterRuntimeClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.VIMgrVcenterRuntime, error) {
	var obj *models.VIMgrVcenterRuntime
	err := client.aviSession.GetObjectByName("vimgrvcenterruntime", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing VIMgrVcenterRuntime by filters like name, cloud, tenant
// Api creates VIMgrVcenterRuntime object with every call.
func (client *VIMgrVcenterRuntimeClient) GetObject(options ...session.ApiOptionsParams) (*models.VIMgrVcenterRuntime, error) {
	var obj *models.VIMgrVcenterRuntime
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vimgrvcenterruntime", newOptions...)
	return obj, err
}

// Create a new VIMgrVcenterRuntime object
func (client *VIMgrVcenterRuntimeClient) Create(obj *models.VIMgrVcenterRuntime, options ...session.ApiOptionsParams) (*models.VIMgrVcenterRuntime, error) {
	var robj *models.VIMgrVcenterRuntime
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing VIMgrVcenterRuntime object
func (client *VIMgrVcenterRuntimeClient) Update(obj *models.VIMgrVcenterRuntime, options ...session.ApiOptionsParams) (*models.VIMgrVcenterRuntime, error) {
	var robj *models.VIMgrVcenterRuntime
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing VIMgrVcenterRuntime object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VIMgrVcenterRuntime
// or it should be json compatible of form map[string]interface{}
func (client *VIMgrVcenterRuntimeClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.VIMgrVcenterRuntime, error) {
	var robj *models.VIMgrVcenterRuntime
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing VIMgrVcenterRuntime object with a given UUID
func (client *VIMgrVcenterRuntimeClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing VIMgrVcenterRuntime object with a given name
func (client *VIMgrVcenterRuntimeClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *VIMgrVcenterRuntimeClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
