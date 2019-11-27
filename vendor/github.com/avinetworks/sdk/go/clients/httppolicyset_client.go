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

// HTTPPolicySetClient is a client for avi HTTPPolicySet resource
type HTTPPolicySetClient struct {
	aviSession *session.AviSession
}

// NewHTTPPolicySetClient creates a new client for HTTPPolicySet resource
func NewHTTPPolicySetClient(aviSession *session.AviSession) *HTTPPolicySetClient {
	return &HTTPPolicySetClient{aviSession: aviSession}
}

func (client *HTTPPolicySetClient) getAPIPath(uuid string) string {
	path := "api/httppolicyset"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of HTTPPolicySet objects
func (client *HTTPPolicySetClient) GetAll() ([]*models.HTTPPolicySet, error) {
	var plist []*models.HTTPPolicySet
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing HTTPPolicySet by uuid
func (client *HTTPPolicySetClient) Get(uuid string) (*models.HTTPPolicySet, error) {
	var obj *models.HTTPPolicySet
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing HTTPPolicySet by name
func (client *HTTPPolicySetClient) GetByName(name string) (*models.HTTPPolicySet, error) {
	var obj *models.HTTPPolicySet
	err := client.aviSession.GetObjectByName("httppolicyset", name, &obj)
	return obj, err
}

// GetObject - Get an existing HTTPPolicySet by filters like name, cloud, tenant
// Api creates HTTPPolicySet object with every call.
func (client *HTTPPolicySetClient) GetObject(options ...session.ApiOptionsParams) (*models.HTTPPolicySet, error) {
	var obj *models.HTTPPolicySet
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("httppolicyset", newOptions...)
	return obj, err
}

// Create a new HTTPPolicySet object
func (client *HTTPPolicySetClient) Create(obj *models.HTTPPolicySet) (*models.HTTPPolicySet, error) {
	var robj *models.HTTPPolicySet
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing HTTPPolicySet object
func (client *HTTPPolicySetClient) Update(obj *models.HTTPPolicySet) (*models.HTTPPolicySet, error) {
	var robj *models.HTTPPolicySet
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing HTTPPolicySet object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.HTTPPolicySet
// or it should be json compatible of form map[string]interface{}
func (client *HTTPPolicySetClient) Patch(uuid string, patch interface{}, patchOp string) (*models.HTTPPolicySet, error) {
	var robj *models.HTTPPolicySet
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing HTTPPolicySet object with a given UUID
func (client *HTTPPolicySetClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing HTTPPolicySet object with a given name
func (client *HTTPPolicySetClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *HTTPPolicySetClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
