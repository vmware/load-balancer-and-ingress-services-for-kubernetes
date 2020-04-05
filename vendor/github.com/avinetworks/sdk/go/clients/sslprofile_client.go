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

// SSLProfileClient is a client for avi SSLProfile resource
type SSLProfileClient struct {
	aviSession *session.AviSession
}

// NewSSLProfileClient creates a new client for SSLProfile resource
func NewSSLProfileClient(aviSession *session.AviSession) *SSLProfileClient {
	return &SSLProfileClient{aviSession: aviSession}
}

func (client *SSLProfileClient) getAPIPath(uuid string) string {
	path := "api/sslprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of SSLProfile objects
func (client *SSLProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.SSLProfile, error) {
	var plist []*models.SSLProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing SSLProfile by uuid
func (client *SSLProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.SSLProfile, error) {
	var obj *models.SSLProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing SSLProfile by name
func (client *SSLProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.SSLProfile, error) {
	var obj *models.SSLProfile
	err := client.aviSession.GetObjectByName("sslprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing SSLProfile by filters like name, cloud, tenant
// Api creates SSLProfile object with every call.
func (client *SSLProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.SSLProfile, error) {
	var obj *models.SSLProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("sslprofile", newOptions...)
	return obj, err
}

// Create a new SSLProfile object
func (client *SSLProfileClient) Create(obj *models.SSLProfile, options ...session.ApiOptionsParams) (*models.SSLProfile, error) {
	var robj *models.SSLProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing SSLProfile object
func (client *SSLProfileClient) Update(obj *models.SSLProfile, options ...session.ApiOptionsParams) (*models.SSLProfile, error) {
	var robj *models.SSLProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing SSLProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.SSLProfile
// or it should be json compatible of form map[string]interface{}
func (client *SSLProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.SSLProfile, error) {
	var robj *models.SSLProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing SSLProfile object with a given UUID
func (client *SSLProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing SSLProfile object with a given name
func (client *SSLProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *SSLProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
