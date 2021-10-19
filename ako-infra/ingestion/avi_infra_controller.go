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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/vmware/alb-sdk/go/models"
	"k8s.io/client-go/kubernetes"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/third_party/github.com/vmware/alb-sdk/go/session"
)

type AviControllerInfra struct {
	AviRestClients *utils.AviRestClientPool
	cs             *kubernetes.Clientset
}

func NewAviControllerInfra(cs *kubernetes.Clientset) *AviControllerInfra {
	PopulateControllerProperties(cs)
	AviRestClientsPool := avicache.SharedAVIClients()
	return &AviControllerInfra{AviRestClients: AviRestClientsPool, cs: cs}
}

func (a *AviControllerInfra) InitInfraController() {
	if a.AviRestClients == nil {
		utils.AviLog.Fatalf("Avi client not initialized during Infra bootup")
	}

	if a.AviRestClients != nil && !avicache.IsAviClusterActive(a.AviRestClients.AviClient[0]) {
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
	err := lib.AviGet(a.AviRestClients.AviClient[0], uri, &response)
	if err != nil {
		utils.AviLog.Warnf("System config Get uri %v returned err %v", uri, err)
		return err
	}

	if *response.DefaultLicenseTier != AVI_ENTERPRISE {
		errStr := fmt.Sprintf("Avi Controller license is not ENTERPRISE. License tier is: %s", *response.DefaultLicenseTier)
		return errors.New(errStr)
	} else {
		utils.AviLog.Infof("Avi Controller is running with ENTERPRISE license, proceeding with bootup")
	}
	return nil
}

func (a *AviControllerInfra) DeriveCloudNameAndSEGroupTmpl(tz string) (error, string, string) {
	// This method queries the Avi controller for all available cloud and then returns the cloud that matches the supplied transport zone
	uri := "/api/cloud/"
	result, err := lib.AviGetCollectionRaw(a.AviRestClients.AviClient[0], uri)
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
	for i := 0; i < len(elems); i++ {
		cloud := models.Cloud{}
		err = json.Unmarshal(elems[i], &cloud)
		if err != nil {
			utils.AviLog.Warnf("Failed to unmarshal cloud data, err: %v", err)
			continue
		}
		vtype := *cloud.Vtype
		if vtype == lib.CLOUD_NSXT && cloud.NsxtConfiguration != nil {
			if cloud.NsxtConfiguration.TransportZone != nil && *cloud.NsxtConfiguration.TransportZone == tz {
				utils.AviLog.Infof("Found NSX-T cloud :%s match Transport Zone : %s", *cloud.Name, tz)
				if *cloud.SeGroupTemplateRef != "" {
					tokenized := strings.Split(*cloud.SeGroupTemplateRef, "/api/serviceenginegroup/")
					if len(tokenized) == 2 {
						return nil, *cloud.Name, tokenized[1]
					}
				}
			}
			return nil, *cloud.Name, ""
		}
	}
	return errors.New("Cloud not found"), "", ""
}

func (a *AviControllerInfra) SetupSEGroup(tz string) bool {
	// This method checks if the cloud in Avi has a SE Group template configured or not. If has the SEG template then it returns true, else false
	err, cloudName, segUuid := a.DeriveCloudNameAndSEGroupTmpl(tz)
	if err != nil {
		return false
	}
	var uri string
	if segUuid == "" {
		// The cloud does not have a SEG template set, use `Default-Group`
		uri = "/api/serviceenginegroup/?include_name&cloud_ref.name=" + cloudName + "&name=Default-Group"
	} else {
		// se group template exists, use the same to fetch the SE group details and use it to create the new SE group
		// The cloud does not have a SEG template set, use `Default-Group`
		uri = "/api/serviceenginegroup/" + segUuid
	}
	result, err := lib.AviGetCollectionRaw(a.AviRestClients.AviClient[0], uri)
	if err != nil {
		utils.AviLog.Errorf("Get uri %v returned err %v", uri, err)
		return false
	}
	// Construct an SE group based on parameters in the `Default-Group`
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Errorf("Failed to unmarshal data, err: %v", err)
		return false
	}

	for _, elem := range elems {
		seg := models.ServiceEngineGroup{}
		if err := json.Unmarshal(elem, &seg); err != nil {
			utils.AviLog.Warnf("Failed to unmarshal data, err: %v", err)
			continue
		}

		if !ConfigureSeGroup(a.AviRestClients.AviClient[0], &seg) {
			return false
		}
	}

	return true
}

// ConfigureSeGroup creates the SE group with the supplied properties, alters just the SE group name and the labels.
func ConfigureSeGroup(client *clients.AviClient, seGroup *models.ServiceEngineGroup) bool {
	// Change the name of the SE group
	*seGroup.Name = lib.GetClusterID()

	uri := "/api/serviceenginegroup/"
	// Add the labels.
	seGroup.Labels = lib.GetLabels()
	response := models.ServiceEngineGroupAPIResponse{}
	// If tenants per cluster is enabled then the X-Avi-Tenant needs to be set to admin for vrfcontext and segroup updates
	if lib.GetTenantsPerCluster() && lib.IsCloudInAdminTenant {
		SetAdminTenant := session.SetTenant(lib.GetAdminTenant())
		SetTenant := session.SetTenant(lib.GetTenant())
		SetAdminTenant(client.AviSession)
		defer SetTenant(client.AviSession)
	}

	err := lib.AviPost(client, uri, seGroup, response)
	if err != nil {
		utils.AviLog.Warnf("Error during POST call to create the SE group :%v", err.Error())
		return false
	}
	utils.AviLog.Infof("labels: %v set on Service Engine Group :%v", utils.Stringify(lib.GetLabels()), *seGroup.Name)
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
