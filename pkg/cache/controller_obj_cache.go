/*
 * [2013] - [2018] Avi Networks Incorporated
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

package cache

import (
	"os"
	"regexp"
	"sync"

	"github.com/avinetworks/sdk/go/clients"
	"github.com/avinetworks/sdk/go/session"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
)

type AviObjCache struct {
	VsCache       *AviCache
	PgCache       *AviCache
	DSCache       *AviCache
	PoolCache     *AviCache
	CloudKeyCache *AviCache
}

func NewAviObjCache() *AviObjCache {
	c := AviObjCache{}
	c.VsCache = NewAviCache()
	c.PgCache = NewAviCache()
	c.DSCache = NewAviCache()
	c.PoolCache = NewAviCache()
	c.CloudKeyCache = NewAviCache()
	return &c
}

var cacheInstance *AviObjCache
var cacheOnce sync.Once

func SharedAviObjCache() *AviObjCache {
	cacheOnce.Do(func() {
		cacheInstance = NewAviObjCache()
	})
	return cacheInstance
}

func (c *AviObjCache) AviObjCachePopulate(client *clients.AviClient,
	version string, cloud string) {
	SetTenant := session.SetTenant("*")
	SetTenant(client.AviSession)
	SetVersion := session.SetVersion(version)
	SetVersion(client.AviSession)

	// Populate the VS cache
	c.AviObjVSCachePopulate(client, cloud)
	c.AviCloudPropertiesPopulate(client, cloud)

}

// TODO (sudswas): Should this be run inside a go routine for parallel population
// to reduce bootup time when the system is loaded. Variable duplication expected.
func (c *AviObjCache) AviObjVSCachePopulate(client *clients.AviClient,
	cloud string) {
	var rest_response interface{}
	// TODO Retrieve just fields we care about
	uri := "/api/virtualservice?include_name=true&cloud_ref.name=" + cloud
	err := client.AviSession.Get(uri, &rest_response)

	if err != nil {
		utils.AviLog.Warning.Printf("Vs Get uri %v returned err %v", uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("Vs Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("Vs Get uri %v returned %v vses", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T", resp["results"])
			return
		}
		for _, vs_intf := range results {
			vs, ok := vs_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("vs_intf not of type map[string] interface{}. Instead of type %T", vs_intf)
				continue
			}
			var sni_child_collection []string
			vh_child, found := vs["vh_child_vs_uuid"]
			if found {
				for _, child := range vh_child.([]interface{}) {
					sni_child_collection = append(sni_child_collection, child.(string))
				}

			}
			if vs["cloud_config_cksum"] != nil {
				k := NamespaceName{Namespace: utils.ADMIN_NS, Name: vs["name"].(string)}
				var vip string
				if len(vs["vip"].([]interface{})) > 0 {
					vip = (vs["vip"].([]interface{})[0].(map[string]interface{})["ip_address"]).(map[string]interface{})["addr"].(string)
				}
				vs_cache, found := c.VsCache.AviCacheGet(k)
				if found {
					vs_cache_obj, ok := vs_cache.(*AviVsCache)
					if ok {
						if vs_cache_obj.Uuid == vs["uuid"].(string) {
							// Same object - let's just refresh the values.
							vs_cache_obj.CloudConfigCksum = vs["cloud_config_cksum"].(string)
							vs_cache_obj.SNIChildCollection = sni_child_collection
							utils.AviLog.Info.Printf("Updated Vs cache k %v val %v",
								k, vs_cache_obj)
						} else {
							// New object
							vs_cache_obj := AviVsCache{Name: vs["name"].(string),
								Tenant: utils.ADMIN_NS, Uuid: vs["uuid"].(string), Vip: vip,
								CloudConfigCksum:   vs["cloud_config_cksum"].(string),
								SNIChildCollection: sni_child_collection}

							c.VsCache.AviCacheAdd(k, &vs_cache_obj)
							utils.AviLog.Info.Printf("Added Vs cache k %v val %v",
								k, vs_cache_obj)
						}
					} else {
						// New object
						vs_cache_obj := AviVsCache{Name: vs["name"].(string),
							Tenant: utils.ADMIN_NS, Uuid: vs["uuid"].(string), Vip: vip,
							CloudConfigCksum:   vs["cloud_config_cksum"].(string),
							SNIChildCollection: sni_child_collection}

						c.VsCache.AviCacheAdd(k, &vs_cache_obj)
						utils.AviLog.Info.Printf("Added Vs cache k %v val %v",
							k, vs_cache_obj)
					}
				} else {
					vs_cache_obj := AviVsCache{Name: vs["name"].(string),
						Tenant: utils.ADMIN_NS, Uuid: vs["uuid"].(string), Vip: vip,
						CloudConfigCksum:   vs["cloud_config_cksum"].(string),
						SNIChildCollection: sni_child_collection}

					c.VsCache.AviCacheAdd(k, &vs_cache_obj)
					utils.AviLog.Info.Printf("Added Vs cache k %v val %v",
						k, vs_cache_obj)
				}

				c.AviPGCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
				c.AviPoolCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
				c.AviDataScriptPopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
			}
		}
	}
}

//Design library methods to remove repeatation of code.
func (c *AviObjCache) AviPGCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName) {
	var rest_response interface{}

	var pg_key_collection []NamespaceName
	uri := "/api/poolgroup?include_name=true&cloud_ref.name=" + cloud + "&referred_by=virtualservice:" + vs_uuid
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("PG Get uri %v returned err %v", uri, err)
		return
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("PG Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("PG Get uri %v returned %v PGs", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T for PGs", resp["results"])
			return
		}
		for _, pg_intf := range results {
			pg, ok := pg_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("pg_intf not of type map[string] interface{}. Instead of type %T", pg_intf)
				continue
			}
			if pg["cloud_config_cksum"] != nil {
				pg_cache_obj := AviPGCache{Name: pg["name"].(string),
					Tenant: tenant, Uuid: pg["uuid"].(string),
					CloudConfigCksum: pg["cloud_config_cksum"].(string)}
				k := NamespaceName{Namespace: tenant, Name: pg["name"].(string)}
				c.PgCache.AviCacheAdd(k, &pg_cache_obj)
				utils.AviLog.Info.Printf("Added PG cache key %v val %v",
					k, pg_cache_obj)
				pg_key_collection = append(pg_key_collection, k)
			}

		}
		vs_cache, found := c.VsCache.AviCacheGet(vsKey)
		if found {
			vs_cache_obj, ok := vs_cache.(*AviVsCache)
			if !ok {
				utils.AviLog.Warning.Printf("Unable to cast to VS object: %v", vsKey)
				return
			}
			vs_cache_obj.PGKeyCollection = pg_key_collection
		} else {
			utils.AviLog.Warning.Printf("VS cache not found for key: %v . Unable to update PG collection", vsKey)
		}
	}
}

func (c *AviObjCache) AviPoolCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName) {
	var rest_response interface{}
	var err error
	var pool_key_collection []NamespaceName
	// TODO Retrieve just fields we care about
	uri := "/api/pool?include_name=true&cloud_ref.name=" + cloud + "&referred_by=virtualservice:" + vs_uuid
	err = client.AviSession.Get(uri, &rest_response)

	if err != nil {
		utils.AviLog.Warning.Printf("Pool Get uri %v returned err %v", uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("Pool Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("Pool Get uri %v returned %v pools", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T", resp["results"])
			return
		}
		for _, pool_intf := range results {
			pool, ok := pool_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("pool_intf not of type map[string] interface{}. Instead of type %T", pool_intf)
				continue
			}

			pool_cache_obj := AviPoolCache{Name: pool["name"].(string),
				Tenant: tenant, Uuid: pool["uuid"].(string),
				CloudConfigCksum: pool["cloud_config_cksum"].(string)}

			k := NamespaceName{Namespace: tenant, Name: pool["name"].(string)}

			c.PoolCache.AviCacheAdd(k, &pool_cache_obj)
			pool_key_collection = append(pool_key_collection, k)
			utils.AviLog.Info.Printf("Added Pool cache key %v val %v",
				k, pool_cache_obj)

		}
		vs_cache, found := c.VsCache.AviCacheGet(vsKey)
		if found {
			vs_cache_obj, ok := vs_cache.(*AviVsCache)
			if !ok {
				utils.AviLog.Warning.Printf("Unable to cast to VS object: %v", vsKey)
				return
			}
			vs_cache_obj.PoolKeyCollection = pool_key_collection
		} else {
			utils.AviLog.Warning.Printf("VS cache not found for key: %v . Unable to update Pool collection", vsKey)
		}
	}
	return
}

func (c *AviObjCache) AviDataScriptPopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName) {
	var rest_response interface{}
	var err error
	var ds_key_collection []NamespaceName
	// TODO Retrieve just fields we care about
	uri := "/api/vsdatascriptset?referred_by=virtualservice:" + vs_uuid
	err = client.AviSession.Get(uri, &rest_response)

	if err != nil {
		utils.AviLog.Warning.Printf("DS Get uri %v returned err %v", uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("DS Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("DS Get uri %v returned %v DSes", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T", resp["results"])
			return
		}
		for _, ds_intf := range results {
			ds, ok := ds_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("ds_intf not of type map[string] interface{}. Instead of type %T", ds_intf)
				continue
			}

			ds_cache_obj := AviDSCache{Name: ds["name"].(string),
				Tenant: tenant, Uuid: ds["uuid"].(string)}

			k := NamespaceName{Namespace: tenant, Name: ds["name"].(string)}

			c.PoolCache.AviCacheAdd(k, &ds_cache_obj)
			ds_key_collection = append(ds_key_collection, k)
			utils.AviLog.Info.Printf("Added DS cache key %v val %v",
				k, ds_cache_obj)

		}
		vs_cache, found := c.VsCache.AviCacheGet(vsKey)
		if found {
			vs_cache_obj, ok := vs_cache.(*AviVsCache)
			if !ok {
				utils.AviLog.Warning.Printf("Unable to cast to VS object: %v", vsKey)
				return
			}
			vs_cache_obj.DSKeyCollection = ds_key_collection
		} else {
			utils.AviLog.Warning.Printf("VS cache not found for key: %v . Unable to update DS collection", vsKey)
		}
	}
	return
}

func (c *AviObjCache) AviCloudPropertiesPopulate(client *clients.AviClient,
	cloud string) {
	vtype := os.Getenv("CLOUD_VTYPE")
	if vtype == "" {
		// Default to vcenter.
		vtype = "CLOUD_VCENTER"
	}
	var rest_response interface{}
	uri := "/api/cloud"
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("CloudProperties Get uri %v returned err %v", uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("CloudProperties Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("CloudProperties Get uri %v returned %v ", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T ", resp["results"])
			return
		}
		for _, cloud_intf := range results {
			cloud_pol, ok := cloud_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("cloud_intf not of type map[string] interface{}. Instead of type %T", cloud_intf)
				continue
			}

			if cloud == cloud_pol["name"] {

				cloud_obj := &AviCloudPropertyCache{Name: cloud, VType: vtype}
				if cloud_pol["dns_provider_ref"] != nil {
					dns_uuid := ExtractDNSUuid(cloud_pol["dns_provider_ref"].(string))
					cloud_obj.NSIpamDNS = c.AviDNSPropertyPopulate(client, dns_uuid)

				} else {
					utils.AviLog.Warning.Printf("Cloud does not have a dns_provider_ref configured %v", cloud)
				}
				c.CloudKeyCache.AviCacheAdd(cloud, cloud_obj)
				utils.AviLog.Info.Printf("Added CloudKeyCache cache key %v val %v",
					cloud, cloud_obj)
			}

		}
	}
}

func (c *AviObjCache) AviDNSPropertyPopulate(client *clients.AviClient,
	nsDNSIpam string) string {
	var rest_response interface{}
	uri := "/api/ipamdnsproviderprofile/"
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("DNSProperty Get uri %v returned err %v", uri, err)
		return ""
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("DNSProperty Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return ""
		}
		utils.AviLog.Info.Printf("DNSProperty Get uri %v returned %v ", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T ", resp["results"])
			return ""
		}
		for _, dns_intf := range results {
			dns_pol, ok := dns_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("dns_intf not of type map[string] interface{}. Instead of type %T", dns_intf)
				continue
			}
			if dns_pol["uuid"] == nsDNSIpam {

				dns_profile := dns_pol["internal_profile"]
				dns_profile_pol, dns_found := dns_profile.(map[string]interface{})
				if dns_found {
					dns_ipam := dns_profile_pol["dns_service_domain"].([]interface{})[0].(map[string]interface{})
					// Pick the first dns profile
					utils.AviLog.Info.Printf("Found DNS_IPAM: %v", dns_ipam["domain_name"])
					return dns_ipam["domain_name"].(string)
				}

			}

		}
	}
	return ""
}

func ExtractDNSUuid(word string) string {
	r, _ := regexp.Compile("ipamdnsproviderprofile-.*")
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])]
	}
	return ""
}
