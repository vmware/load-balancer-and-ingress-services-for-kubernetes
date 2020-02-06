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
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/avinetworks/sdk/go/clients"
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
	"gitlab.eng.vmware.com/orion/akc/pkg/lib"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
)

type AviObjCache struct {
	VsCache         *AviCache
	PgCache         *AviCache
	HTTPPolCache    *AviCache
	DSCache         *AviCache
	PoolCache       *AviCache
	CloudKeyCache   *AviCache
	HTTPPolicyCache *AviCache
	SSLKeyCache     *AviCache
	VSVIPCache      *AviCache
	VrfCache        *AviCache
}

func NewAviObjCache() *AviObjCache {
	c := AviObjCache{}
	c.VsCache = NewAviCache()
	c.PgCache = NewAviCache()
	c.DSCache = NewAviCache()
	c.PoolCache = NewAviCache()
	c.SSLKeyCache = NewAviCache()
	c.CloudKeyCache = NewAviCache()
	c.HTTPPolicyCache = NewAviCache()
	c.VSVIPCache = NewAviCache()
	c.VrfCache = NewAviCache()
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

func VrfChecksum(vrfName string, staticRoutes []*models.StaticRoute) uint32 {
	return (utils.Hash(vrfName) + utils.Hash(utils.Stringify(staticRoutes)))
}

func (c *AviObjCache) AviObjCachePopulate(client *clients.AviClient,
	version string, cloud string) {
	SetTenant := session.SetTenant("*")
	SetTenant(client.AviSession)
	SetVersion := session.SetVersion(version)
	SetVersion(client.AviSession)
	c.AviPopulateAllPGs(client, cloud)
	c.AviPopulateAllDSs(client, cloud)
	c.AviPopulateAllSSLKeys(client, cloud)
	c.AviPopulateAllHttpPolicySets(client, cloud)
	c.AviPopulateAllPools(client, cloud)
	// Populate the VS cache
	c.AviObjVSCachePopulate(client, cloud)
	c.AviCloudPropertiesPopulate(client, cloud)
	c.AviObjVrfCachePopulate(client)
}

func (c *AviObjCache) AviPopulateAllPGs(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var uri string
	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
	} else {
		uri = "/api/poolgroup?include_name=true&cloud_ref.name=" + cloud
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for pg %v", uri, err)
		return
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal pg data, err: %v", err)
		return
	}
	for i := 0; i < result.Count; i++ {
		pg := models.PoolGroup{}
		err = json.Unmarshal(elems[i], &pg)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal pg data, err: %v", err)
			continue
		}

		if pg.Name == nil || pg.UUID == nil || pg.CloudConfigCksum == nil {
			utils.AviLog.Warning.Printf("Incomplete pg data unmarshalled, %s", utils.Stringify(pg))
			continue
		}

		pgCacheObj := AviPGCache{
			Name:             *pg.Name,
			Uuid:             *pg.UUID,
			CloudConfigCksum: *pg.CloudConfigCksum,
		}
		utils.AviLog.Info.Printf("Adding pg to Cache %s\n", *pg.Name)
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: *pg.Name}
		c.PgCache.AviCacheAdd(k, &pgCacheObj)

	}
}

func (c *AviObjCache) AviPopulateAllDSs(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var uri string
	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
	} else {
		uri = "/api/vsdatascriptset?include_name=true&cloud_ref.name=" + cloud
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for datascript %v", uri, err)
		return
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal datascript data, err: %v", err)
		return
	}
	for i := 0; i < result.Count; i++ {
		ds := models.VSDataScriptSet{}
		err = json.Unmarshal(elems[i], &ds)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal datascript data, err: %v", err)
			continue
		}

		dsCacheObj := AviDSCache{
			Name: *ds.Name,
			Uuid: *ds.UUID,
		}
		utils.AviLog.Info.Printf("Adding datascript to Cache %s\n", *ds.Name)
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: *ds.Name}
		c.DSCache.AviCacheAdd(k, &dsCacheObj)

	}
}

