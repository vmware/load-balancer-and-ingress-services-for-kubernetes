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

// MicroServiceGroupClient is a client for avi MicroServiceGroup resource
type MicroServiceGroupClient struct {
	aviSession *session.AviSession
}

// NewMicroServiceGroupClient creates a new client for MicroServiceGroup resource
func NewMicroServiceGroupClient(aviSession *session.AviSession) *MicroServiceGroupClient {
	return &MicroServiceGroupClient{aviSession: aviSession}
}

func (client *MicroServiceGroupClient) getAPIPath(uuid string) string {
	path := "api/microservicegroup"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of MicroServiceGroup objects
func (client *MicroServiceGroupClient) GetAll(options ...session.ApiOptionsParams) ([]*models.MicroServiceGroup, error) {
	var plist []*models.MicroServiceGroup
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing MicroServiceGroup by uuid
func (client *MicroServiceGroupClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.MicroServiceGroup, error) {
	var obj *models.MicroServiceGroup
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing MicroServiceGroup by name
func (client *MicroServiceGroupClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.MicroServiceGroup, error) {
	var obj *models.MicroServiceGroup
	err := client.aviSession.GetObjectByName("microservicegroup", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing MicroServiceGroup by filters like name, cloud, tenant
// Api creates MicroServiceGroup object with every call.
func (client *MicroServiceGroupClient) GetObject(options ...session.ApiOptionsParams) (*models.MicroServiceGroup, error) {
	var obj *models.MicroServiceGroup
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("microservicegroup", newOptions...)
	return obj, err
}

// Create a new MicroServiceGroup object
func (client *MicroServiceGroupClient) Create(obj *models.MicroServiceGroup, options ...session.ApiOptionsParams) (*models.MicroServiceGroup, error) {
	var robj *models.MicroServiceGroup
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing MicroServiceGroup object
func (client *MicroServiceGroupClient) Update(obj *models.MicroServiceGroup, options ...session.ApiOptionsParams) (*models.MicroServiceGroup, error) {
	var robj *models.MicroServiceGroup
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing MicroServiceGroup object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.MicroServiceGroup
// or it should be json compatible of form map[string]interface{}
func (client *MicroServiceGroupClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.MicroServiceGroup, error) {
	var robj *models.MicroServiceGroup
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing MicroServiceGroup object with a given UUID
func (client *MicroServiceGroupClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing MicroServiceGroup object with a given name
func (client *MicroServiceGroupClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *MicroServiceGroupClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
