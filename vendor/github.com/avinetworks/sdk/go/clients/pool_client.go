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

// PoolClient is a client for avi Pool resource
type PoolClient struct {
	aviSession *session.AviSession
}

// NewPoolClient creates a new client for Pool resource
func NewPoolClient(aviSession *session.AviSession) *PoolClient {
	return &PoolClient{aviSession: aviSession}
}

func (client *PoolClient) getAPIPath(uuid string) string {
	path := "api/pool"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of Pool objects
func (client *PoolClient) GetAll() ([]*models.Pool, error) {
	var plist []*models.Pool
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing Pool by uuid
func (client *PoolClient) Get(uuid string) (*models.Pool, error) {
	var obj *models.Pool
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing Pool by name
func (client *PoolClient) GetByName(name string) (*models.Pool, error) {
	var obj *models.Pool
	err := client.aviSession.GetObjectByName("pool", name, &obj)
	return obj, err
}

// GetObject - Get an existing Pool by filters like name, cloud, tenant
// Api creates Pool object with every call.
func (client *PoolClient) GetObject(options ...session.ApiOptionsParams) (*models.Pool, error) {
	var obj *models.Pool
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("pool", newOptions...)
	return obj, err
}

// Create a new Pool object
func (client *PoolClient) Create(obj *models.Pool) (*models.Pool, error) {
	var robj *models.Pool
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing Pool object
func (client *PoolClient) Update(obj *models.Pool) (*models.Pool, error) {
	var robj *models.Pool
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing Pool object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.Pool
// or it should be json compatible of form map[string]interface{}
func (client *PoolClient) Patch(uuid string, patch interface{}, patchOp string) (*models.Pool, error) {
	var robj *models.Pool
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing Pool object with a given UUID
func (client *PoolClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing Pool object with a given name
func (client *PoolClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *PoolClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
