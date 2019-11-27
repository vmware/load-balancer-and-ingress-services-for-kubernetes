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

// MicroServiceClient is a client for avi MicroService resource
type MicroServiceClient struct {
	aviSession *session.AviSession
}

// NewMicroServiceClient creates a new client for MicroService resource
func NewMicroServiceClient(aviSession *session.AviSession) *MicroServiceClient {
	return &MicroServiceClient{aviSession: aviSession}
}

func (client *MicroServiceClient) getAPIPath(uuid string) string {
	path := "api/microservice"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of MicroService objects
func (client *MicroServiceClient) GetAll() ([]*models.MicroService, error) {
	var plist []*models.MicroService
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing MicroService by uuid
func (client *MicroServiceClient) Get(uuid string) (*models.MicroService, error) {
	var obj *models.MicroService
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing MicroService by name
func (client *MicroServiceClient) GetByName(name string) (*models.MicroService, error) {
	var obj *models.MicroService
	err := client.aviSession.GetObjectByName("microservice", name, &obj)
	return obj, err
}

// GetObject - Get an existing MicroService by filters like name, cloud, tenant
// Api creates MicroService object with every call.
func (client *MicroServiceClient) GetObject(options ...session.ApiOptionsParams) (*models.MicroService, error) {
	var obj *models.MicroService
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("microservice", newOptions...)
	return obj, err
}

// Create a new MicroService object
func (client *MicroServiceClient) Create(obj *models.MicroService) (*models.MicroService, error) {
	var robj *models.MicroService
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing MicroService object
func (client *MicroServiceClient) Update(obj *models.MicroService) (*models.MicroService, error) {
	var robj *models.MicroService
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing MicroService object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.MicroService
// or it should be json compatible of form map[string]interface{}
func (client *MicroServiceClient) Patch(uuid string, patch interface{}, patchOp string) (*models.MicroService, error) {
	var robj *models.MicroService
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing MicroService object with a given UUID
func (client *MicroServiceClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing MicroService object with a given name
func (client *MicroServiceClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *MicroServiceClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
