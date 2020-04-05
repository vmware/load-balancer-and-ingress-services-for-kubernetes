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

// GslbClient is a client for avi Gslb resource
type GslbClient struct {
	aviSession *session.AviSession
}

// NewGslbClient creates a new client for Gslb resource
func NewGslbClient(aviSession *session.AviSession) *GslbClient {
	return &GslbClient{aviSession: aviSession}
}

func (client *GslbClient) getAPIPath(uuid string) string {
	path := "api/gslb"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of Gslb objects
func (client *GslbClient) GetAll(options ...session.ApiOptionsParams) ([]*models.Gslb, error) {
	var plist []*models.Gslb
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing Gslb by uuid
func (client *GslbClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.Gslb, error) {
	var obj *models.Gslb
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing Gslb by name
func (client *GslbClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.Gslb, error) {
	var obj *models.Gslb
	err := client.aviSession.GetObjectByName("gslb", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing Gslb by filters like name, cloud, tenant
// Api creates Gslb object with every call.
func (client *GslbClient) GetObject(options ...session.ApiOptionsParams) (*models.Gslb, error) {
	var obj *models.Gslb
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("gslb", newOptions...)
	return obj, err
}

// Create a new Gslb object
func (client *GslbClient) Create(obj *models.Gslb, options ...session.ApiOptionsParams) (*models.Gslb, error) {
	var robj *models.Gslb
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing Gslb object
func (client *GslbClient) Update(obj *models.Gslb, options ...session.ApiOptionsParams) (*models.Gslb, error) {
	var robj *models.Gslb
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing Gslb object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.Gslb
// or it should be json compatible of form map[string]interface{}
func (client *GslbClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.Gslb, error) {
	var robj *models.Gslb
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing Gslb object with a given UUID
func (client *GslbClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing Gslb object with a given name
func (client *GslbClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *GslbClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
