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

// StringGroupClient is a client for avi StringGroup resource
type StringGroupClient struct {
	aviSession *session.AviSession
}

// NewStringGroupClient creates a new client for StringGroup resource
func NewStringGroupClient(aviSession *session.AviSession) *StringGroupClient {
	return &StringGroupClient{aviSession: aviSession}
}

func (client *StringGroupClient) getAPIPath(uuid string) string {
	path := "api/stringgroup"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of StringGroup objects
func (client *StringGroupClient) GetAll() ([]*models.StringGroup, error) {
	var plist []*models.StringGroup
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing StringGroup by uuid
func (client *StringGroupClient) Get(uuid string) (*models.StringGroup, error) {
	var obj *models.StringGroup
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing StringGroup by name
func (client *StringGroupClient) GetByName(name string) (*models.StringGroup, error) {
	var obj *models.StringGroup
	err := client.aviSession.GetObjectByName("stringgroup", name, &obj)
	return obj, err
}

// GetObject - Get an existing StringGroup by filters like name, cloud, tenant
// Api creates StringGroup object with every call.
func (client *StringGroupClient) GetObject(options ...session.ApiOptionsParams) (*models.StringGroup, error) {
	var obj *models.StringGroup
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("stringgroup", newOptions...)
	return obj, err
}

// Create a new StringGroup object
func (client *StringGroupClient) Create(obj *models.StringGroup) (*models.StringGroup, error) {
	var robj *models.StringGroup
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing StringGroup object
func (client *StringGroupClient) Update(obj *models.StringGroup) (*models.StringGroup, error) {
	var robj *models.StringGroup
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing StringGroup object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.StringGroup
// or it should be json compatible of form map[string]interface{}
func (client *StringGroupClient) Patch(uuid string, patch interface{}, patchOp string) (*models.StringGroup, error) {
	var robj *models.StringGroup
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing StringGroup object with a given UUID
func (client *StringGroupClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing StringGroup object with a given name
func (client *StringGroupClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *StringGroupClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
