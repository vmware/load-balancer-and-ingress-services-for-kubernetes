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

// APICLifsRuntimeClient is a client for avi APICLifsRuntime resource
type APICLifsRuntimeClient struct {
	aviSession *session.AviSession
}

// NewAPICLifsRuntimeClient creates a new client for APICLifsRuntime resource
func NewAPICLifsRuntimeClient(aviSession *session.AviSession) *APICLifsRuntimeClient {
	return &APICLifsRuntimeClient{aviSession: aviSession}
}

func (client *APICLifsRuntimeClient) getAPIPath(uuid string) string {
	path := "api/apiclifsruntime"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of APICLifsRuntime objects
func (client *APICLifsRuntimeClient) GetAll() ([]*models.APICLifsRuntime, error) {
	var plist []*models.APICLifsRuntime
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing APICLifsRuntime by uuid
func (client *APICLifsRuntimeClient) Get(uuid string) (*models.APICLifsRuntime, error) {
	var obj *models.APICLifsRuntime
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing APICLifsRuntime by name
func (client *APICLifsRuntimeClient) GetByName(name string) (*models.APICLifsRuntime, error) {
	var obj *models.APICLifsRuntime
	err := client.aviSession.GetObjectByName("apiclifsruntime", name, &obj)
	return obj, err
}

// GetObject - Get an existing APICLifsRuntime by filters like name, cloud, tenant
// Api creates APICLifsRuntime object with every call.
func (client *APICLifsRuntimeClient) GetObject(options ...session.ApiOptionsParams) (*models.APICLifsRuntime, error) {
	var obj *models.APICLifsRuntime
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("apiclifsruntime", newOptions...)
	return obj, err
}

// Create a new APICLifsRuntime object
func (client *APICLifsRuntimeClient) Create(obj *models.APICLifsRuntime) (*models.APICLifsRuntime, error) {
	var robj *models.APICLifsRuntime
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing APICLifsRuntime object
func (client *APICLifsRuntimeClient) Update(obj *models.APICLifsRuntime) (*models.APICLifsRuntime, error) {
	var robj *models.APICLifsRuntime
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing APICLifsRuntime object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.APICLifsRuntime
// or it should be json compatible of form map[string]interface{}
func (client *APICLifsRuntimeClient) Patch(uuid string, patch interface{}, patchOp string) (*models.APICLifsRuntime, error) {
	var robj *models.APICLifsRuntime
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing APICLifsRuntime object with a given UUID
func (client *APICLifsRuntimeClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing APICLifsRuntime object with a given name
func (client *APICLifsRuntimeClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *APICLifsRuntimeClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
