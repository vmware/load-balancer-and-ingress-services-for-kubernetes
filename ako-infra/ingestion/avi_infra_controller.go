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
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/avirest"
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"
)

type AviControllerInfra struct {
	AviRestClient *clients.AviClient
	cs            kubernetes.Interface
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

	if *response.DefaultLicenseTier != AVI_ENTERPRISE && *response.DefaultLicenseTier != AviEnterpriseWithCloudServices {
		errStr := fmt.Sprintf("Avi Controller license is not ENTERPRISE. License tier is: %s", *response.DefaultLicenseTier)
		return errors.New(errStr)
	} else {
		utils.AviLog.Infof("Avi Controller is running with %s license, proceeding with bootup", *response.DefaultLicenseTier)
	}
	return nil
}

func (a *AviControllerInfra) checkNSAnnotations(key string) (string, bool) {
	nsName := utils.GetAKONamespace()
	nsObj, err := a.cs.CoreV1().Namespaces().Get(context.TODO(), nsName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to GET the %s namespace details due to the following error :%v", nsName, err.Error())
		return "", false
	}
	if value, ok := nsObj.Annotations[key]; ok {
		utils.AviLog.Infof("Found key in NS Annotations, key: %s, value: %s", key, value)
		return value, true
	}
	return "", false
}

func (a *AviControllerInfra) checkVirtualService() (string, error) {
	uri := "/api/virtualservice?include_name=True&name.contains=kube-system-kube-apiserver-lb-svc&se_group_ref.name=" + lib.GetClusterID()

	result, err := lib.AviGetCollectionRaw(a.AviRestClient, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err %v", uri, err)
		return "", err
	}
	if result.Count == 0 {
		// Supervisor Control Plane VS not found in Avi
		return "", nil
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return "", err
	}
	for i := 0; i < len(elems); i++ {
		vs := models.VirtualService{}
		err = json.Unmarshal(elems[i], &vs)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal vs data, err: %v", err)
			continue
		}
		if vs.CloudRef != nil && strings.Contains(*vs.CloudRef, "#") {
			cloudName := strings.Split(*vs.CloudRef, "#")[1]
			utils.AviLog.Infof("Found cloud %s associated with vs %s", cloudName, *vs.Name)
			return cloudName, nil
		}
	}
	return "", nil
}

func (a *AviControllerInfra) DeriveCloudNameAndSEGroupTmpl(tz string) (error, string, string) {
	cloudName, found := a.checkNSAnnotations(lib.WCPCloud)
	if !found {
		cloudName, _ = a.checkVirtualService()
	}
	uri := "/api/cloud/"
	if cloudName != "" {
		uri = "/api/cloud/?include_name&name=" + cloudName
	}

	result, err := lib.AviGetCollectionRaw(a.AviRestClient, uri)
	if err != nil {
		utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
		return err, cloudName, ""
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return err, cloudName, ""
	}
	for i := 0; i < len(elems); i++ {
		cloud := models.Cloud{}
		err = json.Unmarshal(elems[i], &cloud)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal cloud data, err: %v", err)
			continue
		}
		if *cloud.Vtype != lib.CLOUD_NSXT || cloud.NsxtConfiguration == nil {
			continue
		}
		if cloud.NsxtConfiguration.ManagementNetworkConfig == nil ||
			cloud.NsxtConfiguration.ManagementNetworkConfig.TransportZone == nil {
			continue
		}
		if cloud.NsxtConfiguration.DataNetworkConfig == nil ||
			cloud.NsxtConfiguration.DataNetworkConfig.TransportZone == nil ||
			*cloud.NsxtConfiguration.DataNetworkConfig.TransportZone != tz {
			continue
		}
		utils.AviLog.Infof("Found NSX-T cloud: %s match Transport Zone: %s", *cloud.Name, tz)
		lib.SetCloudUUID(*cloud.UUID)
		return a.checkSEGroup(cloud)
	}
	return errors.New("cloud not found matching transport zone " + tz), "", ""
}

func (a *AviControllerInfra) checkSEGroup(cloud models.Cloud) (error, string, string) {
	if cloud.SeGroupTemplateRef != nil && *cloud.SeGroupTemplateRef != "" {
		tokenized := strings.Split(*cloud.SeGroupTemplateRef, "/api/serviceenginegroup/")
		if len(tokenized) == 2 {
			return nil, *cloud.Name, tokenized[1]
		}
	}
	// fetch Default-SEGroup uuid
	uri := "/api/serviceenginegroup/?include_name&cloud_ref.name=" + *cloud.Name + "&name=Default-Group"
	results, err := lib.AviGetCollectionRaw(a.AviRestClient, uri)
	if err != nil {
		utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
		return err, *cloud.Name, ""
	}

	elems := make([]json.RawMessage, results.Count)
	err = json.Unmarshal(results.Results, &elems)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return err, *cloud.Name, ""
	}
	if len(elems) == 0 {
		utils.AviLog.Errorf("No ServiceEngine Group with name Default-Group found.")
		return errors.New("No ServiceEngine Group with name Default-Group found."), *cloud.Name, ""
	}

	defaultSEG := models.ServiceEngineGroup{}
	err = json.Unmarshal(elems[0], &defaultSEG)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal cloud data, err: %v", err)
		return err, *cloud.Name, ""
	}
	return nil, *cloud.Name, *defaultSEG.UUID
}

