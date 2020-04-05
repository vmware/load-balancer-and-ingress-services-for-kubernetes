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

// ServiceEngineClient is a client for avi ServiceEngine resource
type ServiceEngineClient struct {
	aviSession *session.AviSession
}

// NewServiceEngineClient creates a new client for ServiceEngine resource
func NewServiceEngineClient(aviSession *session.AviSession) *ServiceEngineClient {
	return &ServiceEngineClient{aviSession: aviSession}
}

func (client *ServiceEngineClient) getAPIPath(uuid string) string {
	path := "api/serviceengine"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ServiceEngine objects
func (client *ServiceEngineClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ServiceEngine, error) {
	var plist []*models.ServiceEngine
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ServiceEngine by uuid
func (client *ServiceEngineClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ServiceEngine, error) {
	var obj *models.ServiceEngine
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ServiceEngine by name
func (client *ServiceEngineClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ServiceEngine, error) {
	var obj *models.ServiceEngine
	err := client.aviSession.GetObjectByName("serviceengine", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ServiceEngine by filters like name, cloud, tenant
// Api creates ServiceEngine object with every call.
func (client *ServiceEngineClient) GetObject(options ...session.ApiOptionsParams) (*models.ServiceEngine, error) {
	var obj *models.ServiceEngine
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("serviceengine", newOptions...)
	return obj, err
}

// Create a new ServiceEngine object
func (client *ServiceEngineClient) Create(obj *models.ServiceEngine, options ...session.ApiOptionsParams) (*models.ServiceEngine, error) {
	var robj *models.ServiceEngine
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ServiceEngine object
func (client *ServiceEngineClient) Update(obj *models.ServiceEngine, options ...session.ApiOptionsParams) (*models.ServiceEngine, error) {
	var robj *models.ServiceEngine
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ServiceEngine object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ServiceEngine
// or it should be json compatible of form map[string]interface{}
func (client *ServiceEngineClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ServiceEngine, error) {
	var robj *models.ServiceEngine
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ServiceEngine object with a given UUID
func (client *ServiceEngineClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ServiceEngine object with a given name
func (client *ServiceEngineClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ServiceEngineClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
