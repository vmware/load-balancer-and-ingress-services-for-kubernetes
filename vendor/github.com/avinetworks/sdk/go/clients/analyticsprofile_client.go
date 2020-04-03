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

// AnalyticsProfileClient is a client for avi AnalyticsProfile resource
type AnalyticsProfileClient struct {
	aviSession *session.AviSession
}

// NewAnalyticsProfileClient creates a new client for AnalyticsProfile resource
func NewAnalyticsProfileClient(aviSession *session.AviSession) *AnalyticsProfileClient {
	return &AnalyticsProfileClient{aviSession: aviSession}
}

func (client *AnalyticsProfileClient) getAPIPath(uuid string) string {
	path := "api/analyticsprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of AnalyticsProfile objects
func (client *AnalyticsProfileClient) GetAll(options ...session.ApiOptionsParams) ([]*models.AnalyticsProfile, error) {
	var plist []*models.AnalyticsProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing AnalyticsProfile by uuid
func (client *AnalyticsProfileClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.AnalyticsProfile, error) {
	var obj *models.AnalyticsProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing AnalyticsProfile by name
func (client *AnalyticsProfileClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.AnalyticsProfile, error) {
	var obj *models.AnalyticsProfile
	err := client.aviSession.GetObjectByName("analyticsprofile", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing AnalyticsProfile by filters like name, cloud, tenant
// Api creates AnalyticsProfile object with every call.
func (client *AnalyticsProfileClient) GetObject(options ...session.ApiOptionsParams) (*models.AnalyticsProfile, error) {
	var obj *models.AnalyticsProfile
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("analyticsprofile", newOptions...)
	return obj, err
}

// Create a new AnalyticsProfile object
func (client *AnalyticsProfileClient) Create(obj *models.AnalyticsProfile, options ...session.ApiOptionsParams) (*models.AnalyticsProfile, error) {
	var robj *models.AnalyticsProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing AnalyticsProfile object
func (client *AnalyticsProfileClient) Update(obj *models.AnalyticsProfile, options ...session.ApiOptionsParams) (*models.AnalyticsProfile, error) {
	var robj *models.AnalyticsProfile
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing AnalyticsProfile object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.AnalyticsProfile
// or it should be json compatible of form map[string]interface{}
func (client *AnalyticsProfileClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.AnalyticsProfile, error) {
	var robj *models.AnalyticsProfile
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing AnalyticsProfile object with a given UUID
func (client *AnalyticsProfileClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing AnalyticsProfile object with a given name
func (client *AnalyticsProfileClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *AnalyticsProfileClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
