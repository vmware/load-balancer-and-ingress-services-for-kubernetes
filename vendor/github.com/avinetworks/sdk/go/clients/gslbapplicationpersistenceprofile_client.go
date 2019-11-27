package clients

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

import (
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
)

// GslbApplicationPersistenceProfileClient is a client for avi GslbApplicationPersistenceProfile resource
type GslbApplicationPersistenceProfileClient struct {
	aviSession *session.AviSession
}

// NewGslbApplicationPersistenceProfileClient creates a new client for GslbApplicationPersistenceProfile resource
func NewGslbApplicationPersistenceProfileClient(aviSession *session.AviSession) *GslbApplicationPersistenceProfileClient {
	return &GslbApplicationPersistenceProfileClient{aviSession: aviSession}
}

func (client *GslbApplicationPersistenceProfileClient) getAPIPath(uuid string) string {
	path := "api/gslbapplicationpersistenceprofile"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of GslbApplicationPersistenceProfile objects
func (client *GslbApplicationPersistenceProfileClient) GetAll() ([]*models.GslbApplicationPersistenceProfile, error) {
	var plist []*models.GslbApplicationPersistenceProfile
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing GslbApplicationPersistenceProfile by uuid
func (client *GslbApplicationPersistenceProfileClient) Get(uuid string) (*models.GslbApplicationPersistenceProfile, error) {
	var obj *models.GslbApplicationPersistenceProfile
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing GslbApplicationPersistenceProfile by name
func (client *GslbApplicationPersistenceProfileClient) GetByName(name string) (*models.GslbApplicationPersistenceProfile, error) {
	var obj *models.GslbApplicationPersistenceProfile
	err := client.aviSession.GetObjectByName("gslbapplicationpersistenceprofile", name, &obj)
	return obj, err
}

// Create a new GslbApplicationPersistenceProfile object
func (client *GslbApplicationPersistenceProfileClient) Create(obj *models.GslbApplicationPersistenceProfile) (*models.GslbApplicationPersistenceProfile, error) {
	var robj *models.GslbApplicationPersistenceProfile
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing GslbApplicationPersistenceProfile object
func (client *GslbApplicationPersistenceProfileClient) Update(obj *models.GslbApplicationPersistenceProfile) (*models.GslbApplicationPersistenceProfile, error) {
	var robj *models.GslbApplicationPersistenceProfile
	path := client.getAPIPath(obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Delete an existing GslbApplicationPersistenceProfile object with a given UUID
func (client *GslbApplicationPersistenceProfileClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing GslbApplicationPersistenceProfile object with a given name
func (client *GslbApplicationPersistenceProfileClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(res.UUID)
}
