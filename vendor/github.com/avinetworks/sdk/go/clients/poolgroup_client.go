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

// PoolGroupClient is a client for avi PoolGroup resource
type PoolGroupClient struct {
	aviSession *session.AviSession
}

// NewPoolGroupClient creates a new client for PoolGroup resource
func NewPoolGroupClient(aviSession *session.AviSession) *PoolGroupClient {
	return &PoolGroupClient{aviSession: aviSession}
}

func (client *PoolGroupClient) getAPIPath(uuid string) string {
	path := "api/poolgroup"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of PoolGroup objects
func (client *PoolGroupClient) GetAll(options ...session.ApiOptionsParams) ([]*models.PoolGroup, error) {
	var plist []*models.PoolGroup
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing PoolGroup by uuid
func (client *PoolGroupClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.PoolGroup, error) {
	var obj *models.PoolGroup
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing PoolGroup by name
func (client *PoolGroupClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.PoolGroup, error) {
	var obj *models.PoolGroup
	err := client.aviSession.GetObjectByName("poolgroup", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing PoolGroup by filters like name, cloud, tenant
// Api creates PoolGroup object with every call.
func (client *PoolGroupClient) GetObject(options ...session.ApiOptionsParams) (*models.PoolGroup, error) {
	var obj *models.PoolGroup
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("poolgroup", newOptions...)
	return obj, err
}

// Create a new PoolGroup object
func (client *PoolGroupClient) Create(obj *models.PoolGroup, options ...session.ApiOptionsParams) (*models.PoolGroup, error) {
	var robj *models.PoolGroup
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing PoolGroup object
func (client *PoolGroupClient) Update(obj *models.PoolGroup, options ...session.ApiOptionsParams) (*models.PoolGroup, error) {
	var robj *models.PoolGroup
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing PoolGroup object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.PoolGroup
// or it should be json compatible of form map[string]interface{}
func (client *PoolGroupClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.PoolGroup, error) {
	var robj *models.PoolGroup
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing PoolGroup object with a given UUID
func (client *PoolGroupClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing PoolGroup object with a given name
func (client *PoolGroupClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *PoolGroupClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
