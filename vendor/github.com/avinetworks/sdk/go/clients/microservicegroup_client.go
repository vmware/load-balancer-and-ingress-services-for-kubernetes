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
func (client *MicroServiceGroupClient) GetAll() ([]*models.MicroServiceGroup, error) {
	var plist []*models.MicroServiceGroup
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing MicroServiceGroup by uuid
func (client *MicroServiceGroupClient) Get(uuid string) (*models.MicroServiceGroup, error) {
	var obj *models.MicroServiceGroup
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing MicroServiceGroup by name
func (client *MicroServiceGroupClient) GetByName(name string) (*models.MicroServiceGroup, error) {
	var obj *models.MicroServiceGroup
	err := client.aviSession.GetObjectByName("microservicegroup", name, &obj)
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
func (client *MicroServiceGroupClient) Create(obj *models.MicroServiceGroup) (*models.MicroServiceGroup, error) {
	var robj *models.MicroServiceGroup
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing MicroServiceGroup object
func (client *MicroServiceGroupClient) Update(obj *models.MicroServiceGroup) (*models.MicroServiceGroup, error) {
	var robj *models.MicroServiceGroup
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing MicroServiceGroup object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.MicroServiceGroup
// or it should be json compatible of form map[string]interface{}
func (client *MicroServiceGroupClient) Patch(uuid string, patch interface{}, patchOp string) (*models.MicroServiceGroup, error) {
	var robj *models.MicroServiceGroup
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing MicroServiceGroup object with a given UUID
func (client *MicroServiceGroupClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing MicroServiceGroup object with a given name
func (client *MicroServiceGroupClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *MicroServiceGroupClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
