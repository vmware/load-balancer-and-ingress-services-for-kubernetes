package clients

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

import (
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
)

// GslbHealthMonitorClient is a client for avi GslbHealthMonitor resource
type GslbHealthMonitorClient struct {
	aviSession *session.AviSession
}

// NewGslbHealthMonitorClient creates a new client for GslbHealthMonitor resource
func NewGslbHealthMonitorClient(aviSession *session.AviSession) *GslbHealthMonitorClient {
	return &GslbHealthMonitorClient{aviSession: aviSession}
}

func (client *GslbHealthMonitorClient) getAPIPath(uuid string) string {
	path := "api/gslbhealthmonitor"
	if uuid != "" {
		path += "/" + uuid
	}
	return path
}

// GetAll is a collection API to get a list of GslbHealthMonitor objects
func (client *GslbHealthMonitorClient) GetAll() ([]*models.GslbHealthMonitor, error) {
	var plist []*models.GslbHealthMonitor
	err := client.aviSession.GetCollection(client.getAPIPath(""), &plist)
	return plist, err
}

// Get an existing GslbHealthMonitor by uuid
func (client *GslbHealthMonitorClient) Get(uuid string) (*models.GslbHealthMonitor, error) {
	var obj *models.GslbHealthMonitor
	err := client.aviSession.Get(client.getAPIPath(uuid), &obj)
	return obj, err
}

// GetByName - Get an existing GslbHealthMonitor by name
func (client *GslbHealthMonitorClient) GetByName(name string) (*models.GslbHealthMonitor, error) {
	var obj *models.GslbHealthMonitor
	err := client.aviSession.GetObjectByName("gslbhealthmonitor", name, &obj)
	return obj, err
}

// Create a new GslbHealthMonitor object
func (client *GslbHealthMonitorClient) Create(obj *models.GslbHealthMonitor) (*models.GslbHealthMonitor, error) {
	var robj *models.GslbHealthMonitor
	err := client.aviSession.Post(client.getAPIPath(""), obj, &robj)
	return robj, err
}

// Update an existing GslbHealthMonitor object
func (client *GslbHealthMonitorClient) Update(obj *models.GslbHealthMonitor) (*models.GslbHealthMonitor, error) {
	var robj *models.GslbHealthMonitor
	path := client.getAPIPath(obj.UUID)
	err := client.aviSession.Put(path, obj, &robj)
	return robj, err
}

// Delete an existing GslbHealthMonitor object with a given UUID
func (client *GslbHealthMonitorClient) Delete(uuid string) error {
	return client.aviSession.Delete(client.getAPIPath(uuid))
}

// DeleteByName - Delete an existing GslbHealthMonitor object with a given name
func (client *GslbHealthMonitorClient) DeleteByName(name string) error {
	res, err := client.GetByName(name)
	if err != nil {
		return err
	}
	return client.Delete(res.UUID)
}
