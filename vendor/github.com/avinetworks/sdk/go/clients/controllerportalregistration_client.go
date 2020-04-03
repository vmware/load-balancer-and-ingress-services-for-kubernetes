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

// ControllerPortalRegistrationClient is a client for avi ControllerPortalRegistration resource
type ControllerPortalRegistrationClient struct {
	aviSession *session.AviSession
}

// NewControllerPortalRegistrationClient creates a new client for ControllerPortalRegistration resource
func NewControllerPortalRegistrationClient(aviSession *session.AviSession) *ControllerPortalRegistrationClient {
	return &ControllerPortalRegistrationClient{aviSession: aviSession}
}

func (client *ControllerPortalRegistrationClient) getAPIPath(uuid string) string {
	path := "api/controllerportalregistration"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ControllerPortalRegistration objects
func (client *ControllerPortalRegistrationClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ControllerPortalRegistration, error) {
	var plist []*models.ControllerPortalRegistration
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ControllerPortalRegistration by uuid
func (client *ControllerPortalRegistrationClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ControllerPortalRegistration, error) {
	var obj *models.ControllerPortalRegistration
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ControllerPortalRegistration by name
func (client *ControllerPortalRegistrationClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ControllerPortalRegistration, error) {
	var obj *models.ControllerPortalRegistration
	err := client.aviSession.GetObjectByName("controllerportalregistration", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ControllerPortalRegistration by filters like name, cloud, tenant
// Api creates ControllerPortalRegistration object with every call.
func (client *ControllerPortalRegistrationClient) GetObject(options ...session.ApiOptionsParams) (*models.ControllerPortalRegistration, error) {
	var obj *models.ControllerPortalRegistration
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("controllerportalregistration", newOptions...)
	return obj, err
}

// Create a new ControllerPortalRegistration object
func (client *ControllerPortalRegistrationClient) Create(obj *models.ControllerPortalRegistration, options ...session.ApiOptionsParams) (*models.ControllerPortalRegistration, error) {
	var robj *models.ControllerPortalRegistration
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ControllerPortalRegistration object
func (client *ControllerPortalRegistrationClient) Update(obj *models.ControllerPortalRegistration, options ...session.ApiOptionsParams) (*models.ControllerPortalRegistration, error) {
	var robj *models.ControllerPortalRegistration
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ControllerPortalRegistration object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ControllerPortalRegistration
// or it should be json compatible of form map[string]interface{}
func (client *ControllerPortalRegistrationClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ControllerPortalRegistration, error) {
	var robj *models.ControllerPortalRegistration
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ControllerPortalRegistration object with a given UUID
func (client *ControllerPortalRegistrationClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ControllerPortalRegistration object with a given name
func (client *ControllerPortalRegistrationClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ControllerPortalRegistrationClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
