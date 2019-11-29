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

// JobEntryClient is a client for avi JobEntry resource
type JobEntryClient struct {
	aviSession *session.AviSession
}

// NewJobEntryClient creates a new client for JobEntry resource
func NewJobEntryClient(aviSession *session.AviSession) *JobEntryClient {
	return &JobEntryClient{aviSession: aviSession}
}

func (client *JobEntryClient) getAPIPath(uuid string) string {
	path := "api/jobentry"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of JobEntry objects
func (client *JobEntryClient) GetAll() ([]*models.JobEntry, error) {
	var plist []*models.JobEntry
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing JobEntry by uuid
func (client *JobEntryClient) Get(uuid string) (*models.JobEntry, error) {
	var obj *models.JobEntry
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing JobEntry by name
func (client *JobEntryClient) GetByName(name string) (*models.JobEntry, error) {
	var obj *models.JobEntry
	err := client.aviSession.GetObjectByName("jobentry", name, &obj)
	return obj, err
}

// GetObject - Get an existing JobEntry by filters like name, cloud, tenant
// Api creates JobEntry object with every call.
func (client *JobEntryClient) GetObject(options ...session.ApiOptionsParams) (*models.JobEntry, error) {
	var obj *models.JobEntry
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("jobentry", newOptions...)
	return obj, err
}

// Create a new JobEntry object
func (client *JobEntryClient) Create(obj *models.JobEntry) (*models.JobEntry, error) {
	var robj *models.JobEntry
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing JobEntry object
func (client *JobEntryClient) Update(obj *models.JobEntry) (*models.JobEntry, error) {
	var robj *models.JobEntry
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Patch an existing JobEntry object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.JobEntry
// or it should be json compatible of form map[string]interface{}
func (client *JobEntryClient) Patch(uuid string, patch interface{}, patchOp string) (*models.JobEntry, error) {
	var robj *models.JobEntry
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj)
	return robj, err
}

// Delete an existing JobEntry object with a given UUID
func (client *JobEntryClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing JobEntry object with a given name
func (client *JobEntryClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID)
}

// GetAviSession
func (client *JobEntryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
