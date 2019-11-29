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

// PriorityLabelsClient is a client for avi PriorityLabels resource
type PriorityLabelsClient struct {
	aviSession *session.AviSession
}

// NewPriorityLabelsClient creates a new client for PriorityLabels resource
func NewPriorityLabelsClient(aviSession *session.AviSession) *PriorityLabelsClient {
	return &PriorityLabelsClient{aviSession: aviSession}
}

func (client *PriorityLabelsClient) getAPIPath(uuid string) string {
	path := "api/prioritylabels"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of PriorityLabels objects
func (client *PriorityLabelsClient) GetAll() ([]*models.PriorityLabels, error) {
	var plist []*models.PriorityLabels
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing PriorityLabels by uuid
func (client *PriorityLabelsClient) Get(uuid string) (*models.PriorityLabels, error) {
	var obj *models.PriorityLabels
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing PriorityLabels by name
func (client *PriorityLabelsClient) GetByName(name string) (*models.PriorityLabels, error) {
	var obj *models.PriorityLabels
	err := client.aviSession.GetObjectByName("prioritylabels", name, &obj)
	return obj, err
}

// GetObject - Get an existing PriorityLabels by filters like name, cloud, tenant
// Api creates PriorityLabels object with every call.
func (client *PriorityLabelsClient) GetObject(options ...session.ApiOptionsParams) (*models.PriorityLabels, error) {
	var obj *models.PriorityLabels
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("prioritylabels", newOptions...)
	return obj, err
}

// Create a new PriorityLabels object
func (client *PriorityLabelsClient) Create(obj *models.PriorityLabels) (*models.PriorityLabels, error) {
	var robj *models.PriorityLabels
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing PriorityLabels object
func (client *PriorityLabelsClient) Update(obj *models.PriorityLabels) (*models.PriorityLabels, error) {
	var robj *models.PriorityLabels
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing PriorityLabels object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.PriorityLabels
// or it should be json compatible of form map[string]interface{}
func (client *PriorityLabelsClient) Patch(uuid string, patch interface{}, patchOp string) (*models.PriorityLabels, error) {
	var robj *models.PriorityLabels
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing PriorityLabels object with a given UUID
func (client *PriorityLabelsClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing PriorityLabels object with a given name
func (client *PriorityLabelsClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *PriorityLabelsClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
