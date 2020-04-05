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

// SCPoolServerStateInfoClient is a client for avi SCPoolServerStateInfo resource
type SCPoolServerStateInfoClient struct {
	aviSession *session.AviSession
}

// NewSCPoolServerStateInfoClient creates a new client for SCPoolServerStateInfo resource
func NewSCPoolServerStateInfoClient(aviSession *session.AviSession) *SCPoolServerStateInfoClient {
	return &SCPoolServerStateInfoClient{aviSession: aviSession}
}

func (client *SCPoolServerStateInfoClient) getAPIPath(uuid string) string {
	path := "api/scpoolserverstateinfo"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of SCPoolServerStateInfo objects
func (client *SCPoolServerStateInfoClient) GetAll(options ...session.ApiOptionsParams) ([]*models.SCPoolServerStateInfo, error) {
	var plist []*models.SCPoolServerStateInfo
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing SCPoolServerStateInfo by uuid
func (client *SCPoolServerStateInfoClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.SCPoolServerStateInfo, error) {
	var obj *models.SCPoolServerStateInfo
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing SCPoolServerStateInfo by name
func (client *SCPoolServerStateInfoClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.SCPoolServerStateInfo, error) {
	var obj *models.SCPoolServerStateInfo
	err := client.aviSession.GetObjectByName("scpoolserverstateinfo", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing SCPoolServerStateInfo by filters like name, cloud, tenant
// Api creates SCPoolServerStateInfo object with every call.
func (client *SCPoolServerStateInfoClient) GetObject(options ...session.ApiOptionsParams) (*models.SCPoolServerStateInfo, error) {
	var obj *models.SCPoolServerStateInfo
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("scpoolserverstateinfo", newOptions...)
	return obj, err
}

// Create a new SCPoolServerStateInfo object
func (client *SCPoolServerStateInfoClient) Create(obj *models.SCPoolServerStateInfo, options ...session.ApiOptionsParams) (*models.SCPoolServerStateInfo, error) {
	var robj *models.SCPoolServerStateInfo
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing SCPoolServerStateInfo object
func (client *SCPoolServerStateInfoClient) Update(obj *models.SCPoolServerStateInfo, options ...session.ApiOptionsParams) (*models.SCPoolServerStateInfo, error) {
	var robj *models.SCPoolServerStateInfo
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing SCPoolServerStateInfo object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.SCPoolServerStateInfo
// or it should be json compatible of form map[string]interface{}
func (client *SCPoolServerStateInfoClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.SCPoolServerStateInfo, error) {
	var robj *models.SCPoolServerStateInfo
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing SCPoolServerStateInfo object with a given UUID
func (client *SCPoolServerStateInfoClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing SCPoolServerStateInfo object with a given name
func (client *SCPoolServerStateInfoClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *SCPoolServerStateInfoClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
