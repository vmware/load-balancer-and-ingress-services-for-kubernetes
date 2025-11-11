// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// ALBServicesJobClient is a client for avi ALBServicesJob resource
type ALBServicesJobClient struct {
	aviSession *session.AviSession
}

// NewALBServicesJobClient creates a new client for ALBServicesJob resource
func NewALBServicesJobClient(aviSession *session.AviSession) *ALBServicesJobClient {
	return &ALBServicesJobClient{aviSession: aviSession}
}

func (client *ALBServicesJobClient) getAPIPath(uuid string) string {
	path := "api/albservicesjob"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of ALBServicesJob objects
func (client *ALBServicesJobClient) GetAll(options ...session.ApiOptionsParams) ([]*models.ALBServicesJob, error) {
	var plist []*models.ALBServicesJob
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing ALBServicesJob by uuid
func (client *ALBServicesJobClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.ALBServicesJob, error) {
	var obj *models.ALBServicesJob
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing ALBServicesJob by name
func (client *ALBServicesJobClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.ALBServicesJob, error) {
	var obj *models.ALBServicesJob
	err := client.aviSession.GetObjectByName("albservicesjob", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing ALBServicesJob by filters like name, cloud, tenant
// Api creates ALBServicesJob object with every call.
func (client *ALBServicesJobClient) GetObject(options ...session.ApiOptionsParams) (*models.ALBServicesJob, error) {
	var obj *models.ALBServicesJob
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("albservicesjob", newOptions...)
	return obj, err
}

// Create a new ALBServicesJob object
func (client *ALBServicesJobClient) Create(obj *models.ALBServicesJob, options ...session.ApiOptionsParams) (*models.ALBServicesJob, error) {
	var robj *models.ALBServicesJob
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing ALBServicesJob object
func (client *ALBServicesJobClient) Update(obj *models.ALBServicesJob, options ...session.ApiOptionsParams) (*models.ALBServicesJob, error) {
	var robj *models.ALBServicesJob
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing ALBServicesJob object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.ALBServicesJob
// or it should be json compatible of form map[string]interface{}
func (client *ALBServicesJobClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.ALBServicesJob, error) {
	var robj *models.ALBServicesJob
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing ALBServicesJob object with a given UUID
func (client *ALBServicesJobClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing ALBServicesJob object with a given name
func (client *ALBServicesJobClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *ALBServicesJobClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
