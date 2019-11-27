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

// SnmpTrapProfileClient is a client for avi SnmpTrapProfile resource
type SnmpTrapProfileClient struct {
	aviSession *session.AviSession
}

// NewSnmpTrapProfileClient creates a new client for SnmpTrapProfile resource
func NewSnmpTrapProfileClient(aviSession *session.AviSession) *SnmpTrapProfileClient {
	return &SnmpTrapProfileClient{aviSession: aviSession}
}

func (client *SnmpTrapProfileClient) getAPIPath(uuid string) string {
	path := "api/snmptrapprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of SnmpTrapProfile objects
func (client *SnmpTrapProfileClient) GetAll() ([]*models.SnmpTrapProfile, error) {
	var plist []*models.SnmpTrapProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing SnmpTrapProfile by uuid
func (client *SnmpTrapProfileClient) Get(uuid string) (*models.SnmpTrapProfile, error) {
	var obj *models.SnmpTrapProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing SnmpTrapProfile by name
func (client *SnmpTrapProfileClient) GetByName(name string) (*models.SnmpTrapProfile, error) {
	var obj *models.SnmpTrapProfile
	err := client.aviSession.GetObjectByName("snmptrapprofile", name, &obj)
	return obj, err
}

// GetObject - Get an existing SnmpTrapProfile by filters like name, cloud, tenant
// Api creates SnmpTrapProfile object with every call.
func (client *SnmpTrapProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.SnmpTrapProfile, error) {
	var obj *models.SnmpTrapProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("snmptrapprofile", newOptions...)
	return obj, err
}

// Create a new SnmpTrapProfile object
func (client *SnmpTrapProfileClient) Create(obj *models.SnmpTrapProfile) (*models.SnmpTrapProfile, error) {
	var robj *models.SnmpTrapProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing SnmpTrapProfile object
func (client *SnmpTrapProfileClient) Update(obj *models.SnmpTrapProfile) (*models.SnmpTrapProfile, error) {
	var robj *models.SnmpTrapProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing SnmpTrapProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.SnmpTrapProfile
// or it should be json compatible of form map[string]interface{}
func (client *SnmpTrapProfileClient) Patch(uuid string, patch interface{}, patchOp string) (*models.SnmpTrapProfile, error) {
	var robj *models.SnmpTrapProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing SnmpTrapProfile object with a given UUID
func (client *SnmpTrapProfileClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing SnmpTrapProfile object with a given name
func (client *SnmpTrapProfileClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *SnmpTrapProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
