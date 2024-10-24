/*
 * Copyright 2021 VMware, Inc.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

// This file is used to create all Avi infra related changes and can be used as a library if required in other places.

package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/vmware/alb-sdk/go/models"
	avimodels "github.com/vmware/alb-sdk/go/models"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/session"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/avirest"
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

type AviControllerInfra struct {
	AviRestClient *clients.AviClient
	cs            kubernetes.Interface
}

var acceptedLicensesInAvi = []string{
	AVI_ENTERPRISE,
	AviEnterpriseWithCloudServices,
}

func NewAviControllerInfra(cs kubernetes.Interface) *AviControllerInfra {
	PopulateControllerProperties(cs)
	aviClient := avirest.InfraAviClientInstance()
	return &AviControllerInfra{AviRestClient: aviClient, cs: cs}
}

func (a *AviControllerInfra) InitInfraController() {
	if a.AviRestClient == nil {
		utils.AviLog.Fatalf("Avi client not initialized during Infra bootup")
	}

	if a.AviRestClient != nil && !avicache.IsAviClusterActive(a.AviRestClient) {
		utils.AviLog.Fatalf("Avi Controller Cluster state is not Active, shutting down AKO infa container")
	}

	// First verify the license of the Avi controller. If it's not Avi Enterprise, then fail the infra container bootup.
	err := a.VerifyAviControllerLicense()
	if err != nil {
		utils.AviLog.Fatalf(err.Error())
	}
}

func (a *AviControllerInfra) VerifyAviControllerLicense() error {
	uri := "/api/systemconfiguration"
	response := models.SystemConfiguration{}
	err := lib.AviGet(a.AviRestClient, uri, &response)
	if err != nil {
		utils.AviLog.Warnf("System config Get uri %v returned err %v", uri, err)
		return err
	}

	for _, license := range acceptedLicensesInAvi {
		if *response.DefaultLicenseTier == license {
			utils.AviLog.Infof("Avi Controller is running with %s license, proceeding with bootup", *response.DefaultLicenseTier)
			return nil
		}
	}

	return fmt.Errorf("Avi Controller license is not in accepted list %s. License tier is: %s", acceptedLicensesInAvi, *response.DefaultLicenseTier)
}

func (a *AviControllerInfra) checkVirtualService() (error, string) {
	createdBy := "ako-" + lib.GetClusterID()
	uri := "/api/virtualservice/?include_name&created_by=" + createdBy
	result, err := lib.AviGetCollectionRaw(a.AviRestClient, uri)
	if err != nil {
		utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
		return err, ""
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return err, ""
	}
	expName := lib.GetClusterID() + "--kube-system-kube-apiserver-lb-svc"
	for i := 0; i < len(elems); i++ {
		vs := models.VirtualService{}
		err = json.Unmarshal(elems[i], &vs)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal vs data, err: %v", err)
			continue
		}
		if vs.CloudRef != nil && strings.Contains(*vs.CloudRef, "#") {
			cloudName := strings.Split(*vs.CloudRef, "#")[1]
			if *vs.Name == expName {
				utils.AviLog.Infof("Found vs %s associated with cloud %s", expName, cloudName)
				return nil, cloudName
			}
		}
	}
	return nil, ""
}

func (a *AviControllerInfra) DeriveCloudNameAndSEGroupTmpl(tz string) (error, string, string) {
	// This method queries the Avi controller for all available cloud and then returns the cloud that matches the supplied transport zone
	uri := "/api/cloud/"
	_, cloudName := a.checkVirtualService()
	if cloudName != "" {
		uri = "/api/cloud/?include_name&name=" + cloudName
	}

	result, err := lib.AviGetCollectionRaw(a.AviRestClient, uri)
	if err != nil {
		utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
		return err, "", ""
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return err, "", ""
	}
	matchCloud := new(models.Cloud)
	for i := 0; i < len(elems); i++ {
		cloud := new(models.Cloud)
		err = json.Unmarshal(elems[i], &cloud)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal cloud data, err: %v", err)
			continue
		}
		if *cloud.Vtype != lib.CLOUD_NSXT || cloud.NsxtConfiguration == nil {
			continue
		}
		if lib.GetVPCMode() && (cloud.NsxtConfiguration.VpcMode == nil || !*cloud.NsxtConfiguration.VpcMode) {
			continue
		}
		if cloud.NsxtConfiguration.ManagementNetworkConfig == nil ||
			cloud.NsxtConfiguration.ManagementNetworkConfig.TransportZone == nil {
			continue
		}
		// In case of VPC mode, no need to match tranport zone as there would be only 1 cloud presnt in the avi controller
		if cloud.NsxtConfiguration.DataNetworkConfig == nil ||
			cloud.NsxtConfiguration.DataNetworkConfig.TransportZone == nil ||
			(!lib.GetVPCMode() && *cloud.NsxtConfiguration.DataNetworkConfig.TransportZone != tz) {
			continue
		}
		utils.AviLog.Infof("Found NSX-T cloud: %s match Transport Zone: %s", *cloud.Name, tz)
		matchCloud = cloud
		break
	}
	if matchCloud == nil {
		return errors.New("cloud not found matching transport zone " + tz), "", ""
	}
	if matchCloud.SeGroupTemplateRef != nil && *matchCloud.SeGroupTemplateRef != "" {
		tokenized := strings.Split(*matchCloud.SeGroupTemplateRef, "/api/serviceenginegroup/")
		if len(tokenized) == 2 {
			return nil, *matchCloud.Name, tokenized[1]
		}
	}

	// fetch Default-SEGroup uuid
	uri = "/api/serviceenginegroup/?include_name&cloud_ref.name=" + *matchCloud.Name + "&name=Default-Group"
	results, err := lib.AviGetCollectionRaw(a.AviRestClient, uri)
	if err != nil {
		utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
		return err, *matchCloud.Name, ""
	}

	elem := make([]json.RawMessage, results.Count)
	err = json.Unmarshal(results.Results, &elem)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return err, *matchCloud.Name, ""
	}
	if len(elem) == 0 {
		utils.AviLog.Errorf("No ServiceEngine Group with name Default-Group found.")
		return errors.New("No ServiceEngine Group with name Default-Group found."), *matchCloud.Name, ""
	}

	defaultSEG := models.ServiceEngineGroup{}
	err = json.Unmarshal(elems[0], &defaultSEG)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal cloud data, err: %v", err)
		return err, *matchCloud.Name, ""
	}
	return nil, *matchCloud.Name, *defaultSEG.UUID
}

func isPlacementScopeConfigured(configuredSEGroup *avimodels.ServiceEngineGroup) bool {
	configured := false
	for _, vc := range configuredSEGroup.Vcenters {
		if vc.NsxtClusters == nil {
			continue
		}
		configured = true
	}
	return configured
}

func (a *AviControllerInfra) SetupSEGroup(tz string) bool {
	err, cloudName, segTemplateUuid := a.DeriveCloudNameAndSEGroupTmpl(tz)
	if err != nil {
		lib.AKOControlConfig().PodEventf(corev1.EventTypeWarning, "CloudMatchingTZNotFound", err.Error())
		utils.AviLog.Fatalf("Failed to derive cloud, err: %s", err)
	}
	utils.AviLog.Infof("Obtained matching cloud to be used: %s", cloudName)
	utils.SetCloudName(cloudName)

	clusterName := lib.GetClusterID()
	err, configuredSEGroup := fetchSEGroup(a.AviRestClient)
	seGroupExists := false
	if err == nil && configuredSEGroup != nil {
		seGroupExists = true
		if len(configuredSEGroup.Markers) == 1 &&
			*configuredSEGroup.Markers[0].Key == lib.ClusterNameLabelKey &&
			len(configuredSEGroup.Markers[0].Values) == 1 &&
			configuredSEGroup.Markers[0].Values[0] == clusterName &&
			isPlacementScopeConfigured(configuredSEGroup) {
			utils.AviLog.Infof("SE Group: %s already configured with the markers: %s", *configuredSEGroup.Name, utils.Stringify(configuredSEGroup.Markers))
			cloudName := strings.Split(*configuredSEGroup.CloudRef, "#")[1]
			utils.AviLog.Infof("Obtained matching cloud to be used: %s", cloudName)
			utils.SetCloudName(cloudName)
			return true
		}
	}

	// This method checks if the cloud in Avi has a SE Group template configured or not. If has the SEG template then it returns true, else false
	if configuredSEGroup == nil {
		uri := "/api/serviceenginegroup/" + segTemplateUuid
		err = lib.AviGet(a.AviRestClient, uri, &configuredSEGroup)
		if err != nil {
			utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
			return false
		}
	}

	if !ConfigureSeGroup(a.AviRestClient, configuredSEGroup, seGroupExists) {
		return false
	}

	return true
}

func fetchSEGroup(client *clients.AviClient, overrideUri ...lib.NextPage) (error, *models.ServiceEngineGroup) {
	var uri string
	if len(overrideUri) == 1 {
		uri = overrideUri[0].NextURI
	} else {
		uri = "/api/serviceenginegroup/?include_name&page_size=100&cloud_ref.name=" + utils.CloudName
	}
	var result session.AviCollectionResult
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
		return err, nil
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return err, nil
	}

	// Using clusterID for advl4.
	clusterName := lib.GetClusterID()
	for _, elem := range elems {
		seg := models.ServiceEngineGroup{}
		err = json.Unmarshal(elem, &seg)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
			continue
		}

		if *seg.Name == clusterName {
			return nil, &seg
		}
	}

	if result.Next != "" {
		// It has a next page, let's recursively call the same method.
		next_uri := strings.Split(result.Next, "/api/serviceenginegroup")
		if len(next_uri) > 1 {
			overrideUri := "/api/serviceenginegroup" + next_uri[1]
			nextPage := lib.NextPage{NextURI: overrideUri}
			return fetchSEGroup(client, nextPage)
		}
	}

	utils.AviLog.Infof("No Service Engine Group found for Cluster.")
	return nil, nil
}

func fetchVcenterServer(client *clients.AviClient) (string, error) {
	uri := "/api/vcenterserver"
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		return "", err
	}
	if result.Count == 0 {
		return "", fmt.Errorf("vcenterServer object not found")
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		return "", err
	}
	vc := avimodels.VCenterServer{}
	err = json.Unmarshal(elems[0], &vc)
	if err != nil {
		return "", err
	}
	return *vc.Name, nil
}

func updateSEGroup() {
	clusterIDs, err := lib.GetAvailabilityZonesCRData(lib.GetDynamicClientSet())
	if err != nil {
		utils.AviLog.Warnf("Failed to get Availability Zones for the supervisor cluster, err: %s", err.Error())
		return
	}
	// Skip Placement scope reconfig if only 1 AZ CR is present
	if len(clusterIDs) < 2 {
		return
	}
	client := avirest.InfraAviClientInstance()
	uri := "/api/serviceenginegroup/?include_name&name=" + lib.GetClusterID()
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		utils.AviLog.Errorf("SE Group Get uri %v returned err %v", uri, err)
		return
	}
	if result.Count != 1 {
		utils.AviLog.Warnf("Expected single SE group for uri: %s", uri)
		return
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return
	}

	seGroup := &models.ServiceEngineGroup{}
	if err = json.Unmarshal(elems[0], &seGroup); err != nil {
		utils.AviLog.Warnf("Failed to unmarshal SE group data, err: %v", err)
		return
	}
	include := true
	vcenterServerName, err := fetchVcenterServer(client)
	if err != nil {
		utils.AviLog.Warnf("Error during API call to fetch Vcenter Server Info, err: %s", err.Error())
		return
	}
	vcRef := fmt.Sprintf("/api/vcenterserver/?name=%s", vcenterServerName)
	if len(seGroup.Vcenters) == 0 {
		seGroup.Vcenters = []*avimodels.PlacementScopeConfig{
			{
				VcenterRef: &vcRef,
				NsxtClusters: &avimodels.NsxtClusters{
					ClusterIds: clusterIDs,
					Include:    &include,
				},
			},
		}
	} else {
		seGroup.Vcenters[0].VcenterRef = &vcRef
		seGroup.Vcenters[0].NsxtClusters = &avimodels.NsxtClusters{
			ClusterIds: clusterIDs,
			Include:    &include,
		}
	}

	response := models.ServiceEngineGroupAPIResponse{}
	uri = "/api/serviceenginegroup/" + *seGroup.UUID
	err = lib.AviPut(client, uri, seGroup, response)
	if err != nil {
		utils.AviLog.Warnf("Failed to update SE group, uri: %s, err: %s", uri, err)
	}
	utils.AviLog.Infof("Successfully updated placement scope in SE Group %s", *seGroup.Name)
}

// ConfigureSeGroup creates the SE group with the supplied properties, alters just the SE group name and the markers.
func ConfigureSeGroup(client *clients.AviClient, seGroup *models.ServiceEngineGroup, segExists bool) bool {
	var err error
	// Change the name of the SE group, and add markers
	*seGroup.Name = lib.GetClusterID()
	markers := []*avimodels.RoleFilterMatchLabel{{
		Key:    proto.String(lib.ClusterNameLabelKey),
		Values: []string{lib.GetClusterID()},
	}}
	seGroup.Markers = markers
	if len(seGroup.Vcenters) == 0 {
		include := true
		vcenterServerName, err := fetchVcenterServer(client)
		if err != nil {
			utils.AviLog.Warnf("Error during API call to fetch Vcenter Server Info, err: %s", err.Error())
			return false
		}
		vcRef := fmt.Sprintf("/api/vcenterserver/?name=%s", vcenterServerName)
		seGroup.Vcenters = append(seGroup.Vcenters,
			&avimodels.PlacementScopeConfig{
				VcenterRef: &vcRef,
				NsxtClusters: &avimodels.NsxtClusters{
					ClusterIds: []string{lib.GetClusterName()},
					Include:    &include,
				},
			})
	}
	response := models.ServiceEngineGroupAPIResponse{}
	var uri string
	if segExists {
		uri = "/api/serviceenginegroup/" + *seGroup.UUID
		err = lib.AviPut(client, uri, seGroup, response)
	} else {
		uri = "/api/serviceenginegroup/"
		err = lib.AviPost(client, uri, seGroup, response)
	}

	if err != nil {
		utils.AviLog.Warnf("Error during API call to CreateOrUpdate the SE group :%v", err.Error())
		return false
	}

	utils.AviLog.Infof("Markers: %v set on Service Engine Group: %v", utils.Stringify(markers), *seGroup.Name)
	return true
}

func (a *AviControllerInfra) AnnotateSystemNamespace(seGroup string, cloudName string, retries ...int) bool {
	retryCount := 0
	if len(retries) > 0 {
		retryCount = retries[0]
	}
	if retryCount > 3 {
		utils.AviLog.Fatalf("Failed to Annotate the %s namespace, shutting down", utils.GetAKONamespace())
	}
	nsName := utils.GetAKONamespace()
	nsObj, err := a.cs.CoreV1().Namespaces().Get(context.TODO(), nsName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to GET the vmware-system-ako namespace details due to the following error :%v", err.Error())
		return a.AnnotateSystemNamespace(seGroup, cloudName, retryCount+1)
	}
	if nsObj.Annotations == nil {
		nsObj.Annotations = make(map[string]string)
	}
	// Update the namespace with the required annotations
	nsObj.Annotations["ako.vmware.com/wcp-se-group"] = seGroup
	nsObj.Annotations["ako.vmware.com/wcp-cloud-name"] = cloudName
	_, err = a.cs.CoreV1().Namespaces().Update(context.TODO(), nsObj, metav1.UpdateOptions{})
	if err != nil {
		utils.AviLog.Warnf("Error occurred while Updating namespace: %v", err)
		return a.AnnotateSystemNamespace(seGroup, cloudName, retryCount+1)
	}
	utils.AviLog.Infof("System Namespace %s annotated with cloud and segroup name", nsName)
	return true
}

func PopulateControllerProperties(cs kubernetes.Interface) error {
	ctrlPropCache := utils.SharedCtrlProp()
	ctrlProps, err := lib.GetControllerPropertiesFromSecret(cs)
	if err != nil {
		return err
	}
	ctrlPropCache.PopulateCtrlProp(ctrlProps)
	return nil
}
