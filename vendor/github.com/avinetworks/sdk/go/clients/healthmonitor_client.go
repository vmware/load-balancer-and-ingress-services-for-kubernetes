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

// HealthMonitorClient is a client for avi HealthMonitor resource
type HealthMonitorClient struct {
	aviSession *session.AviSession
}

// NewHealthMonitorClient creates a new client for HealthMonitor resource
func NewHealthMonitorClient(aviSession *session.AviSession) *HealthMonitorClient {
	return &HealthMonitorClient{aviSession: aviSession}
}

func (client *HealthMonitorClient) getAPIPath(uuid string) string {
	path := "api/healthmonitor"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of HealthMonitor objects
func (client *HealthMonitorClient) GetAll() ([]*models.HealthMonitor, error) {
	var plist []*models.HealthMonitor
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing HealthMonitor by uuid
func (client *HealthMonitorClient) Get(uuid string) (*models.HealthMonitor, error) {
	var obj *models.HealthMonitor
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing HealthMonitor by name
func (client *HealthMonitorClient) GetByName(name string) (*models.HealthMonitor, error) {
	var obj *models.HealthMonitor
	err := client.aviSession.GetObjectByName("healthmonitor", name, &obj)
	return obj, err
}

// GetObject - Get an existing HealthMonitor by filters like name, cloud, tenant
// Api creates HealthMonitor object with every call.
func (client *HealthMonitorClient) GetObject(options ...session.ApiOptionsParams) (*models.HealthMonitor, error) {
	var obj *models.HealthMonitor
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("healthmonitor", newOptions...)
	return obj, err
}

// Create a new HealthMonitor object
func (client *HealthMonitorClient) Create(obj *models.HealthMonitor) (*models.HealthMonitor, error) {
	var robj *models.HealthMonitor
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing HealthMonitor object
func (client *HealthMonitorClient) Update(obj *models.HealthMonitor) (*models.HealthMonitor, error) {
	var robj *models.HealthMonitor
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing HealthMonitor object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.HealthMonitor
// or it should be json compatible of form map[string]interface{}
func (client *HealthMonitorClient) Patch(uuid string, patch interface{}, patchOp string) (*models.HealthMonitor, error) {
	var robj *models.HealthMonitor
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing HealthMonitor object with a given UUID
func (client *HealthMonitorClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing HealthMonitor object with a given name
func (client *HealthMonitorClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *HealthMonitorClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
