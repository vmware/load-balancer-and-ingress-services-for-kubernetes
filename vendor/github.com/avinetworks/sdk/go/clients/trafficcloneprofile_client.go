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

// TrafficCloneProfileClient is a client for avi TrafficCloneProfile resource
type TrafficCloneProfileClient struct {
	aviSession *session.AviSession
}

// NewTrafficCloneProfileClient creates a new client for TrafficCloneProfile resource
func NewTrafficCloneProfileClient(aviSession *session.AviSession) *TrafficCloneProfileClient {
	return &TrafficCloneProfileClient{aviSession: aviSession}
}

func (client *TrafficCloneProfileClient) getAPIPath(uuid string) string {
	path := "api/trafficcloneprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of TrafficCloneProfile objects
func (client *TrafficCloneProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.TrafficCloneProfile, error) {
	var plist []*models.TrafficCloneProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing TrafficCloneProfile by uuid
func (client *TrafficCloneProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.TrafficCloneProfile, error) {
	var obj *models.TrafficCloneProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing TrafficCloneProfile by name
func (client *TrafficCloneProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.TrafficCloneProfile, error) {
	var obj *models.TrafficCloneProfile
	err := client.aviSession.GetObjectByName("trafficcloneprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing TrafficCloneProfile by filters like name, cloud, tenant
// Api creates TrafficCloneProfile object with every call.
func (client *TrafficCloneProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.TrafficCloneProfile, error) {
	var obj *models.TrafficCloneProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("trafficcloneprofile", newOptions...)
	return obj, err
}

// Create a new TrafficCloneProfile object
func (client *TrafficCloneProfileClient) Create(obj *models.TrafficCloneProfile, options ...session.ApiOptionsParams) (*models.TrafficCloneProfile, error) {
	var robj *models.TrafficCloneProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing TrafficCloneProfile object
func (client *TrafficCloneProfileClient) Update(obj *models.TrafficCloneProfile, options ...session.ApiOptionsParams) (*models.TrafficCloneProfile, error) {
	var robj *models.TrafficCloneProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing TrafficCloneProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.TrafficCloneProfile
// or it should be json compatible of form map[string]interface{}
func (client *TrafficCloneProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.TrafficCloneProfile, error) {
	var robj *models.TrafficCloneProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing TrafficCloneProfile object with a given UUID
func (client *TrafficCloneProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing TrafficCloneProfile object with a given name
func (client *TrafficCloneProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *TrafficCloneProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
