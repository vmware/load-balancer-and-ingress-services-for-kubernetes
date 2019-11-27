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

// ApplicationClient is a client for avi Application resource
type ApplicationClient struct {
	aviSession *session.AviSession
}

// NewApplicationClient creates a new client for Application resource
func NewApplicationClient(aviSession *session.AviSession) *ApplicationClient {
	return &ApplicationClient{aviSession: aviSession}
}

func (client *ApplicationClient) getAPIPath(uuid string) string {
	path := "api/application"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of Application objects
func (client *ApplicationClient) GetAll() ([]*models.Application, error) {
	var plist []*models.Application
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing Application by uuid
func (client *ApplicationClient) Get(uuid string) (*models.Application, error) {
	var obj *models.Application
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing Application by name
func (client *ApplicationClient) GetByName(name string) (*models.Application, error) {
	var obj *models.Application
	err := client.aviSession.GetObjectByName("application", name, &obj)
	return obj, err
}

// GetObject - Get an existing Application by filters like name, cloud, tenant
// Api creates Application object with every call.
func (client *ApplicationClient) GetObject(options ...session.ApiOptionsParams) (*models.Application, error) {
	var obj *models.Application
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("application", newOptions...)
	return obj, err
}

// Create a new Application object
func (client *ApplicationClient) Create(obj *models.Application) (*models.Application, error) {
	var robj *models.Application
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing Application object
func (client *ApplicationClient) Update(obj *models.Application) (*models.Application, error) {
	var robj *models.Application
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing Application object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.Application
// or it should be json compatible of form map[string]interface{}
func (client *ApplicationClient) Patch(uuid string, patch interface{}, patchOp string) (*models.Application, error) {
	var robj *models.Application
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing Application object with a given UUID
func (client *ApplicationClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing Application object with a given name
func (client *ApplicationClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *ApplicationClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
