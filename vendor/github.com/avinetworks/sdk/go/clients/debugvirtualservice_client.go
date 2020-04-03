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

// DebugVirtualServiceClient is a client for avi DebugVirtualService resource
type DebugVirtualServiceClient struct {
	aviSession *session.AviSession
}

// NewDebugVirtualServiceClient creates a new client for DebugVirtualService resource
func NewDebugVirtualServiceClient(aviSession *session.AviSession) *DebugVirtualServiceClient {
	return &DebugVirtualServiceClient{aviSession: aviSession}
}

func (client *DebugVirtualServiceClient) getAPIPath(uuid string) string {
	path := "api/debugvirtualservice"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of DebugVirtualService objects
func (client *DebugVirtualServiceClient) GetAll(options ...session.ApiOptionsParams) ([]*models.DebugVirtualService, error) {
	var plist []*models.DebugVirtualService
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing DebugVirtualService by uuid
func (client *DebugVirtualServiceClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.DebugVirtualService, error) {
	var obj *models.DebugVirtualService
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing DebugVirtualService by name
func (client *DebugVirtualServiceClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.DebugVirtualService, error) {
	var obj *models.DebugVirtualService
	err := client.aviSession.GetObjectByName("debugvirtualservice", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing DebugVirtualService by filters like name, cloud, tenant
// Api creates DebugVirtualService object with every call.
func (client *DebugVirtualServiceClient) GetObject(options ...session.ApiOptionsParams) (*models.DebugVirtualService, error) {
	var obj *models.DebugVirtualService
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("debugvirtualservice", newOptions...)
	return obj, err
}

// Create a new DebugVirtualService object
func (client *DebugVirtualServiceClient) Create(obj *models.DebugVirtualService, options ...session.ApiOptionsParams) (*models.DebugVirtualService, error) {
	var robj *models.DebugVirtualService
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing DebugVirtualService object
func (client *DebugVirtualServiceClient) Update(obj *models.DebugVirtualService, options ...session.ApiOptionsParams) (*models.DebugVirtualService, error) {
	var robj *models.DebugVirtualService
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing DebugVirtualService object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.DebugVirtualService
// or it should be json compatible of form map[string]interface{}
func (client *DebugVirtualServiceClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.DebugVirtualService, error) {
	var robj *models.DebugVirtualService
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing DebugVirtualService object with a given UUID
func (client *DebugVirtualServiceClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing DebugVirtualService object with a given name
func (client *DebugVirtualServiceClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *DebugVirtualServiceClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
