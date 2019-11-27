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

// ServiceEngineGroupClient is a client for avi ServiceEngineGroup resource
type ServiceEngineGroupClient struct {
	aviSession *session.AviSession
}

// NewServiceEngineGroupClient creates a new client for ServiceEngineGroup resource
func NewServiceEngineGroupClient(aviSession *session.AviSession) *ServiceEngineGroupClient {
	return &ServiceEngineGroupClient{aviSession: aviSession}
}

func (client *ServiceEngineGroupClient) getAPIPath(uuid string) string {
	path := "api/serviceenginegroup"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ServiceEngineGroup objects
func (client *ServiceEngineGroupClient) GetAll() ([]*models.ServiceEngineGroup, error) {
	var plist []*models.ServiceEngineGroup
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing ServiceEngineGroup by uuid
func (client *ServiceEngineGroupClient) Get(uuid string) (*models.ServiceEngineGroup, error) {
	var obj *models.ServiceEngineGroup
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing ServiceEngineGroup by name
func (client *ServiceEngineGroupClient) GetByName(name string) (*models.ServiceEngineGroup, error) {
	var obj *models.ServiceEngineGroup
	err := client.aviSession.GetObjectByName("serviceenginegroup", name, &obj)
	return obj, err
}

// GetObject - Get an existing ServiceEngineGroup by filters like name, cloud, tenant
// Api creates ServiceEngineGroup object with every call.
func (client *ServiceEngineGroupClient) GetObject(options ...session.ApiOptionsParams) (*models.ServiceEngineGroup, error) {
	var obj *models.ServiceEngineGroup
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("serviceenginegroup", newOptions...)
	return obj, err
}

// Create a new ServiceEngineGroup object
func (client *ServiceEngineGroupClient) Create(obj *models.ServiceEngineGroup) (*models.ServiceEngineGroup, error) {
	var robj *models.ServiceEngineGroup
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing ServiceEngineGroup object
func (client *ServiceEngineGroupClient) Update(obj *models.ServiceEngineGroup) (*models.ServiceEngineGroup, error) {
	var robj *models.ServiceEngineGroup
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing ServiceEngineGroup object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ServiceEngineGroup
// or it should be json compatible of form map[string]interface{}
func (client *ServiceEngineGroupClient) Patch(uuid string, patch interface{}, patchOp string) (*models.ServiceEngineGroup, error) {
	var robj *models.ServiceEngineGroup
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing ServiceEngineGroup object with a given UUID
func (client *ServiceEngineGroupClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing ServiceEngineGroup object with a given name
func (client *ServiceEngineGroupClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *ServiceEngineGroupClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