func (c *AviObjCache) AviPopulateAllSSLKeys(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var uri string
	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
	} else {
		uri = "/api/sslkeyandcertificate?include_name=true&cloud_ref.name=" + cloud
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for sslkeyandcertificate %v", uri, err)
		return
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal sslkeyandcertificate data, err: %v", err)
		return
	}
	for i := 0; i < result.Count; i++ {
		sslkey := models.SSLKeyAndCertificate{}
		err = json.Unmarshal(elems[i], &sslkey)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal sslkeyandcertificate data, err: %v", err)
			continue
		}

		sslCacheObj := AviSSLCache{
			Name: *sslkey.Name,
			Uuid: *sslkey.UUID,
		}
		utils.AviLog.Info.Printf("Adding sslkeyandcertificate to Cache %s\n", *sslkey.Name)
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: *sslkey.Name}
		c.SSLKeyCache.AviCacheAdd(k, &sslCacheObj)

	}
}

func (c *AviObjCache) AviPopulateAllHttpPolicySets(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var uri string
	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
	} else {
		uri = "/api/httppolicyset?include_name=true&cloud_ref.name=" + cloud
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for httppolicyset %v", uri, err)
		return
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal httppolicyset data, err: %v", err)
		return
	}
	for i := 0; i < result.Count; i++ {
		httppol := models.HTTPPolicySet{}
		err = json.Unmarshal(elems[i], &httppol)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal httppolicyset data, err: %v", err)
			continue
		}

		httpPolCacheObj := AviHTTPPolicyCache{
			Name:             *httppol.Name,
			Uuid:             *httppol.UUID,
			CloudConfigCksum: *httppol.CloudConfigCksum,
		}
		utils.AviLog.Info.Printf("Adding httppolicyset to Cache %s\n", *httppol.Name)
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: *httppol.Name}
		c.HTTPPolCache.AviCacheAdd(k, &httpPolCacheObj)

	}
}

func (c *AviObjCache) AviPopulateAllPools(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var uri string
	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
	} else {
		uri = "/api/pool?include_name=true&cloud_ref.name=" + cloud
	}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err for pool %v", uri, err)
		return
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal pool data, err: %v", err)
		return
	}
	for i := 0; i < result.Count; i++ {
		pool := models.Pool{}
		err = json.Unmarshal(elems[i], &pool)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal pool data, err: %v", err)
			continue
		}

		poolCacheObj := AviPoolCache{
			Name:             *pool.Name,
			Uuid:             *pool.UUID,
			CloudConfigCksum: *pool.CloudConfigCksum,
		}
		utils.AviLog.Info.Printf("Adding pool to Cache %s\n", *pool.Name)
		k := NamespaceName{Namespace: utils.ADMIN_NS, Name: *pool.Name}
		c.PoolCache.AviCacheAdd(k, &poolCacheObj)

	}
}

func (c *AviObjCache) AviObjVrfCachePopulate(client *clients.AviClient) {
	disableStaticRoute := os.Getenv(lib.DISABLE_STATIC_ROUTE_SYNC)
	if disableStaticRoute == "true" {
		utils.AviLog.Info.Printf("Static route sync disabled, skipping vrf cache population")
		return
	}
	uri := "/api/vrfcontext"

	vrfList := []*models.VrfContext{}
	result, err := client.AviSession.GetCollectionRaw(uri)
	if err != nil {
		utils.AviLog.Warning.Printf("Get uri %v returned err %v", uri, err)
		return
	}
	elems := make([]json.RawMessage, result.Count)
	err = json.Unmarshal(result.Results, &elems)
	if err != nil {
		utils.AviLog.Warning.Printf("Failed to unmarshal data, err: %v", err)
		return
	}
	for i := 0; i < result.Count; i++ {
		vrf := models.VrfContext{}
		err = json.Unmarshal(elems[i], &vrf)
		if err != nil {
			utils.AviLog.Warning.Printf("Failed to unmarshal data, err: %v", err)
			continue
		}
		vrfList = append(vrfList, &vrf)

		vrfName := *vrf.Name
		checksum := VrfChecksum(vrfName, vrf.StaticRoutes)
		vrfCacheObj := AviVrfCache{
			Name:             vrfName,
			Uuid:             *vrf.UUID,
			CloudConfigCksum: checksum,
		}
		utils.AviLog.Info.Printf("Adding vrf to Cache %s\n", vrfName)
		c.VrfCache.AviCacheAdd(vrfName, &vrfCacheObj)
	}
}

