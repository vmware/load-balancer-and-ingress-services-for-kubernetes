// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// MemoryBalancerRequestClient is a client for avi MemoryBalancerRequest resource
type MemoryBalancerRequestClient struct {
	aviSession *session.AviSession
}

// NewMemoryBalancerRequestClient creates a new client for MemoryBalancerRequest resource
func NewMemoryBalancerRequestClient(aviSession *session.AviSession) *MemoryBalancerRequestClient {
	return &MemoryBalancerRequestClient{aviSession: aviSession}
}

func (client *MemoryBalancerRequestClient) getAPIPath(uuid string) string {
	path := "api/memorybalancerrequest"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of MemoryBalancerRequest objects
func (client *MemoryBalancerRequestClient) GetAll(options ...session.ApiOptionsParams) ([]*models.MemoryBalancerRequest, error) {
	var plist []*models.MemoryBalancerRequest
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing MemoryBalancerRequest by uuid
func (client *MemoryBalancerRequestClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.MemoryBalancerRequest, error) {
	var obj *models.MemoryBalancerRequest
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing MemoryBalancerRequest by name
func (client *MemoryBalancerRequestClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.MemoryBalancerRequest, error) {
	var obj *models.MemoryBalancerRequest
	err := client.aviSession.GetObjectByName("memorybalancerrequest", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing MemoryBalancerRequest by filters like name, cloud, tenant
// Api creates MemoryBalancerRequest object with every call.
func (client *MemoryBalancerRequestClient) GetObject(options ...session.ApiOptionsParams) (*models.MemoryBalancerRequest, error) {
	var obj *models.MemoryBalancerRequest
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("memorybalancerrequest", newOptions...)
	return obj, err
}

// Create a new MemoryBalancerRequest object
func (client *MemoryBalancerRequestClient) Create(obj *models.MemoryBalancerRequest, options ...session.ApiOptionsParams) (*models.MemoryBalancerRequest, error) {
	var robj *models.MemoryBalancerRequest
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing MemoryBalancerRequest object
func (client *MemoryBalancerRequestClient) Update(obj *models.MemoryBalancerRequest, options ...session.ApiOptionsParams) (*models.MemoryBalancerRequest, error) {
	var robj *models.MemoryBalancerRequest
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing MemoryBalancerRequest object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.MemoryBalancerRequest
// or it should be json compatible of form map[string]interface{}
func (client *MemoryBalancerRequestClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.MemoryBalancerRequest, error) {
	var robj *models.MemoryBalancerRequest
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing MemoryBalancerRequest object with a given UUID
func (client *MemoryBalancerRequestClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing MemoryBalancerRequest object with a given name
func (client *MemoryBalancerRequestClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *MemoryBalancerRequestClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
