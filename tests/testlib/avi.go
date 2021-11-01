/*
 * Copyright 2019-2020 VMware, Inc.
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

package testlib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

const defaultMockFilePath = "../avimockobjects"

var AviFakeClientInstance *httptest.Server
var FakeServerMiddleware InjectFault
var FakeAviObjects = []string{
	"cloud",
	"ipamdnsproviderprofile",
	"ipamdnsproviderprofiledomainlist",
	"network",
	"pool",
	"poolgroup",
	"virtualservice",
	"vrfcontext",
	"vsdatascriptset",
	"serviceenginegroup",
	"tenant",
	"vsvip",
}

type InjectFault func(w http.ResponseWriter, r *http.Request)

func AddMiddleware(exec InjectFault) {
	FakeServerMiddleware = exec
}

func ResetMiddleware() {
	FakeServerMiddleware = nil
}

func NewAviFakeClientInstance(kubeclient *k8sfake.Clientset, skipCachePopulation ...bool) {
	if AviFakeClientInstance == nil {
		AviFakeClientInstance = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			utils.AviLog.Infof("[fakeAPI]: %s %s\n", r.Method, r.URL)

			if FakeServerMiddleware != nil {
				FakeServerMiddleware(w, r)
				return
			}

			NormalControllerServer(w, r)
		}))

		url := strings.Split(AviFakeClientInstance.URL, "https://")[1]
		os.Setenv("CTRL_IPADDRESS", url)
		os.Setenv("FULL_SYNC_INTERVAL", "600")
		// resets avi client pool instance, allows to connect with the new `ts` server
		cache.AviClientInstance = nil
		k8s.PopulateControllerProperties(kubeclient)
		if len(skipCachePopulation) == 0 || skipCachePopulation[0] == false {
			k8s.PopulateCache()
		}
	}
}

func NormalControllerServer(w http.ResponseWriter, r *http.Request, args ...string) {
	mockFilePath := defaultMockFilePath
	if len(args) > 0 {
		mockFilePath = args[0]
	}
	url := r.URL.EscapedPath()
	var resp map[string]interface{}
	var finalResponse []byte
	var shardVSNum string
	var object string
	addrPrefix := "10.250.250"
	publicAddrPrefix := "35.250.250"
	urlSlice := strings.Split(strings.Trim(url, "/"), "/")
	if len(urlSlice) > 1 {
		object = urlSlice[1]
	}

	if r.Method == "POST" && !strings.Contains(url, "login") {
		data, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(data, &resp)
		rName := resp["name"].(string)
		objURL := fmt.Sprintf("https://localhost/api/%s/%s-%s-%s#%s", object, object, rName, RANDOMUUID, rName)

		// adding additional 'uuid' and 'url' (read-only) fields in the response
		resp["url"] = objURL
		resp["uuid"] = fmt.Sprintf("%s-%s-%s", object, rName, RANDOMUUID)

		if strings.Contains(url, "virtualservice") {
			objURL := fmt.Sprintf("https://localhost/api/%s/%s-%s-%s#%s", object, object, rName, RANDOMUUID, rName)
			// adding additional 'uuid' and 'url' (read-only) fields in the response
			resp["url"] = objURL
			resp["uuid"] = fmt.Sprintf("%s-%s-%s", object, rName, RANDOMUUID)

			// add vip for status update checks
			// use vh_parent_vs_uuid for sniVS, and name for normal VSes

			if strings.Contains(rName, "public") {
				resp["vip"] = []interface{}{map[string]interface{}{"floating_ip": map[string]string{"addr": "35.250.250.1", "type": "V4"}}}
			} else if strings.Contains(rName, "multivip") {
				if strings.Contains(rName, "public") {
					resp["vip"] = []interface{}{
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".1", "type": "V4"},
							"floating_ip": map[string]string{"addr": publicAddrPrefix + ".1", "type": "V4"}},
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".2", "type": "V4"},
							"floating_ip": map[string]string{"addr": publicAddrPrefix + ".2", "type": "V4"}},
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".3", "type": "V4"},
							"floating_ip": map[string]string{"addr": publicAddrPrefix + ".3", "type": "V4"}},
					}
				} else {
					resp["vip"] = []interface{}{
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".1", "type": "V4"}},
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".2", "type": "V4"}},
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".3", "type": "V4"}},
					}
				}
			} else if vsType := resp["type"]; vsType == "VS_TYPE_VH_CHILD" {
				parentVSName := strings.Split(resp["vh_parent_vs_uuid"].(string), "name=")[1]
				shardVSNum = strings.Split(parentVSName, "cluster--Shared-L7-")[1]

				resp["vh_parent_vs_ref"] = fmt.Sprintf("https://localhost/api/virtualservice/virtualservice-%s-%s#%s", parentVSName, RANDOMUUID, parentVSName)
				resp["vip"] = []interface{}{map[string]interface{}{"ip_address": map[string]string{"addr": fmt.Sprintf("%s.1%s", addrPrefix, shardVSNum), "type": "V4"}}}
			} else if strings.Contains(rName, "Shared-L7-EVH-") {
				shardVSNum = strings.Split(rName, "Shared-L7-EVH-")[1]
				if strings.Contains(shardVSNum, "NS-") {
					shardVSNum = "0"
				}
				resp["vip"] = []interface{}{map[string]interface{}{"ip_address": map[string]string{"addr": fmt.Sprintf("%s.1%s", addrPrefix, shardVSNum), "type": "V4"}}}
			} else if strings.Contains(rName, "Shared-L7") {
				shardVSNum = strings.Split(rName, "Shared-L7-")[1]
				resp["vip"] = []interface{}{map[string]interface{}{"ip_address": map[string]string{"addr": fmt.Sprintf("%s.1%s", addrPrefix, shardVSNum), "type": "V4"}}}
			} else {
				resp["vip"] = []interface{}{map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".1", "type": "V4"}}}
			}
			resp["vsvip_ref"] = fmt.Sprintf("https://localhost/api/vsvip/vsvip-%s-%s#%s", rName, RANDOMUUID, rName)
		} else if strings.Contains(url, "vsvip") {
			objURL := fmt.Sprintf("https://localhost/api/%s/%s-%s-%s#%s", object, object, rName, RANDOMUUID, rName)
			// adding additional 'uuid' and 'url' (read-only) fields in the response
			resp["url"] = objURL
			resp["uuid"] = fmt.Sprintf("%s-%s-%s", object, rName, RANDOMUUID)

			if strings.Contains(rName, "public") {
				fipAddress := "35.250.250.1"
				resp["vip"].([]interface{})[0].(map[string]interface{})["floating_ip"] = map[string]string{"addr": fipAddress, "type": "V4"}
			} else if strings.Contains(rName, "multivip") {
				if strings.Contains(rName, "public") {
					resp["vip"] = []interface{}{
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".1", "type": "V4"},
							"floating_ip": map[string]string{"addr": publicAddrPrefix + ".1", "type": "V4"}},
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".2", "type": "V4"},
							"floating_ip": map[string]string{"addr": publicAddrPrefix + ".2", "type": "V4"}},
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".3", "type": "V4"},
							"floating_ip": map[string]string{"addr": publicAddrPrefix + ".3", "type": "V4"}},
					}
				} else {
					resp["vip"] = []interface{}{
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".1", "type": "V4"}},
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".2", "type": "V4"}},
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".3", "type": "V4"}},
					}
				}
			} else if vsType := resp["type"]; vsType == "VS_TYPE_VH_CHILD" {
				parentVSName := strings.Split(resp["vh_parent_vs_uuid"].(string), "name=")[1]
				shardVSNum = strings.Split(parentVSName, "cluster--Shared-L7-")[1]
				resp["vip"] = []interface{}{map[string]interface{}{"ip_address": map[string]string{"addr": fmt.Sprintf("%s.1%s", addrPrefix, shardVSNum), "type": "V4"}}}
			} else if strings.Contains(rName, "Shared-L7-EVH-") {
				shardVSNum = strings.Split(rName, "Shared-L7-EVH-")[1]
				if strings.Contains(shardVSNum, "NS-") {
					shardVSNum = "0"
				}
				resp["vip"] = []interface{}{map[string]interface{}{"ip_address": map[string]string{"addr": fmt.Sprintf("%s.1%s", addrPrefix, shardVSNum), "type": "V4"}}}
			} else if strings.Contains(rName, "Shared-L7") {
				shardVSNum = strings.Split(rName, "Shared-L7-")[1]
				resp["vip"] = []interface{}{map[string]interface{}{"ip_address": map[string]string{"addr": fmt.Sprintf("%s.1%s", addrPrefix, shardVSNum), "type": "V4"}}}
			} else {
				resp["vip"] = []interface{}{map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".1", "type": "V4"}}}
			}
		}
		finalResponse, _ = json.Marshal(resp)
		w.WriteHeader(http.StatusOK)
		w.Write(finalResponse)

	} else if r.Method == "PUT" {
		data, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(data, &resp)
		resp["uuid"] = strings.Split(strings.Trim(url, "/"), "/")[2]
		if vsType, ok := resp["type"]; ok && vsType == "VS_TYPE_VH_CHILD" {
			parentVSName := strings.Split(resp["vh_parent_vs_uuid"].(string), "name=")[1]
			resp["vh_parent_vs_ref"] = fmt.Sprintf("https://localhost/api/virtualservice/virtualservice-%s-%s#%s", parentVSName, RANDOMUUID, parentVSName)
		}
		if val, ok := resp["name"]; !ok || val == nil {
			resp["name"] = strings.ReplaceAll(object, "-random-uuid", "")
		}
		if strings.Contains(url, "vsvip") {
			if resp["vip"] == nil || resp["vip"].([]interface{})[0].(map[string]interface{})["ip_address"] == nil {
				resp["vip"] = []interface{}{map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".1", "type": "V4"}}}
			}
			if strings.Contains(url, "public") {
				resp["vip"].([]interface{})[0].(map[string]interface{})["floating_ip"] = map[string]string{"addr": publicAddrPrefix + ".1", "type": "V4"}
			}
			if strings.Contains(url, "multivip") {
				if strings.Contains(url, "public") {
					resp["vip"] = []interface{}{
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".1", "type": "V4"},
							"floating_ip": map[string]string{"addr": publicAddrPrefix + ".1", "type": "V4"}},
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".2", "type": "V4"},
							"floating_ip": map[string]string{"addr": publicAddrPrefix + ".2", "type": "V4"}},
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".3", "type": "V4"},
							"floating_ip": map[string]string{"addr": publicAddrPrefix + ".3", "type": "V4"}},
					}
				} else {
					resp["vip"] = []interface{}{
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".1", "type": "V4"}},
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".2", "type": "V4"}},
						map[string]interface{}{"ip_address": map[string]string{"addr": addrPrefix + ".3", "type": "V4"}},
					}
				}
			}
		}
		finalResponse, _ = json.Marshal(resp)
		w.WriteHeader(http.StatusOK)
		w.Write(finalResponse)

	} else if r.Method == "DELETE" {
		w.WriteHeader(http.StatusNoContent)
		w.Write(finalResponse)

	} else if r.Method == "PATCH" && strings.Contains(url, "vrfcontext") {
		// This won't help in checking for Cache values, since we are sending back static content
		// It is only to remove API call warning related to vrfcontext PATCH calls.
		w.WriteHeader(http.StatusOK)
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/vrfcontext_uuid_mock.json", mockFilePath))
		w.Write(data)

	} else if r.Method == "GET" && strings.Contains(r.URL.RawQuery, "aviref") {
		// block to handle
		if strings.Contains(r.URL.RawQuery, "thisisaviref") {
			w.WriteHeader(http.StatusOK)
			data, _ := ioutil.ReadFile(fmt.Sprintf("%s/crd_mock.json", mockFilePath))
			w.Write(data)
		} else if strings.Contains(r.URL.RawQuery, "thisisBADaviref") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"results": [], "count": 0}`))
		}

	} else if r.Method == "GET" && strings.Contains(url, "/api/cloud/") {
		var data []byte
		if strings.HasSuffix(r.URL.RawQuery, "CLOUD_NONE") {
			data, _ = ioutil.ReadFile(fmt.Sprintf("%s/%s_mock.json", mockFilePath, "CLOUD_NONE"))
		} else if strings.HasSuffix(r.URL.RawQuery, "CLOUD_AZURE") {
			data, _ = ioutil.ReadFile(fmt.Sprintf("%s/%s_mock.json", mockFilePath, "CLOUD_AZURE"))
		} else if strings.HasSuffix(r.URL.RawQuery, "CLOUD_AWS") {
			data, _ = ioutil.ReadFile(fmt.Sprintf("%s/%s_mock.json", mockFilePath, "CLOUD_AWS"))
		} else {
			data, _ = ioutil.ReadFile(fmt.Sprintf("%s/%s_mock.json", mockFilePath, "CLOUD_VCENTER"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)

	} else if r.Method == "GET" && inArray(FakeAviObjects, object) {
		FeedMockCollectionData(w, r, mockFilePath)

	} else if strings.Contains(url, "login") {
		// This is used for /login --> first request to controller
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": "true"}`))
	} else if strings.Contains(url, "initial-data") {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"version": {"Version": "20.1.2"}}`))
	} else if strings.Contains(url, "/api/cluster/runtime") {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"node_states": [{"name": "10.79.169.60","role": "CLUSTER_LEADER","up_since": "2020-10-28 04:58:48"}]}`))
	}
}

func inArray(a []string, b string) bool {
	for _, k := range a {
		if k == b {
			return true
		}
	}
	return false
}

// FeedMockCollectionData reads data from avimockobjects/*.json files and returns mock data
// for GET objects list API. GET /api/virtualservice returns from virtualservice_mock.json and so on
func FeedMockCollectionData(w http.ResponseWriter, r *http.Request, mockFilePath string) {
	url := r.URL.EscapedPath() // url = //api/<object>/:objectId
	splitURL := strings.Split(strings.Trim(url, "/"), "/")

	if r.Method == "GET" {
		var data []byte
		var err error
		if len(splitURL) == 2 {
			data, err = ioutil.ReadFile(fmt.Sprintf("%s/%s_mock.json", mockFilePath, splitURL[1]))
			if err != nil {
				fmt.Printf("Error in reading from file: %v", err)
			}
		} else if len(splitURL) == 3 {
			// with uuid
			data, err = ioutil.ReadFile(fmt.Sprintf("%s/%s_uuid_mock.json", mockFilePath, splitURL[1]))
			if err != nil {
				fmt.Printf("Error in reading from file: %v", err)
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	} else if strings.Contains(url, "login") {
		// This is used for /login --> first request to controller
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": "true"}`))
	}
}