// TODO (sudswas): Should this be run inside a go routine for parallel population
// to reduce bootup time when the system is loaded. Variable duplication expected.
func (c *AviObjCache) AviObjVSCachePopulate(client *clients.AviClient,
	cloud string, override_uri ...NextPage) {
	var rest_response interface{}
	var uri string
	if len(override_uri) == 1 {
		uri = override_uri[0].Next_uri
	} else {
		uri = "/api/virtualservice?include_name=true&cloud_ref.name=" + cloud + "&vrf_context_ref.name=" + lib.GetVrf()
	}
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
			svc_mdata_intf, ok := vs["service_metadata"]
			var svc_mdata_obj LBServiceMetadataObj
			var svc_mdata interface{}
			var svc_mdata_map map[string]interface{}
			if ok {
				if err := json.Unmarshal([]byte(svc_mdata_intf.(string)),
					&svc_mdata); err == nil {
					svc_mdata_map, ok = svc_mdata.(map[string]interface{})
					if !ok {
						utils.AviLog.Warning.Printf(`resp %v svc_mdata %T has invalid
								 service_metadata type for vs`, vs, svc_mdata)
					} else {
						svcName, ok := svc_mdata_map["svc_name"]
						if ok {
							svc_mdata_obj.ServiceName = svcName.(string)
						} else {
							utils.AviLog.Warning.Printf(`service_metadata %v 
									  malformed for vs`, svc_mdata_map)
						}
						namespace, ok := svc_mdata_map["namespace"]
						if ok {
							svc_mdata_obj.Namespace = namespace.(string)
						} else {
							utils.AviLog.Warning.Printf(`service_metadata %v 
									  malformed for vs`, svc_mdata_map)
						}
					}
				}
			}
			var sni_child_collection []string
			vh_child, found := vs["vh_child_vs_uuid"]
			if found {
				for _, child := range vh_child.([]interface{}) {
					sni_child_collection = append(sni_child_collection, child.(string))
				}

			}
			vs_parent_ref, foundParent := vs["vh_parent_vs_ref"]
			var parentVSKey NamespaceName
			if foundParent {
				vs_uuid := ExtractVsUuid(vs_parent_ref.(string))
				utils.AviLog.Info.Printf("extracted the vs uuid from parent ref during cache population: %s", vs_uuid)
				// Now let's get the VS key from this uuid
				vsKey, gotVS := c.VsCache.AviCacheGetKeyByUuid(vs_uuid)
				if gotVS {
					parentVSKey = vsKey.(NamespaceName)
				}

			}
			if vs["cloud_config_cksum"] != nil {
				k := NamespaceName{Namespace: utils.ADMIN_NS, Name: vs["name"].(string)}
				var vip string
				if vs["vip"] != nil && len(vs["vip"].([]interface{})) > 0 {
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
							vs_cache_obj.ParentVSRef = parentVSKey
							utils.AviLog.Info.Printf("Updated Vs cache k %v val %v",
								k, vs_cache_obj)
						} else {
							// New object
							vs_cache_obj := AviVsCache{Name: vs["name"].(string),
								Tenant: utils.ADMIN_NS, Uuid: vs["uuid"].(string), Vip: vip,
								CloudConfigCksum:   vs["cloud_config_cksum"].(string),
								SNIChildCollection: sni_child_collection,
								ParentVSRef:        parentVSKey,
								ServiceMetadataObj: svc_mdata_obj}

							c.VsCache.AviCacheAdd(k, &vs_cache_obj)
							utils.AviLog.Info.Printf("Added Vs cache k %v val %v",
								k, vs_cache_obj)
						}
					} else {
						// New object
						vs_cache_obj := AviVsCache{Name: vs["name"].(string),
							Tenant: utils.ADMIN_NS, Uuid: vs["uuid"].(string), Vip: vip,
							CloudConfigCksum:   vs["cloud_config_cksum"].(string),
							SNIChildCollection: sni_child_collection,
							ParentVSRef:        parentVSKey,
							ServiceMetadataObj: svc_mdata_obj}

						c.VsCache.AviCacheAdd(k, &vs_cache_obj)
						utils.AviLog.Info.Printf("Added Vs cache k %v val %v",
							k, vs_cache_obj)
					}
				} else {
					vs_cache_obj := AviVsCache{Name: vs["name"].(string),
						Tenant: utils.ADMIN_NS, Uuid: vs["uuid"].(string), Vip: vip,
						CloudConfigCksum:   vs["cloud_config_cksum"].(string),
						SNIChildCollection: sni_child_collection,
						ParentVSRef:        parentVSKey,
						ServiceMetadataObj: svc_mdata_obj}

					c.VsCache.AviCacheAdd(k, &vs_cache_obj)
					utils.AviLog.Info.Printf("Added Vs cache k %v val %v",
						k, vs_cache_obj)
				}
				c.AviHTTPolicyCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
				c.AviPGCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
				c.AviPoolCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
				c.AviDataScriptPopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
				c.AviSSLKeyCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
				c.AviVSVIPCachePopulate(client, cloud, vs["uuid"].(string), utils.ADMIN_NS, k)
			}
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/virtualservice")
			utils.AviLog.Info.Printf("Found next page for vs, uri: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/virtualservice" + next_uri[1]
				utils.AviLog.Info.Printf("Next page uri for vs: %s", override_uri)
				nextPage := NextPage{Next_uri: override_uri}
				c.AviObjVSCachePopulate(client, cloud, nextPage)
			}
		}
	}
}

