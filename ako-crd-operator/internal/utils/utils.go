package utils

import "github.com/vmware/alb-sdk/go/models"

// createMarkers creates markers for the health monitor with cluster name and namespace
func CreateMarkers(clusterName, namespace string) []*models.RoleFilterMatchLabel {
	markers := []*models.RoleFilterMatchLabel{}

	// Add cluster name marker

	if clusterName != "" {
		clusterNameKey := "clustername"
		clusterMarker := &models.RoleFilterMatchLabel{
			Key:    &clusterNameKey,
			Values: []string{clusterName},
		}
		markers = append(markers, clusterMarker)
	}

	// Add namespace marker

	if namespace != "" {
		namespaceKey := "namespace"
		namespaceMarker := &models.RoleFilterMatchLabel{
			Key:    &namespaceKey,
			Values: []string{namespace},
		}
		markers = append(markers, namespaceMarker)
	}

	return markers
}
