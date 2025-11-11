// Copyright 2019 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// BotDetectionPolicyClient is a client for avi BotDetectionPolicy resource
type BotDetectionPolicyClient struct {
	aviSession *session.AviSession
}

// NewBotDetectionPolicyClient creates a new client for BotDetectionPolicy resource
func NewBotDetectionPolicyClient(aviSession *session.AviSession) *BotDetectionPolicyClient {
	return &BotDetectionPolicyClient{aviSession: aviSession}
}

func (client *BotDetectionPolicyClient) getAPIPath(uuid string) string {
	path := "api/botdetectionpolicy"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of BotDetectionPolicy objects
func (client *BotDetectionPolicyClient) GetAll(options ...session.ApiOptionsParams) ([]*models.BotDetectionPolicy, error) {
	var plist []*models.BotDetectionPolicy
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing BotDetectionPolicy by uuid
func (client *BotDetectionPolicyClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.BotDetectionPolicy, error) {
	var obj *models.BotDetectionPolicy
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing BotDetectionPolicy by name
func (client *BotDetectionPolicyClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.BotDetectionPolicy, error) {
	var obj *models.BotDetectionPolicy
	err := client.aviSession.GetObjectByName("botdetectionpolicy", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing BotDetectionPolicy by filters like name, cloud, tenant
// Api creates BotDetectionPolicy object with every call.
func (client *BotDetectionPolicyClient) GetObject(options ...session.ApiOptionsParams) (*models.BotDetectionPolicy, error) {
	var obj *models.BotDetectionPolicy
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("botdetectionpolicy", newOptions...)
	return obj, err
}

// Create a new BotDetectionPolicy object
func (client *BotDetectionPolicyClient) Create(obj *models.BotDetectionPolicy, options ...session.ApiOptionsParams) (*models.BotDetectionPolicy, error) {
	var robj *models.BotDetectionPolicy
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing BotDetectionPolicy object
func (client *BotDetectionPolicyClient) Update(obj *models.BotDetectionPolicy, options ...session.ApiOptionsParams) (*models.BotDetectionPolicy, error) {
	var robj *models.BotDetectionPolicy
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing BotDetectionPolicy object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.BotDetectionPolicy
// or it should be json compatible of form map[string]interface{}
func (client *BotDetectionPolicyClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.BotDetectionPolicy, error) {
	var robj *models.BotDetectionPolicy
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing BotDetectionPolicy object with a given UUID
func (client *BotDetectionPolicyClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing BotDetectionPolicy object with a given name
func (client *BotDetectionPolicyClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *BotDetectionPolicyClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