func isPlacementScopeConfigured(configuredSEGroup *models.ServiceEngineGroup) bool {
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
		utils.AviLog.Fatalf("Failed to derive cloud and template SE Group in Avi, transportZone: %s, err: %s", tz, err.Error())
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
			return seGroupExists
		}
	}

	// This method checks if the cloud in Avi has a SE Group template configured or not. If has the SEG template then it returns true, else false
	if configuredSEGroup == nil {
		uri := "/api/serviceenginegroup/" + segTemplateUuid
		err = lib.AviGet(a.AviRestClient, uri, &configuredSEGroup)
		if err != nil {
			utils.AviLog.Fatalf("Failed to fetch template SE Group in Avi, segID: %s, err: %s", segTemplateUuid, err.Error())
		}
	}

	if err = ConfigureSeGroup(a.AviRestClient, configuredSEGroup, seGroupExists); err != nil {
		utils.AviLog.Fatalf("Failed to configure SE Group in Avi, err: %s", err.Error())
	}

	return seGroupExists
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

func fetchVcenterServer(vCenters map[string]string, client *clients.AviClient, nextPage ...lib.NextPage) error {
	uri := "/api/vcenterserver?include_name&cloud_ref.name=" + utils.CloudName
	if len(nextPage) > 0 {
		uri = nextPage[0].NextURI
	}
	result, err := lib.AviGetCollectionRaw(client, uri)
	if err != nil {
		return err
	}
	if result.Count == 0 {
		return fmt.Errorf("vcenterServer object not found")
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		return err
	}
	for _, elem := range elems {
		vc := models.VCenterServer{}
		err = json.Unmarshal(elem, &vc)
		if err != nil {
			return err
		}
		vCenters[*vc.UUID] = *vc.Name
	}

	if result.Next != "" {
		next_uri := strings.Split(result.Next, "/api/vcenterserver")
		if len(next_uri) > 1 {
			overrideUri := "/api/vcenterserver" + next_uri[1]
			nextPage := lib.NextPage{NextURI: overrideUri}
			return fetchVcenterServer(vCenters, client, nextPage)
		}
	}

	return nil
}

func fetchClustersInVC(vcenterServerUUID string, client *clients.AviClient, nextPage ...lib.NextPage) ([]string, error) {
	clusters := []string{}
	uri := "/api/nsxt/clusters"
	if len(nextPage) > 0 {
		uri = nextPage[0].NextURI
	}
	var response interface{}
	payload := map[string]string{
		"cloud_uuid":   lib.GetCloudUUID(),
		"vcenter_uuid": vcenterServerUUID,
	}
	err := lib.AviPost(client, uri, payload, &response)
	if err != nil {
		utils.AviLog.Errorf("Faled to get NSXT Clusters, vcServer: %s, err: %s", vcenterServerUUID, err.Error())
		return []string{}, err
	}
	res, _ := response.(map[string]interface{})
	resNSXTClusters, _ := res["resource"].(map[string]interface{})
	resClusters, _ := resNSXTClusters["nsxt_clusters"].([]interface{})
	for _, cluster := range resClusters {
		cl, _ := cluster.(map[string]interface{})
		clusters = append(clusters, cl["vc_mobj_id"].(string))
	}
	return clusters, nil
}

func inSlice(supSlice []string, subSlice []string) bool {
	m := make(map[string]struct{})
	for _, s := range supSlice {
		m[s] = struct{}{}
	}
	for _, s := range subSlice {
		if _, ok := m[s]; !ok {
			return false
		}
	}
	return true
}

