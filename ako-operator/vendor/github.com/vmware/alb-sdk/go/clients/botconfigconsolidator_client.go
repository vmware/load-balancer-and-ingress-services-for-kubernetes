// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

// This file is auto-generated.

import (
	"github.com/vmware/alb-sdk/go/models"
	"github.com/vmware/alb-sdk/go/session"
)

// BotConfigConsolidatorClient is a client for avi BotConfigConsolidator resource
type BotConfigConsolidatorClient struct {
	aviSession *session.AviSession
}

// NewBotConfigConsolidatorClient creates a new client for BotConfigConsolidator resource
func NewBotConfigConsolidatorClient(aviSession *session.AviSession) *BotConfigConsolidatorClient {
	return &BotConfigConsolidatorClient{aviSession: aviSession}
}

func (client *BotConfigConsolidatorClient) getAPIPath(uuid string) string {
	path := "api/botconfigconsolidator"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of BotConfigConsolidator objects
func (client *BotConfigConsolidatorClient) GetAll(options ...session.ApiOptionsParams) ([]*models.BotConfigConsolidator, error) {
	var plist []*models.BotConfigConsolidator
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist, options...)
	return plist, err
}

// Get an existing BotConfigConsolidator by uuid
func (client *BotConfigConsolidatorClient) Get(uuid string, options ...session.ApiOptionsParams) (*models.BotConfigConsolidator, error) {
	var obj *models.BotConfigConsolidator
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj, options...)
	return obj, err
}

// GetByName - Get an existing BotConfigConsolidator by name
func (client *BotConfigConsolidatorClient) GetByName(name string, options ...session.ApiOptionsParams) (*models.BotConfigConsolidator, error) {
	var obj *models.BotConfigConsolidator
	err := client.aviSession.GetObjectByName("botconfigconsolidator", name, &obj, options...)
	return obj, err
}

// GetObject - Get an existing BotConfigConsolidator by filters like name, cloud, tenant
// Api creates BotConfigConsolidator object with every call.
func (client *BotConfigConsolidatorClient) GetObject(options ...session.ApiOptionsParams) (*models.BotConfigConsolidator, error) {
	var obj *models.BotConfigConsolidator
	newOptions := make([]session.ApiOptionsParams, len(options)+1)
	for i, p := range options {
		newOptions[i] = p
	}
	newOptions[len(options)] = session.SetResult(&obj)
	err := client.aviSession.GetObject("botconfigconsolidator", newOptions...)
	return obj, err
}

// Create a new BotConfigConsolidator object
func (client *BotConfigConsolidatorClient) Create(obj *models.BotConfigConsolidator, options ...session.ApiOptionsParams) (*models.BotConfigConsolidator, error) {
	var robj *models.BotConfigConsolidator
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj, options...)
	return robj, err
}

// Update an existing BotConfigConsolidator object
func (client *BotConfigConsolidatorClient) Update(obj *models.BotConfigConsolidator, options ...session.ApiOptionsParams) (*models.BotConfigConsolidator, error) {
	var robj *models.BotConfigConsolidator
	path := client.getAPIPath(*obj.UUID)
	err := client.aviSession.Put(path, obj, &robj, options...)
	return robj, err
}

// Patch an existing BotConfigConsolidator object specified using uuid
// patchOp: Patch operation - add, replace, or delete
// patch: Patch payload should be compatible with the models.BotConfigConsolidator
// or it should be json compatible of form map[string]interface{}
func (client *BotConfigConsolidatorClient) Patch(uuid string, patch interface{}, patchOp string, options ...session.ApiOptionsParams) (*models.BotConfigConsolidator, error) {
	var robj *models.BotConfigConsolidator
	path := client.getAPIPath(uuid)
	err := client.aviSession.Patch(path, patch, patchOp, &robj, options...)
	return robj, err
}

// Delete an existing BotConfigConsolidator object with a given UUID
func (client *BotConfigConsolidatorClient) Delete(uuid string, options ...session.ApiOptionsParams) error {
	if len(options) == 0 {
		return client.aviSession.Delete(client.getAPIPath(uuid))
	} else {
		return client.aviSession.DeleteObject(client.getAPIPath(uuid), options...)
	}
}

// DeleteByName - Delete an existing BotConfigConsolidator object with a given name
func (client *BotConfigConsolidatorClient) DeleteByName(name string, options ...session.ApiOptionsParams) error {
	res, err := client.GetByName(name, options...)
	if err != nil {
		return err
	}
	return client.Delete(*res.UUID, options...)
}

// GetAviSession
func (client *BotConfigConsolidatorClient) GetAviSession() *session.AviSession {
	return client.aviSession
}
