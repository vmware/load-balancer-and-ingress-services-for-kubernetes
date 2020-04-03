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

// ErrorPageProfileClient is a client for avi ErrorPageProfile resource
type ErrorPageProfileClient struct {
	aviSession *session.AviSession
}

// NewErrorPageProfileClient creates a new client for ErrorPageProfile resource
func NewErrorPageProfileClient(aviSession *session.AviSession) *ErrorPageProfileClient {
	return &ErrorPageProfileClient{aviSession: aviSession}
}

func (client *ErrorPageProfileClient) getAPIPath(uuid string) string {
	path := "api/errorpageprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ErrorPageProfile objects
func (client *ErrorPageProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ErrorPageProfile, error) {
	var plist []*models.ErrorPageProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ErrorPageProfile by uuid
func (client *ErrorPageProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ErrorPageProfile, error) {
	var obj *models.ErrorPageProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ErrorPageProfile by name
func (client *ErrorPageProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ErrorPageProfile, error) {
	var obj *models.ErrorPageProfile
	err := client.aviSession.GetObjectByName("errorpageprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ErrorPageProfile by filters like name, cloud, tenant
// Api creates ErrorPageProfile object with every call.
func (client *ErrorPageProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.ErrorPageProfile, error) {
	var obj *models.ErrorPageProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("errorpageprofile", newOptions...)
	return obj, err
}

// Create a new ErrorPageProfile object
func (client *ErrorPageProfileClient) Create(obj *models.ErrorPageProfile, options ...session.ApiOptionsParams) (*models.ErrorPageProfile, error) {
	var robj *models.ErrorPageProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ErrorPageProfile object
func (client *ErrorPageProfileClient) Update(obj *models.ErrorPageProfile, options ...session.ApiOptionsParams) (*models.ErrorPageProfile, error) {
	var robj *models.ErrorPageProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ErrorPageProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ErrorPageProfile
// or it should be json compatible of form map[string]interface{}
func (client *ErrorPageProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ErrorPageProfile, error) {
	var robj *models.ErrorPageProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ErrorPageProfile object with a given UUID
func (client *ErrorPageProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ErrorPageProfile object with a given name
func (client *ErrorPageProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ErrorPageProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