func getVCServerName(wcpClusters []string, vCenters map[string]string, client *clients.AviClient) (string, error) {
	if len(vCenters) == 1 {
		for _, vcName := range vCenters {
			return vcName, nil
		}
	}
	for vcUUID, vcName := range vCenters {
		clusters, err := fetchClustersInVC(vcUUID, client)
		if err != nil {
			continue
		}
		if inSlice(wcpClusters, clusters) {
			return vcName, nil
		}
	}
	return "", fmt.Errorf("vCenterServer not found corresponding to WCP cluster")
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
	vCenters := make(map[string]string)
	err = fetchVcenterServer(vCenters, client)
	if err != nil {
		utils.AviLog.Warnf("Error during API call to fetch Vcenter Server Info, err: %s", err.Error())
		return
	}
	vcenterServerName, err := getVCServerName(clusterIDs, vCenters, client)
	if err != nil {
		return
	}
	vcRef := fmt.Sprintf("/api/vcenterserver/?name=%s", vcenterServerName)
	if len(seGroup.Vcenters) == 0 {
		seGroup.Vcenters = []*models.PlacementScopeConfig{
			{
				VcenterRef: &vcRef,
				NsxtClusters: &models.NsxtClusters{
					ClusterIds: clusterIDs,
					Include:    &include,
				},
			},
		}
	} else {
		seGroup.Vcenters[0].VcenterRef = &vcRef
		seGroup.Vcenters[0].NsxtClusters = &models.NsxtClusters{
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
func ConfigureSeGroup(client *clients.AviClient, seGroup *models.ServiceEngineGroup, segExists bool) error {
	var err error
	// Change the name of the SE group, and add markers
	*seGroup.Name = lib.GetClusterID()
	markers := []*models.RoleFilterMatchLabel{{
		Key:    proto.String(lib.ClusterNameLabelKey),
		Values: []string{lib.GetClusterID()},
	}}
	seGroup.Markers = markers
	if len(seGroup.Vcenters) == 0 {
		include := true
		vCenters := make(map[string]string)
		err := fetchVcenterServer(vCenters, client)
		if err != nil {
			utils.AviLog.Warnf("Error during API call to fetch Vcenter Server Info, err: %s", err.Error())
			return err
		}
		clusterIDs, err := lib.GetAvailabilityZonesCRData(lib.GetDynamicClientSet())
		if err != nil {
			utils.AviLog.Warnf("Failed to get Availability Zones for the supervisor cluster, err: %s", err.Error())
			clusterIDs = []string{lib.GetClusterName()}
		}
		vcenterServerName, err := getVCServerName(clusterIDs, vCenters, client)
		if err != nil {
			return err
		}
		vcRef := fmt.Sprintf("/api/vcenterserver/?name=%s", vcenterServerName)
		seGroup.Vcenters = append(seGroup.Vcenters,
			&models.PlacementScopeConfig{
				VcenterRef: &vcRef,
				NsxtClusters: &models.NsxtClusters{
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
		return err
	}

	utils.AviLog.Infof("Markers: %v set on Service Engine Group: %v", utils.Stringify(markers), *seGroup.Name)
	return nil
}

func (a *AviControllerInfra) AnnotateSystemNamespace(seGroup, cloudName, akoUser string, retries ...int) bool {
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
		return a.AnnotateSystemNamespace(seGroup, cloudName, akoUser, retryCount+1)
	}
	if nsObj.Annotations == nil {
		nsObj.Annotations = make(map[string]string)
	}
	// Update the namespace with the required annotations
	nsObj.Annotations[lib.WCPSEGroup] = seGroup
	nsObj.Annotations[lib.WCPCloud] = cloudName
	nsObj.Annotations[lib.WCPAKOUserClusterName] = akoUser
	_, err = a.cs.CoreV1().Namespaces().Update(context.TODO(), nsObj, metav1.UpdateOptions{})
	if err != nil {
		utils.AviLog.Warnf("Error occurred while Updating namespace: %v", err)
		return a.AnnotateSystemNamespace(seGroup, cloudName, akoUser, retryCount+1)
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

func (a *AviControllerInfra) GetClusterNameToBeUsedInAKOUser(segExists bool) (string, error) {
	clusterID := lib.GetClusterID()
	clusterIDArr := strings.Split(clusterID, ":")
	if !segExists {
		// Include first 5 characters to add more uniqueness to cluster name
		return clusterIDArr[0] + "-" + clusterIDArr[1][:5], nil
	}
	clusterName, found := a.checkNSAnnotations(lib.WCPAKOUserClusterName)
	if found {
		return clusterName, nil
	}
	uri := "/api/virtualservice?name.contains=kube-system-kube-apiserver-lb-svc&se_group_ref.name=" + clusterID
	result, err := lib.AviGetCollectionRaw(a.AviRestClient, uri)
	if err != nil {
		utils.AviLog.Warnf("Get uri %v returned err %v", uri, err)
		return "", err
	}
	if result.Count == 0 {
		// Supervisor Control Plane VS not found in Avi
		return clusterIDArr[0] + "-" + clusterIDArr[1][:5], nil
	}

	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
		return "", err
	}

	for i := 0; i < len(elems); i++ {
		vs := models.VirtualService{}
		err = json.Unmarshal(elems[i], &vs)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal vs data, err: %v", err)
			continue
		}
		if vs.CreatedBy != nil {
			_, clusterName, found := strings.Cut(*vs.CreatedBy, "ako-")
			if found {
				return clusterName, nil
			}
			err = fmt.Errorf("createdBy field does not follow the expected pattern (ako-<cluster_name>-<uuid_substring>), vs: %s, created_by: %s", *vs.Name, *vs.CreatedBy)
		} else {
			err = fmt.Errorf("createdBy field not set for VS, need to set it for AKO to boot up properly: %s", *vs.Name)
		}
	}
	return "", err
}
