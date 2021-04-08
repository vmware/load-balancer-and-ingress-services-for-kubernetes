/***************************************************************************
 * ------------------------------------------------------------------------
 * Copyright 2020 VMware, Inc.  All rights reserved. VMware Confidential
 * ------------------------------------------------------------------------
 */

package clients

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
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
func (client *JobEntryClient) GetAll(options ...session.ApiOptionsParams) ([]*models.JobEntry, error) {
	var plist []*models.JobEntry
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing JobEntry by uuid
func (client *JobEntryClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.JobEntry, error) {
	var obj *models.JobEntry
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing JobEntry by name
func (client *JobEntryClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.JobEntry, error) {
	var obj *models.JobEntry
	err := client.aviSession.GetObjectByName("jobentry", name, &obj, options...)
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
func (client *JobEntryClient) Create(obj *models.JobEntry, options ...session.ApiOptionsParams) (*models.JobEntry, error) {
	var robj *models.JobEntry
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing JobEntry object
func (client *JobEntryClient) Update(obj *models.JobEntry, options ...session.ApiOptionsParams) (*models.JobEntry, error) {
	var robj *models.JobEntry
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing JobEntry object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.JobEntry
// or it should be json compatible of form map[string]interface{}
func (client *JobEntryClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.JobEntry, error) {
	var robj *models.JobEntry
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing JobEntry object with a given UUID
func (client *JobEntryClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing JobEntry object with a given name
func (client *JobEntryClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *JobEntryClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
