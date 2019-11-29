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

// ApplicationPersistenceProfileClient is a client for avi ApplicationPersistenceProfile resource
type ApplicationPersistenceProfileClient struct {
	aviSession *session.AviSession
}

// NewApplicationPersistenceProfileClient creates a new client for ApplicationPersistenceProfile resource
func NewApplicationPersistenceProfileClient(aviSession *session.AviSession) *ApplicationPersistenceProfileClient {
	return &ApplicationPersistenceProfileClient{aviSession: aviSession}
}

func (client *ApplicationPersistenceProfileClient) getAPIPath(uuid string) string {
	path := "api/applicationpersistenceprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ApplicationPersistenceProfile objects
func (client *ApplicationPersistenceProfileClient) GetAll() ([]*models.ApplicationPersistenceProfile, error) {
	var plist []*models.ApplicationPersistenceProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing ApplicationPersistenceProfile by uuid
func (client *ApplicationPersistenceProfileClient) Get(uuid string) (*models.ApplicationPersistenceProfile, error) {
	var obj *models.ApplicationPersistenceProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing ApplicationPersistenceProfile by name
func (client *ApplicationPersistenceProfileClient) GetByName(name string) (*models.ApplicationPersistenceProfile, error) {
	var obj *models.ApplicationPersistenceProfile
	err := client.aviSession.GetObjectByName("applicationpersistenceprofile", name, &obj)
	return obj, err
}

// GetObject - Get an existing ApplicationPersistenceProfile by filters like name, cloud, tenant
// Api creates ApplicationPersistenceProfile object with every call.
func (client *ApplicationPersistenceProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.ApplicationPersistenceProfile, error) {
	var obj *models.ApplicationPersistenceProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("applicationpersistenceprofile", newOptions...)
	return obj, err
}

// Create a new ApplicationPersistenceProfile object
func (client *ApplicationPersistenceProfileClient) Create(obj *models.ApplicationPersistenceProfile) (*models.ApplicationPersistenceProfile, error) {
	var robj *models.ApplicationPersistenceProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing ApplicationPersistenceProfile object
func (client *ApplicationPersistenceProfileClient) Update(obj *models.ApplicationPersistenceProfile) (*models.ApplicationPersistenceProfile, error) {
	var robj *models.ApplicationPersistenceProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing ApplicationPersistenceProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ApplicationPersistenceProfile
// or it should be json compatible of form map[string]interface{}
func (client *ApplicationPersistenceProfileClient) Patch(uuid string, patch interface{}, patchOp string) (*models.ApplicationPersistenceProfile, error) {
	var robj *models.ApplicationPersistenceProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing ApplicationPersistenceProfile object with a given UUID
func (client *ApplicationPersistenceProfileClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing ApplicationPersistenceProfile object with a given name
func (client *ApplicationPersistenceProfileClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *ApplicationPersistenceProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
