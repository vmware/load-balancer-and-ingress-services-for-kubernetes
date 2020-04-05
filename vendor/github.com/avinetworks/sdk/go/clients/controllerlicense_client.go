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

// ControllerLicenseClient is a client for avi ControllerLicense resource
type ControllerLicenseClient struct {
	aviSession *session.AviSession
}

// NewControllerLicenseClient creates a new client for ControllerLicense resource
func NewControllerLicenseClient(aviSession *session.AviSession) *ControllerLicenseClient {
	return &ControllerLicenseClient{aviSession: aviSession}
}

func (client *ControllerLicenseClient) getAPIPath(uuid string) string {
	path := "api/controllerlicense"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ControllerLicense objects
func (client *ControllerLicenseClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ControllerLicense, error) {
	var plist []*models.ControllerLicense
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ControllerLicense by uuid
func (client *ControllerLicenseClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ControllerLicense, error) {
	var obj *models.ControllerLicense
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ControllerLicense by name
func (client *ControllerLicenseClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ControllerLicense, error) {
	var obj *models.ControllerLicense
	err := client.aviSession.GetObjectByName("controllerlicense", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ControllerLicense by filters like name, cloud, tenant
// Api creates ControllerLicense object with every call.
func (client *ControllerLicenseClient) GetObject(options ...session.ApiOptionsParams) (*models.ControllerLicense, error) {
	var obj *models.ControllerLicense
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("controllerlicense", newOptions...)
	return obj, err
}

// Create a new ControllerLicense object
func (client *ControllerLicenseClient) Create(obj *models.ControllerLicense, options ...session.ApiOptionsParams) (*models.ControllerLicense, error) {
	var robj *models.ControllerLicense
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ControllerLicense object
func (client *ControllerLicenseClient) Update(obj *models.ControllerLicense, options ...session.ApiOptionsParams) (*models.ControllerLicense, error) {
	var robj *models.ControllerLicense
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ControllerLicense object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ControllerLicense
// or it should be json compatible of form map[string]interface{}
func (client *ControllerLicenseClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ControllerLicense, error) {
	var robj *models.ControllerLicense
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ControllerLicense object with a given UUID
func (client *ControllerLicenseClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ControllerLicense object with a given name
func (client *ControllerLicenseClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ControllerLicenseClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