//Design library methods to remove repeatation of code.
func (c *AviObjCache) AviPGCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName, nextPage ...NextPage) {
	var rest_response interface{}

	var pg_key_collection []NamespaceName
	var uri string
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
		pg_key_collection = nextPage[0].Collection
	} else {
		uri = "/api/poolgroup?include_name=true&cloud_ref.name=" + cloud + "&referred_by=virtualservice:" + vs_uuid
	}
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
				k := NamespaceName{Namespace: tenant, Name: pg["name"].(string)}
				utils.AviLog.Info.Printf("Added PG cache key %v to VS", k)
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
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/poolgroup")
			utils.AviLog.Info.Printf("Found next page for pg, uri: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/poolgroup" + next_uri[1]
				utils.AviLog.Info.Printf("Next page uri for pg: %s", override_uri)
				nextPage := NextPage{Next_uri: override_uri}
				c.AviPGCachePopulate(client, cloud, vs_uuid, tenant, vsKey, nextPage)
			}
		}
	}
}

func (c *AviObjCache) AviVSVIPCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName, nextPage ...NextPage) {
	var rest_response interface{}

	var vsvip_key_collection []NamespaceName
	var uri string
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
		vsvip_key_collection = nextPage[0].Collection
	} else {
		uri = "/api/vsvip?include_name=true&cloud_ref.name=" + cloud + "&referred_by=virtualservice:" + vs_uuid
	}
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("VSVIP Get uri %v returned err %v", uri, err)
		return
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("VSVIP Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("VSVIP Get uri %v returned %v VSVIPs", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T for VSVIP", resp["results"])
			return
		}
		for _, vsvip_intf := range results {
			vsvip, ok := vsvip_intf.(map[string]interface{})
			var fqdns []string
			if !ok {
				utils.AviLog.Warning.Printf("vsvip_intf not of type map[string] interface{}. Instead of type %T", vsvip_intf)
				continue
			}
			if vsvip["dns_info"] != nil {
				for _, aRecord := range vsvip["dns_info"].([]interface{}) {
					aRecordMap, success := aRecord.(map[string]interface{})
					if success {
						fqdn, ok := aRecordMap["fqdn"].(string)
						if ok {
							fqdns = append(fqdns, fqdn)
						}
					}
				}
			}
			vsvip_cache_obj := AviVSVIPCache{Name: vsvip["name"].(string),
				Tenant: tenant, Uuid: vsvip["uuid"].(string), FQDNs: fqdns,
			}
			k := NamespaceName{Namespace: tenant, Name: vsvip["name"].(string)}
			c.VSVIPCache.AviCacheAdd(k, &vsvip_cache_obj)
			utils.AviLog.Info.Printf("Added VSVIP cache key %v val %v",
				k, vsvip_cache_obj)
			vsvip_key_collection = append(vsvip_key_collection, k)

		}
		vs_cache, found := c.VsCache.AviCacheGet(vsKey)
		if found {
			vs_cache_obj, ok := vs_cache.(*AviVsCache)
			if !ok {
				utils.AviLog.Warning.Printf("Unable to cast to VS object: %v", vsKey)
				return
			}
			vs_cache_obj.VSVipKeyCollection = vsvip_key_collection
		} else {
			utils.AviLog.Warning.Printf("VS cache not found for key: %v . Unable to update VSVIP collection", vsKey)
			return
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/vsvip")
			utils.AviLog.Info.Printf("Found next page for vsvip, uri: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/vsvip" + next_uri[1]
				utils.AviLog.Info.Printf("Next page uri for vsvip: %s", override_uri)
				nextPage := NextPage{Next_uri: override_uri}
				c.AviVSVIPCachePopulate(client, cloud, vs_uuid, tenant, vsKey, nextPage)
			}
		}
	}
}

