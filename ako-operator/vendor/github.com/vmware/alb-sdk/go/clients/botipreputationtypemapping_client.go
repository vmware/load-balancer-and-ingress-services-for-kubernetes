// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// BotIPReputationTypeMappingClient is a client for avi BotIPReputationTypeMapping resource
type BotIPReputationTypeMappingClient struct {
	aviSession *session.AviSession
}

// NewBotIPReputationTypeMappingClient creates a new client for BotIPReputationTypeMapping resource
func NewBotIPReputationTypeMappingClient(aviSession *session.AviSession) *BotIPReputationTypeMappingClient {
	return &BotIPReputationTypeMappingClient{aviSession: aviSession}
}

func (client *BotIPReputationTypeMappingClient) getAPIPath(uuid string) string {
	path := "api/botipreputationtypemapping"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of BotIPReputationTypeMapping objects
func (client *BotIPReputationTypeMappingClient) GetAll(options ...session.ApiOptionsParams) ([]*models.BotIPReputationTypeMapping, error) {
	var plist []*models.BotIPReputationTypeMapping
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing BotIPReputationTypeMapping by uuid
func (client *BotIPReputationTypeMappingClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.BotIPReputationTypeMapping, error) {
	var obj *models.BotIPReputationTypeMapping
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing BotIPReputationTypeMapping by name
func (client *BotIPReputationTypeMappingClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.BotIPReputationTypeMapping, error) {
	var obj *models.BotIPReputationTypeMapping
	err := client.aviSession.GetObjectByName("botipreputationtypemapping", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing BotIPReputationTypeMapping by filters like name, cloud, tenant
// Api creates BotIPReputationTypeMapping object with every call.
func (client *BotIPReputationTypeMappingClient) GetObject(options ...session.ApiOptionsParams) (*models.BotIPReputationTypeMapping, error) {
	var obj *models.BotIPReputationTypeMapping
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("botipreputationtypemapping", newOptions...)
	return obj, err
}

// Create a new BotIPReputationTypeMapping object
func (client *BotIPReputationTypeMappingClient) Create(obj *models.BotIPReputationTypeMapping, options ...session.ApiOptionsParams) (*models.BotIPReputationTypeMapping, error) {
	var robj *models.BotIPReputationTypeMapping
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing BotIPReputationTypeMapping object
func (client *BotIPReputationTypeMappingClient) Update(obj *models.BotIPReputationTypeMapping, options ...session.ApiOptionsParams) (*models.BotIPReputationTypeMapping, error) {
	var robj *models.BotIPReputationTypeMapping
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing BotIPReputationTypeMapping object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.BotIPReputationTypeMapping
// or it should be json compatible of form map[string]interface{}
func (client *BotIPReputationTypeMappingClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.BotIPReputationTypeMapping, error) {
	var robj *models.BotIPReputationTypeMapping
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing BotIPReputationTypeMapping object with a given UUID
func (client *BotIPReputationTypeMappingClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing BotIPReputationTypeMapping object with a given name
func (client *BotIPReputationTypeMappingClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *BotIPReputationTypeMappingClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
