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

// UserActivityClient is a client for avi UserActivity resource
type UserActivityClient struct {
	aviSession *session.AviSession
}

// NewUserActivityClient creates a new client for UserActivity resource
func NewUserActivityClient(aviSession *session.AviSession) *UserActivityClient {
	return &UserActivityClient{aviSession: aviSession}
}

func (client *UserActivityClient) getAPIPath(uuid string) string {
	path := "api/useractivity"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of UserActivity objects
func (client *UserActivityClient) GetAll() ([]*models.UserActivity, error) {
	var plist []*models.UserActivity
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing UserActivity by uuid
func (client *UserActivityClient) Get(uuid string) (*models.UserActivity, error) {
	var obj *models.UserActivity
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing UserActivity by name
func (client *UserActivityClient) GetByName(name string) (*models.UserActivity, error) {
	var obj *models.UserActivity
	err := client.aviSession.GetObjectByName("useractivity", name, &obj)
	return obj, err
}

// GetObject - Get an existing UserActivity by filters like name, cloud, tenant
// Api creates UserActivity object with every call.
func (client *UserActivityClient) GetObject(options ...session.ApiOptionsParams) (*models.UserActivity, error) {
	var obj *models.UserActivity
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("useractivity", newOptions...)
	return obj, err
}

// Create a new UserActivity object
func (client *UserActivityClient) Create(obj *models.UserActivity) (*models.UserActivity, error) {
	var robj *models.UserActivity
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing UserActivity object
func (client *UserActivityClient) Update(obj *models.UserActivity) (*models.UserActivity, error) {
	var robj *models.UserActivity
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing UserActivity object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.UserActivity
// or it should be json compatible of form map[string]interface{}
func (client *UserActivityClient) Patch(uuid string, patch interface{}, patchOp string) (*models.UserActivity, error) {
	var robj *models.UserActivity
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing UserActivity object with a given UUID
func (client *UserActivityClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing UserActivity object with a given name
func (client *UserActivityClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *UserActivityClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