func (c *AviObjCache) AviHTTPolicyCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName, nextPage ...NextPage) {
	var rest_response interface{}

	var http_pol_key_collection []NamespaceName
	var uri string
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
		http_pol_key_collection = nextPage[0].Collection
	} else {
		uri = "/api/httppolicyset?include_name=true&cloud_ref.name=" + cloud + "&referred_by=virtualservice:" + vs_uuid
	}
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("HTTP Policy Get uri %v returned err %v", uri, err)
		return
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("HTTP Policy Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("HTTP Policy Get uri %v returned %v PGs", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T for HTTP Policy", resp["results"])
			return
		}
		for _, pol_intf := range results {
			pol, ok := pol_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("http_intf not of type map[string] interface{}. Instead of type %T", pol_intf)
				continue
			}

			k := NamespaceName{Namespace: tenant, Name: pol["name"].(string)}
			utils.AviLog.Info.Printf("Added HTTP Policy cache key %v",
				k)
			http_pol_key_collection = append(http_pol_key_collection, k)

		}
		vs_cache, found := c.VsCache.AviCacheGet(vsKey)
		if found {
			vs_cache_obj, ok := vs_cache.(*AviVsCache)
			if !ok {
				utils.AviLog.Warning.Printf("Unable to cast to VS object: %v", vsKey)
				return
			}
			vs_cache_obj.HTTPKeyCollection = http_pol_key_collection
		} else {
			utils.AviLog.Warning.Printf("VS cache not found for key: %v . Unable to update HTTP Policy collection", vsKey)
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/httppolicyset")
			utils.AviLog.Info.Printf("Found next page for http policyset objs, uri: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/httppolicyset" + next_uri[1]
				utils.AviLog.Info.Printf("Next page uri for http policyset objs: %s", override_uri)
				nextPage := NextPage{Next_uri: override_uri}
				c.AviHTTPolicyCachePopulate(client, cloud, vs_uuid, tenant, vsKey, nextPage)
			}
		}
	}
}

func (c *AviObjCache) AviSSLKeyCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName, nextPage ...NextPage) {
	var rest_response interface{}

	var ssl_key_collection []NamespaceName
	var uri string
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
		ssl_key_collection = nextPage[0].Collection
	} else {
		uri = "/api/sslkeyandcertificate?include_name=true&cloud_ref.name=" + cloud + "&referred_by=virtualservice:" + vs_uuid
	}
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		utils.AviLog.Warning.Printf("SSL Keys Get uri %v returned err %v", uri, err)
		return
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("SSL Keys Get uri %v returned %v type %T", uri,
				rest_response, rest_response)
			return
		}
		utils.AviLog.Info.Printf("SSL Keys Get uri %v returned %v PGs", uri,
			resp["count"])
		results, ok := resp["results"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("results not of type []interface{} Instead of type %T for SSL Keys", resp["results"])
			return
		}
		for _, ssl_intf := range results {
			ssl, ok := ssl_intf.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("ssl_intf not of type map[string] interface{}. Instead of type %T", ssl_intf)
				continue
			}

			k := NamespaceName{Namespace: tenant, Name: ssl["name"].(string)}
			utils.AviLog.Info.Printf("Added SSL Key cache key %v",
				k)
			ssl_key_collection = append(ssl_key_collection, k)

		}
		vs_cache, found := c.VsCache.AviCacheGet(vsKey)
		if found {
			vs_cache_obj, ok := vs_cache.(*AviVsCache)
			if !ok {
				utils.AviLog.Warning.Printf("Unable to cast to VS object: %v", vsKey)
				return
			}
			vs_cache_obj.SSLKeyCertCollection = ssl_key_collection
		} else {
			utils.AviLog.Warning.Printf("VS cache not found for key: %v . Unable to update HTTP Policy collection", vsKey)
		}
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/sslkeyandcertificate")
			utils.AviLog.Info.Printf("Found next page for ssl key objs, uri: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/sslkeyandcertificate" + next_uri[1]
				utils.AviLog.Info.Printf("Next page uri for ssl key objs: %s", override_uri)
				nextPage := NextPage{Next_uri: override_uri}
				c.AviHTTPolicyCachePopulate(client, cloud, vs_uuid, tenant, vsKey, nextPage)
			}
		}
	}
}

