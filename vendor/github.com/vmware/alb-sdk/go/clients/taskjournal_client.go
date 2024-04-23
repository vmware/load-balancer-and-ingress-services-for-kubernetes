// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// TaskJournalClient is a client for avi TaskJournal resource
type TaskJournalClient struct {
	aviSession *session.AviSession
}

// NewTaskJournalClient creates a new client for TaskJournal resource
func NewTaskJournalClient(aviSession *session.AviSession) *TaskJournalClient {
	return &TaskJournalClient{aviSession: aviSession}
}

func (client *TaskJournalClient) getAPIPath(uuid string) string {
	path := "api/taskjournal"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of TaskJournal objects
func (client *TaskJournalClient) GetAll(options ...session.ApiOptionsParams) ([]*models.TaskJournal, error) {
	var plist []*models.TaskJournal
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing TaskJournal by uuid
func (client *TaskJournalClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.TaskJournal, error) {
	var obj *models.TaskJournal
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing TaskJournal by name
func (client *TaskJournalClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.TaskJournal, error) {
	var obj *models.TaskJournal
	err := client.aviSession.GetObjectByName("taskjournal", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing TaskJournal by filters like name, cloud, tenant
// Api creates TaskJournal object with every call.
func (client *TaskJournalClient) GetObject(options ...session.ApiOptionsParams) (*models.TaskJournal, error) {
	var obj *models.TaskJournal
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("taskjournal", newOptions...)
	return obj, err
}

// Create a new TaskJournal object
func (client *TaskJournalClient) Create(obj *models.TaskJournal, options ...session.ApiOptionsParams) (*models.TaskJournal, error) {
	var robj *models.TaskJournal
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing TaskJournal object
func (client *TaskJournalClient) Update(obj *models.TaskJournal, options ...session.ApiOptionsParams) (*models.TaskJournal, error) {
	var robj *models.TaskJournal
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing TaskJournal object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.TaskJournal
// or it should be json compatible of form map[string]interface{}
func (client *TaskJournalClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.TaskJournal, error) {
	var robj *models.TaskJournal
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing TaskJournal object with a given UUID
func (client *TaskJournalClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing TaskJournal object with a given name
func (client *TaskJournalClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *TaskJournalClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
