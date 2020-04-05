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

// SecureChannelMappingClient is a client for avi SecureChannelMapping resource
type SecureChannelMappingClient struct {
	aviSession *session.AviSession
}

// NewSecureChannelMappingClient creates a new client for SecureChannelMapping resource
func NewSecureChannelMappingClient(aviSession *session.AviSession) *SecureChannelMappingClient {
	return &SecureChannelMappingClient{aviSession: aviSession}
}

func (client *SecureChannelMappingClient) getAPIPath(uuid string) string {
	path := "api/securechannelmapping"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of SecureChannelMapping objects
func (client *SecureChannelMappingClient) GetAll(options ...session.ApiOptionsParams) ([]*models.SecureChannelMapping, error) {
	var plist []*models.SecureChannelMapping
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing SecureChannelMapping by uuid
func (client *SecureChannelMappingClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.SecureChannelMapping, error) {
	var obj *models.SecureChannelMapping
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing SecureChannelMapping by name
func (client *SecureChannelMappingClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.SecureChannelMapping, error) {
	var obj *models.SecureChannelMapping
	err := client.aviSession.GetObjectByName("securechannelmapping", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing SecureChannelMapping by filters like name, cloud, tenant
// Api creates SecureChannelMapping object with every call.
func (client *SecureChannelMappingClient) GetObject(options ...session.ApiOptionsParams) (*models.SecureChannelMapping, error) {
	var obj *models.SecureChannelMapping
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("securechannelmapping", newOptions...)
	return obj, err
}

// Create a new SecureChannelMapping object
func (client *SecureChannelMappingClient) Create(obj *models.SecureChannelMapping, options ...session.ApiOptionsParams) (*models.SecureChannelMapping, error) {
	var robj *models.SecureChannelMapping
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing SecureChannelMapping object
func (client *SecureChannelMappingClient) Update(obj *models.SecureChannelMapping, options ...session.ApiOptionsParams) (*models.SecureChannelMapping, error) {
	var robj *models.SecureChannelMapping
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing SecureChannelMapping object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.SecureChannelMapping
// or it should be json compatible of form map[string]interface{}
func (client *SecureChannelMappingClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.SecureChannelMapping, error) {
	var robj *models.SecureChannelMapping
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing SecureChannelMapping object with a given UUID
func (client *SecureChannelMappingClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing SecureChannelMapping object with a given name
func (client *SecureChannelMappingClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *SecureChannelMappingClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