func (c *AviObjCache) AviPoolCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string, tenant string, vsKey NamespaceName, nextPage ...NextPage) {
	var rest_response interface{}
	var err error
	var pool_key_collection []NamespaceName
	var uri string
	if len(nextPage) == 1 {
		uri = nextPage[0].Next_uri
		pool_key_collection = nextPage[0].Collection
	} else {
		uri = "/api/pool?include_name=true&cloud_ref.name=" + cloud + "&referred_by=virtualservice:" + vs_uuid
	}
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
			svc_mdata_intf, ok := pool["service_metadata"]
			var svc_mdata_obj ServiceMetadataObj
			var svc_mdata interface{}
			var svc_mdata_map map[string]interface{}
			if ok {
				if err := json.Unmarshal([]byte(svc_mdata_intf.(string)),
					&svc_mdata); err == nil {
					svc_mdata_map, ok = svc_mdata.(map[string]interface{})
					if !ok {
						utils.AviLog.Warning.Printf(`resp %v svc_mdata %T has invalid
								 service_metadata type`, pool, svc_mdata)
					} else {
						ingressName, ok := svc_mdata_map["ingress_name"]
						if ok {
							svc_mdata_obj.IngressName = ingressName.(string)
						} else {
							utils.AviLog.Warning.Printf(`service_metadata %v 
									  malformed`, svc_mdata_map)
						}
						namespace, ok := svc_mdata_map["namespace"]
						if ok {
							svc_mdata_obj.Namespace = namespace.(string)
						} else {
							utils.AviLog.Warning.Printf(`service_metadata %v 
									  malformed`, svc_mdata_map)
						}
					}
				}
			} else {
				utils.AviLog.Warning.Printf("service_metadata %v malformed", pool)
				// Not caching a pool with malformed metadata?
				continue
			}

			k := NamespaceName{Namespace: tenant, Name: pool["name"].(string)}
			pool_key_collection = append(pool_key_collection, k)
			utils.AviLog.Info.Printf("Added Pool cache key %v val.",
				k)

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
		if resp["next"] != nil {
			// It has a next page, let's recursively call the same method.
			next_uri := strings.Split(resp["next"].(string), "/api/pool")
			utils.AviLog.Info.Printf("Found next page, uri for pool: %s", next_uri)
			if len(next_uri) > 1 {
				override_uri := "/api/pool" + next_uri[1]
				utils.AviLog.Info.Printf("Next page uri for pool: %s", override_uri)
				nextPage := NextPage{Next_uri: override_uri, Collection: pool_key_collection}
				c.AviPoolCachePopulate(client, cloud, vs_uuid, tenant, vsKey, nextPage)
			}
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

			k := NamespaceName{Namespace: tenant, Name: ds["name"].(string)}

			ds_key_collection = append(ds_key_collection, k)
			utils.AviLog.Info.Printf("Added DS cache key %v ", k)
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
					dns_uuid := ExtractPattern(cloud_pol["dns_provider_ref"].(string), "ipamdnsproviderprofile-.*")
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

func ExtractPattern(word string, pattern string) string {
	r, _ := regexp.Compile(pattern)
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])]
	}
	return ""
}

func ExtractVsUuid(word string) string {
	r, _ := regexp.Compile("virtualservice-.*.#")
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][:len(result[0])-1]
	}
	return ""
}
