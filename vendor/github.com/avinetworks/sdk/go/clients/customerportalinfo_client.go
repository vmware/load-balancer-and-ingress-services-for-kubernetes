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

// CustomerPortalInfoClient is a client for avi CustomerPortalInfo resource
type CustomerPortalInfoClient struct {
	aviSession *session.AviSession
}

// NewCustomerPortalInfoClient creates a new client for CustomerPortalInfo resource
func NewCustomerPortalInfoClient(aviSession *session.AviSession) *CustomerPortalInfoClient {
	return &CustomerPortalInfoClient{aviSession: aviSession}
}

func (client *CustomerPortalInfoClient) getAPIPath(uuid string) string {
	path := "api/customerportalinfo"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of CustomerPortalInfo objects
func (client *CustomerPortalInfoClient) GetAll(options ...session.ApiOptionsParams) ([]*models.CustomerPortalInfo, error) {
	var plist []*models.CustomerPortalInfo
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing CustomerPortalInfo by uuid
func (client *CustomerPortalInfoClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.CustomerPortalInfo, error) {
	var obj *models.CustomerPortalInfo
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing CustomerPortalInfo by name
func (client *CustomerPortalInfoClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.CustomerPortalInfo, error) {
	var obj *models.CustomerPortalInfo
	err := client.aviSession.GetObjectByName("customerportalinfo", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing CustomerPortalInfo by filters like name, cloud, tenant
// Api creates CustomerPortalInfo object with every call.
func (client *CustomerPortalInfoClient) GetObject(options ...session.ApiOptionsParams) (*models.CustomerPortalInfo, error) {
	var obj *models.CustomerPortalInfo
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("customerportalinfo", newOptions...)
	return obj, err
}

// Create a new CustomerPortalInfo object
func (client *CustomerPortalInfoClient) Create(obj *models.CustomerPortalInfo, options ...session.ApiOptionsParams) (*models.CustomerPortalInfo, error) {
	var robj *models.CustomerPortalInfo
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing CustomerPortalInfo object
func (client *CustomerPortalInfoClient) Update(obj *models.CustomerPortalInfo, options ...session.ApiOptionsParams) (*models.CustomerPortalInfo, error) {
	var robj *models.CustomerPortalInfo
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing CustomerPortalInfo object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.CustomerPortalInfo
// or it should be json compatible of form map[string]interface{}
func (client *CustomerPortalInfoClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.CustomerPortalInfo, error) {
	var robj *models.CustomerPortalInfo
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing CustomerPortalInfo object with a given UUID
func (client *CustomerPortalInfoClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing CustomerPortalInfo object with a given name
func (client *CustomerPortalInfoClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *CustomerPortalInfoClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
