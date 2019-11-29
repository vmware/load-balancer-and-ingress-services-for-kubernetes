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

// DebugServiceEngineClient is a client for avi DebugServiceEngine resource
type DebugServiceEngineClient struct {
	aviSession *session.AviSession
}

// NewDebugServiceEngineClient creates a new client for DebugServiceEngine resource
func NewDebugServiceEngineClient(aviSession *session.AviSession) *DebugServiceEngineClient {
	return &DebugServiceEngineClient{aviSession: aviSession}
}

func (client *DebugServiceEngineClient) getAPIPath(uuid string) string {
	path := "api/debugserviceengine"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of DebugServiceEngine objects
func (client *DebugServiceEngineClient) GetAll() ([]*models.DebugServiceEngine, error) {
	var plist []*models.DebugServiceEngine
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing DebugServiceEngine by uuid
func (client *DebugServiceEngineClient) Get(uuid string) (*models.DebugServiceEngine, error) {
	var obj *models.DebugServiceEngine
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing DebugServiceEngine by name
func (client *DebugServiceEngineClient) GetByName(name string) (*models.DebugServiceEngine, error) {
	var obj *models.DebugServiceEngine
	err := client.aviSession.GetObjectByName("debugserviceengine", name, &obj)
	return obj, err
}

// GetObject - Get an existing DebugServiceEngine by filters like name, cloud, tenant
// Api creates DebugServiceEngine object with every call.
func (client *DebugServiceEngineClient) GetObject(options ...session.ApiOptionsParams) (*models.DebugServiceEngine, error) {
	var obj *models.DebugServiceEngine
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("debugserviceengine", newOptions...)
	return obj, err
}

// Create a new DebugServiceEngine object
func (client *DebugServiceEngineClient) Create(obj *models.DebugServiceEngine) (*models.DebugServiceEngine, error) {
	var robj *models.DebugServiceEngine
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing DebugServiceEngine object
func (client *DebugServiceEngineClient) Update(obj *models.DebugServiceEngine) (*models.DebugServiceEngine, error) {
	var robj *models.DebugServiceEngine
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing DebugServiceEngine object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.DebugServiceEngine
// or it should be json compatible of form map[string]interface{}
func (client *DebugServiceEngineClient) Patch(uuid string, patch interface{}, patchOp string) (*models.DebugServiceEngine, error) {
	var robj *models.DebugServiceEngine
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing DebugServiceEngine object with a given UUID
func (client *DebugServiceEngineClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing DebugServiceEngine object with a given name
func (client *DebugServiceEngineClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *DebugServiceEngineClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
