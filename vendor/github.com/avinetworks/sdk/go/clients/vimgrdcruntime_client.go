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

// VIMgrDCRuntimeClient is a client for avi VIMgrDCRuntime resource
type VIMgrDCRuntimeClient struct {
	aviSession *session.AviSession
}

// NewVIMgrDCRuntimeClient creates a new client for VIMgrDCRuntime resource
func NewVIMgrDCRuntimeClient(aviSession *session.AviSession) *VIMgrDCRuntimeClient {
	return &VIMgrDCRuntimeClient{aviSession: aviSession}
}

func (client *VIMgrDCRuntimeClient) getAPIPath(uuid string) string {
	path := "api/vimgrdcruntime"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of VIMgrDCRuntime objects
func (client *VIMgrDCRuntimeClient) GetAll() ([]*models.VIMgrDCRuntime, error) {
	var plist []*models.VIMgrDCRuntime
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing VIMgrDCRuntime by uuid
func (client *VIMgrDCRuntimeClient) Get(uuid string) (*models.VIMgrDCRuntime, error) {
	var obj *models.VIMgrDCRuntime
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing VIMgrDCRuntime by name
func (client *VIMgrDCRuntimeClient) GetByName(name string) (*models.VIMgrDCRuntime, error) {
	var obj *models.VIMgrDCRuntime
	err := client.aviSession.GetObjectByName("vimgrdcruntime", name, &obj)
	return obj, err
}

// GetObject - Get an existing VIMgrDCRuntime by filters like name, cloud, tenant
// Api creates VIMgrDCRuntime object with every call.
func (client *VIMgrDCRuntimeClient) GetObject(options ...session.ApiOptionsParams) (*models.VIMgrDCRuntime, error) {
	var obj *models.VIMgrDCRuntime
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("vimgrdcruntime", newOptions...)
	return obj, err
}

// Create a new VIMgrDCRuntime object
func (client *VIMgrDCRuntimeClient) Create(obj *models.VIMgrDCRuntime) (*models.VIMgrDCRuntime, error) {
	var robj *models.VIMgrDCRuntime
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing VIMgrDCRuntime object
func (client *VIMgrDCRuntimeClient) Update(obj *models.VIMgrDCRuntime) (*models.VIMgrDCRuntime, error) {
	var robj *models.VIMgrDCRuntime
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing VIMgrDCRuntime object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.VIMgrDCRuntime
// or it should be json compatible of form map[string]interface{}
func (client *VIMgrDCRuntimeClient) Patch(uuid string, patch interface{}, patchOp string) (*models.VIMgrDCRuntime, error) {
	var robj *models.VIMgrDCRuntime
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing VIMgrDCRuntime object with a given UUID
func (client *VIMgrDCRuntimeClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing VIMgrDCRuntime object with a given name
func (client *VIMgrDCRuntimeClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *VIMgrDCRuntimeClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
