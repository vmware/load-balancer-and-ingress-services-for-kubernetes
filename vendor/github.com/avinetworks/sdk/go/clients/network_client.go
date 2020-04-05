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

// NetworkClient is a client for avi Network resource
type NetworkClient struct {
	aviSession *session.AviSession
}

// NewNetworkClient creates a new client for Network resource
func NewNetworkClient(aviSession *session.AviSession) *NetworkClient {
	return &NetworkClient{aviSession: aviSession}
}

func (client *NetworkClient) getAPIPath(uuid string) string {
	path := "api/network"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of Network objects
func (client *NetworkClient) GetAll(options ...session.ApiOptionsParams) ([]*models.Network, error) {
	var plist []*models.Network
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing Network by uuid
func (client *NetworkClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.Network, error) {
	var obj *models.Network
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing Network by name
func (client *NetworkClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.Network, error) {
	var obj *models.Network
	err := client.aviSession.GetObjectByName("network", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing Network by filters like name, cloud, tenant
// Api creates Network object with every call.
func (client *NetworkClient) GetObject(options ...session.ApiOptionsParams) (*models.Network, error) {
	var obj *models.Network
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("network", newOptions...)
	return obj, err
}

// Create a new Network object
func (client *NetworkClient) Create(obj *models.Network, options ...session.ApiOptionsParams) (*models.Network, error) {
	var robj *models.Network
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing Network object
func (client *NetworkClient) Update(obj *models.Network, options ...session.ApiOptionsParams) (*models.Network, error) {
	var robj *models.Network
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing Network object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.Network
// or it should be json compatible of form map[string]interface{}
func (client *NetworkClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.Network, error) {
	var robj *models.Network
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing Network object with a given UUID
func (client *NetworkClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing Network object with a given name
func (client *NetworkClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *NetworkClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
