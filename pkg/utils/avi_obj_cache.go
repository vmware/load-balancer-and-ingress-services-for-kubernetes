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

package utils

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/avinetworks/sdk/go/clients"
	"github.com/avinetworks/sdk/go/session"
)

type AviObjCache struct {
	VsCache         *AviCache
	PgCache         *AviCache
	HTTPCache       *AviCache
	SSLKeyCache     *AviCache
	PkiProfileCache *AviCache
	CloudKeyCache   *AviCache
	PoolCache       *AviCache
	SvcToPoolCache  *AviMultiCache
}

func NewAviObjCache() *AviObjCache {
	c := AviObjCache{}
	c.VsCache = NewAviCache()
	c.PgCache = NewAviCache()
	c.HTTPCache = NewAviCache()
	c.PoolCache = NewAviCache()
	c.PkiProfileCache = NewAviCache()
	c.SSLKeyCache = NewAviCache()
	c.CloudKeyCache = NewAviCache()
	c.SvcToPoolCache = NewAviMultiCache()
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

func (c *AviObjCache) AviPoolCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string) []NamespaceName {
	var rest_response interface{}
	var svc_mdata_obj ServiceMetadataObj
	var svc_mdata interface{}
	var svc_mdata_map map[string]interface{}
	var err error
	//var pool_name string
	var pool_key_collection []NamespaceName
	// TODO Retrieve just fields we care about
	uri := "/api/pool?include_name=true&cloud_ref.name=" + cloud + "&referred_by=virtualservice:" + vs_uuid
	err = client.AviSession.Get(uri, &rest_response)

	if err != nil {
		AviLog.Warnf(`Pool Get uri %v returned err %v`, uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			AviLog.Warnf(`Pool Get uri %v returned %v type %T`, uri,
				rest_response, rest_response)
		} else {
			AviLog.Infof("Pool Get uri %v returned %v pools", uri,
				resp["count"])
			results, ok := resp["results"].([]interface{})
			if !ok {
				AviLog.Warnf(`results not of type []interface{}
								 Instead of type %T`, resp["results"])
				return nil
			}
			for _, pool_intf := range results {
				pool, ok := pool_intf.(map[string]interface{})
				if !ok {
					AviLog.Warnf(`pool_intf not of type map[string]
									 interface{}. Instead of type %T`, pool_intf)
					continue
				}
				svc_mdata_intf, ok := pool["service_metadata"]
				if ok {
					if err := json.Unmarshal([]byte(svc_mdata_intf.(string)),
						&svc_mdata); err == nil {
						svc_mdata_map, ok = svc_mdata.(map[string]interface{})
						if !ok {
							AviLog.Warnf(`resp %v svc_mdata %T has invalid
									 service_metadata type`, pool, svc_mdata)
						} else {
							crkhey, ok := svc_mdata_map["crud_hash_key"]
							if ok {
								svc_mdata_obj.CrudHashKey = crkhey.(string)
							} else {
								AviLog.Warnf(`service_metadata %v 
										  malformed`, svc_mdata_map)
							}
						}
					}
				} else {
					AviLog.Warnf("service_metadata %v malformed", pool)
					// Not caching a pool with malformed metadata?
					continue
				}

				var tenant string
				url, err := url.Parse(pool["tenant_ref"].(string))
				if err != nil {
					AviLog.Warnf(`Error parsing tenant_ref %v in 
											   pool %v`, pool["tenant_ref"], pool)
					continue
				} else if url.Fragment == "" {
					AviLog.Warnf(`Error extracting name tenant_ref %v 
										 in pool %v`, pool["tenant_ref"], pool)
					continue
				} else {
					tenant = url.Fragment
				}

				pool_cache_obj := AviPoolCache{Name: pool["name"].(string),
					Tenant: tenant, Uuid: pool["uuid"].(string),
					LbAlgorithm:      pool["lb_algorithm"].(string),
					CloudConfigCksum: pool["cloud_config_cksum"].(string),
					ServiceMetadata:  svc_mdata_obj}

				k := NamespaceName{Namespace: tenant, Name: pool["name"].(string)}

				c.PoolCache.AviCacheAdd(k, &pool_cache_obj)
				pool_key_collection = append(pool_key_collection, k)
				AviLog.Infof("Added Pool cache key %v val %v",
					k, pool_cache_obj)
			}
		}
	}
	return pool_key_collection
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
	c.IstioMutualSSLKeyCert(client, cloud)
	c.IstioMutualPkiProfile(client, cloud)

}

// TODO (sudswas): Should this be run inside a go routine for parallel population
// to reduce bootup time when the system is loaded. Variable duplication expected.
func (c *AviObjCache) AviObjVSCachePopulate(client *clients.AviClient,
	cloud string) {
	var rest_response interface{}
	var svc_mdata interface{}
	var svc_mdata_map map[string]interface{}
	var svc_mdata_obj ServiceMetadataObj
	// TODO Retrieve just fields we care about
	uri := "/api/virtualservice?include_name=true&cloud_ref.name=" + cloud
	err := client.AviSession.Get(uri, &rest_response)

	if err != nil {
		AviLog.Warnf(`Vs Get uri %v returned err %v`, uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			AviLog.Warnf(`Vs Get uri %v returned %v type %T`, uri,
				rest_response, rest_response)
		} else {
			AviLog.Infof("Vs Get uri %v returned %v vses", uri,
				resp["count"])
			results, ok := resp["results"].([]interface{})
			if !ok {
				AviLog.Warnf(`results not of type []interface{}
							 Instead of type %T`, resp["results"])
				return
			}
			for _, vs_intf := range results {
				vs, ok := vs_intf.(map[string]interface{})
				if !ok {
					AviLog.Warnf(`vs_intf not of type map[string]
								 interface{}. Instead of type %T`, vs_intf)
					continue
				}
				svc_mdata_intf, ok := vs["service_metadata"]
				if ok {
					if err := json.Unmarshal([]byte(svc_mdata_intf.(string)),
						&svc_mdata); err == nil {
						svc_mdata_map, ok = svc_mdata.(map[string]interface{})
						if !ok {
							AviLog.Warnf(`resp %v svc_mdata %T has invalid
								 service_metadata type`, vs, svc_mdata)
						} else {
							crkhey, ok := svc_mdata_map["crud_hash_key"]
							if ok {
								svc_mdata_obj.CrudHashKey = crkhey.(string)
							} else {
								AviLog.Warnf(`service_metadata %v 
									  malformed`, svc_mdata_map)
							}
						}
					}
				}
				var tenant string
				url, err := url.Parse(vs["tenant_ref"].(string))
				if err != nil {
					AviLog.Warnf(`Error parsing tenant_ref %v in 
										   vs %v`, vs["tenant_ref"], vs)
					continue
				} else if url.Fragment == "" {
					AviLog.Warnf(`Error extracting name tenant_ref %v 
									 in vs %v`, vs["tenant_ref"], vs)
					continue
				} else {
					tenant = url.Fragment
				}
				pg_key_collection := c.AviPGCachePopulate(client, cloud, vs["uuid"].(string))
				pool_key_collection := c.AviPoolCachePopulate(client, cloud, vs["uuid"].(string))
				http_policy_collection := c.AviHTTPPolicyCachePopulate(client, cloud, vs["uuid"].(string))
				ssl_key_collection := c.AviSSLKeyAndCertPopulate(client, cloud, vs["uuid"].(string))
				var sni_child_collection []string
				vh_child, found := vs["vh_child_vs_uuid"]
				if found {
					for _, child := range vh_child.([]interface{}) {
						sni_child_collection = append(sni_child_collection, child.(string))
					}

				}
				if vs["cloud_config_cksum"] != nil {
					vs_cache_obj := AviVsCache{Name: vs["name"].(string),
						Tenant: tenant, Uuid: vs["uuid"].(string), Vip: nil,
						CloudConfigCksum: vs["cloud_config_cksum"].(string),
						ServiceMetadata:  svc_mdata_obj, PGKeyCollection: pg_key_collection, PoolKeyCollection: pool_key_collection, HTTPKeyCollection: http_policy_collection,
						SSLKeyCertCollection: ssl_key_collection, SNIChildCollection: sni_child_collection}
					k := NamespaceName{Namespace: tenant, Name: vs["name"].(string)}
					c.VsCache.AviCacheAdd(k, &vs_cache_obj)

					AviLog.Infof("Added Vs cache k %v val %v",
						k, vs_cache_obj)
				}
			}
		}
	}
}

//Design library methods to remove repeatation of code.
func (c *AviObjCache) AviPGCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string) []NamespaceName {
	var rest_response interface{}
	var svc_mdata interface{}
	var svc_mdata_map map[string]interface{}
	var svc_mdata_obj ServiceMetadataObj
	var pg_key_collection []NamespaceName
	uri := "/api/poolgroup?include_name=true&cloud_ref.name=" + cloud + "&referred_by=virtualservice:" + vs_uuid
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		AviLog.Warnf(`PG Get uri %v returned err %v`, uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			AviLog.Warnf(`PG Get uri %v returned %v type %T`, uri,
				rest_response, rest_response)
		} else {
			AviLog.Infof("PG Get uri %v returned %v PGs", uri,
				resp["count"])
			results, ok := resp["results"].([]interface{})
			if !ok {
				AviLog.Warnf(`results not of type []interface{}
								 Instead of type %T for PGs`, resp["results"])
				return nil
			}
			for _, pg_intf := range results {
				pg, ok := pg_intf.(map[string]interface{})
				if !ok {
					AviLog.Warnf(`pg_intf not of type map[string]
									 interface{}. Instead of type %T`, pg_intf)
					continue
				}
				svc_mdata_intf, ok := pg["service_metadata"]
				if ok {
					if err := json.Unmarshal([]byte(svc_mdata_intf.(string)),
						&svc_mdata); err == nil {
						svc_mdata_map, ok = svc_mdata.(map[string]interface{})
						if !ok {
							AviLog.Warnf(`resp %v svc_mdata %T has invalid
									 service_metadata type for PGs`, pg, svc_mdata)
						} else {
							crkhey, ok := svc_mdata_map["crud_hash_key"]
							if ok {
								svc_mdata_obj.CrudHashKey = crkhey.(string)
							} else {
								AviLog.Warnf(`service_metadata %v
										  malformed`, svc_mdata_map)
							}
						}
					}
				}

				var tenant string
				url, err := url.Parse(pg["tenant_ref"].(string))
				if err != nil {
					AviLog.Warnf(`Error parsing tenant_ref %v in
											   PG %v`, pg["tenant_ref"], pg)
					continue
				} else if url.Fragment == "" {
					AviLog.Warnf(`Error extracting name tenant_ref %v
										 in PG %v`, pg["tenant_ref"], pg)
					continue
				} else {
					tenant = url.Fragment
				}

				pg_cache_obj := AviPGCache{Name: pg["name"].(string),
					Tenant: tenant, Uuid: pg["uuid"].(string),
					CloudConfigCksum: pg["cloud_config_cksum"].(string),
					ServiceMetadata:  svc_mdata_obj}
				k := NamespaceName{Namespace: tenant, Name: pg["name"].(string)}
				c.PgCache.AviCacheAdd(k, &pg_cache_obj)
				AviLog.Infof("Added PG cache key %v val %v",
					k, pg_cache_obj)
				pg_key_collection = append(pg_key_collection, k)
			}
		}
	}
	return pg_key_collection
}

func (c *AviObjCache) AviHTTPPolicyCachePopulate(client *clients.AviClient,
	cloud string, vs_uuid string) []NamespaceName {
	var rest_response interface{}
	var http_key_collection []NamespaceName
	uri := "/api/httppolicyset?include_name=true&referred_by=virtualservice:" + vs_uuid
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		AviLog.Warnf(`HTTPPolicySet Get uri %v returned err %v`, uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			AviLog.Warnf(`HTTPPolicySet Get uri %v returned %v type %T`, uri,
				rest_response, rest_response)
		} else {
			AviLog.Infof("HTTPPolicySet Get uri %v returned %v HTTP Policies", uri,
				resp["count"])
			results, ok := resp["results"].([]interface{})
			if !ok {
				AviLog.Warnf(`results not of type []interface{}
								 Instead of type %T for HTTP Policies`, resp["results"])
				return nil
			}
			for _, http_intf := range results {
				http_pol, ok := http_intf.(map[string]interface{})
				if !ok {
					AviLog.Warnf(`http_intf not of type map[string]
									 interface{}. Instead of type %T`, http_intf)
					continue
				}

				var tenant string
				url, err := url.Parse(http_pol["tenant_ref"].(string))
				if err != nil {
					AviLog.Warnf(`Error parsing tenant_ref %v in
											   HTTP Policy %v`, http_pol["tenant_ref"], http_pol)
					continue
				} else if url.Fragment == "" {
					AviLog.Warnf(`Error extracting name tenant_ref %v
										 in HTTP Policy set %v`, http_pol["tenant_ref"], http_pol)
					continue
				} else {
					tenant = url.Fragment
				}
				if http_pol != nil {
					http_cache_obj := AviHTTPCache{Name: http_pol["name"].(string),
						Tenant: tenant, Uuid: http_pol["uuid"].(string)}
					if http_pol["cloud_config_cksum"] != nil {
						http_cache_obj.CloudConfigCksum = http_pol["cloud_config_cksum"].(string)
					}
					k := NamespaceName{Namespace: tenant, Name: http_pol["name"].(string)}
					c.HTTPCache.AviCacheAdd(k, &http_cache_obj)
					AviLog.Infof("Added HTTP Policy cache key %v val %v",
						k, http_cache_obj)
					http_key_collection = append(http_key_collection, k)
				}
			}
		}
	}
	return http_key_collection
}

func (c *AviObjCache) AviSSLKeyAndCertPopulate(client *clients.AviClient,
	cloud string, vs_uuid string) []NamespaceName {
	var rest_response interface{}
	var ssl_key_collection []NamespaceName
	uri := "/api/sslkeyandcertificate?include_name=true&referred_by=virtualservice:" + vs_uuid
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		AviLog.Warnf(`SSLKeyAndCert Get uri %v returned err %v`, uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			AviLog.Warnf(`SSLKeyAndCert Get uri %v returned %v type %T`, uri,
				rest_response, rest_response)
		} else {
			AviLog.Infof("SSLKeyAndCert Get uri %v returned %v SSLKeys.", uri,
				resp["count"])
			results, ok := resp["results"].([]interface{})
			if !ok {
				AviLog.Warnf(`results not of type []interface{}
								 Instead of type %T for SSLKeys`, resp["results"])
				return nil
			}
			for _, ssl_intf := range results {
				ssl_pol, ok := ssl_intf.(map[string]interface{})
				if !ok {
					AviLog.Warnf(`ssl_intf not of type map[string]
									 interface{}. Instead of type %T`, ssl_intf)
					continue
				}

				var tenant string
				url, err := url.Parse(ssl_pol["tenant_ref"].(string))
				if err != nil {
					AviLog.Warnf(`Error parsing tenant_ref %v in
					SSLKeyAndCert %v`, ssl_pol["tenant_ref"], ssl_pol)
					continue
				} else if url.Fragment == "" {
					AviLog.Warnf(`Error extracting name tenant_ref %v
										 in SSLKeyAndCert set %v`, ssl_pol["tenant_ref"], ssl_pol)
					continue
				} else {
					AviLog.Infof("URL FRAGMENT :%s tenant_ref: %s", url, ssl_pol["tenant_ref"])
					tenant = url.Fragment
				}
				if ssl_pol != nil {
					ssl_cache_obj := AviSSLCache{Name: ssl_pol["name"].(string),
						Tenant: tenant, Uuid: ssl_pol["uuid"].(string)}
					k := NamespaceName{Namespace: tenant, Name: ssl_pol["name"].(string)}
					c.SSLKeyCache.AviCacheAdd(k, &ssl_cache_obj)
					AviLog.Infof("Added SSLKeyAndCert cache key %v val %v",
						k, ssl_cache_obj)
					ssl_key_collection = append(ssl_key_collection, k)
				}
			}
		}
	}
	return ssl_key_collection
}

func (c *AviObjCache) IstioMutualSSLKeyCert(client *clients.AviClient,
	cloud string) {
	var rest_response interface{}
	uri := "/api/sslkeyandcertificate?include_name=true"
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		AviLog.Warnf(`IstioMutualSSLKeyCert Get uri %v returned err %v`, uri, err)
		return
	}
	resp, ok := rest_response.(map[string]interface{})
	if !ok {
		AviLog.Warnf(`IstioMutualSSLKeyCert Get uri %v returned %v type %T`, uri,
			rest_response, rest_response)
		return
	}
	AviLog.Infof("IstioMutualSSLKeyCert Get uri %v returned %v SSLKeys", uri,
		resp["count"])
	results, ok := resp["results"].([]interface{})
	if !ok {
		AviLog.Warnf(`results not of type []interface{}
								 Instead of type %T for SSLKeys`, resp["results"])
	}
	for _, ssl_intf := range results {
		ssl_pol, ok := ssl_intf.(map[string]interface{})
		if !ok {
			AviLog.Warnf(`ssl_intf not of type map[string]
									 interface{}. Instead of type %T`, ssl_intf)
			continue
		}
		if !strings.Contains(ssl_pol["name"].(string), "istio.default") {
			// Don't parse non-istio.default secrets.
			AviLog.Infof("Skipping SSL Key cert with name :%s", ssl_pol["name"].(string))
			continue
		}
		var tenant string
		url, err := url.Parse(ssl_pol["tenant_ref"].(string))
		if err != nil {
			AviLog.Warnf(`Error parsing tenant_ref %v in
					IstioMutualSSLKeyCert %v`, ssl_pol["tenant_ref"], ssl_pol)
			continue
		} else if url.Fragment == "" {
			AviLog.Warnf(`Error extracting name tenant_ref %v
										 in IstioMutualSSLKeyCert set %v`, ssl_pol["tenant_ref"], ssl_pol)
			continue
		} else {
			tenant = url.Fragment
		}
		if ssl_pol != nil {
			ssl_cache_obj := AviSSLCache{Name: ssl_pol["name"].(string),
				Tenant: tenant, Uuid: ssl_pol["uuid"].(string)}
			k := NamespaceName{Namespace: tenant, Name: ssl_pol["name"].(string)}
			c.SSLKeyCache.AviCacheAdd(k, &ssl_cache_obj)
			AviLog.Infof("Added IstioMutualSSLKeyCert cache key %v val %v",
				k, ssl_cache_obj)
		}

	}

}

func (c *AviObjCache) IstioMutualPkiProfile(client *clients.AviClient,
	cloud string) {
	var rest_response interface{}
	uri := "/api/pkiprofile?include_name=true"
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		AviLog.Warnf(`IstioMutualPkiProfile Get uri %v returned err %v`, uri, err)
		return
	}
	resp, ok := rest_response.(map[string]interface{})
	if !ok {
		AviLog.Warnf(`IstioMutualPkiProfile Get uri %v returned %v type %T`, uri,
			rest_response, rest_response)
		return
	}
	AviLog.Infof("IstioMutualPkiProfile Get uri %v returned %v PkiProfile", uri,
		resp["count"])
	results, ok := resp["results"].([]interface{})
	if !ok {
		AviLog.Warnf(`results not of type []interface{}
								 Instead of type %T for PkiProfile`, resp["results"])
	}
	for _, pki_intf := range results {
		pki_pro, ok := pki_intf.(map[string]interface{})
		if !ok {
			AviLog.Warnf(`pki_intf not of type map[string]
									 interface{}. Instead of type %T`, pki_intf)
			continue
		}

		var tenant string
		url, err := url.Parse(pki_pro["tenant_ref"].(string))
		if err != nil {
			AviLog.Warnf(`Error parsing tenant_ref %v in
					IstioMutualPkiProfile %v`, pki_pro["tenant_ref"], pki_pro)
			continue
		} else if url.Fragment == "" {
			AviLog.Warnf(`Error extracting name tenant_ref %v
										 in IstioMutualPkiProfile set %v`, pki_pro["tenant_ref"], pki_pro)
			continue
		} else {
			tenant = url.Fragment
		}
		if pki_pro != nil {
			pki_cache_obj := AviPkiProfileCache{Name: pki_pro["name"].(string),
				Tenant: tenant, Uuid: pki_pro["uuid"].(string)}
			k := NamespaceName{Namespace: tenant, Name: pki_pro["name"].(string)}
			c.PkiProfileCache.AviCacheAdd(k, &pki_cache_obj)
			AviLog.Infof("Added IstioMutualPkiProfile cache key %v val %v",
				k, pki_cache_obj)
		}
	}

}

func (c *AviObjCache) AviCloudPropertiesPopulate(client *clients.AviClient,
	cloud string) {
	var rest_response interface{}
	uri := "/api/cloud"
	err := client.AviSession.Get(uri, &rest_response)
	if err != nil {
		AviLog.Warnf(`CloudProperties Get uri %v returned err %v`, uri, err)
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			AviLog.Warnf(`CloudProperties Get uri %v returned %v type %T`, uri,
				rest_response, rest_response)
		} else {
			AviLog.Infof("CloudProperties Get uri %v returned %v ", uri,
				resp["count"])
			results, ok := resp["results"].([]interface{})
			if !ok {
				AviLog.Warnf(`results not of type []interface{}
								 Instead of type %T `, resp["results"])
			}
			for _, cloud_intf := range results {
				cloud_pol, ok := cloud_intf.(map[string]interface{})
				if !ok {
					AviLog.Warnf(`cloud_intf not of type map[string]
									 interface{}. Instead of type %T`, cloud_intf)
					continue
				}

				if cloud == cloud_pol["name"] {

					cloud_obj := &AviCloudPropertyCache{Name: cloud, VType: "CLOUD_OSHIFT_K8S"}
					if cloud_pol["dns_provider_ref"] != nil {
						dns_uuid := ExtractDNSUuid(cloud_pol["dns_provider_ref"].(string))
						cloud_obj.NSIpamDNS = c.AviDNSPropertyPopulate(client, dns_uuid)
					}
					c.CloudKeyCache.AviCacheAdd(cloud, cloud_obj)
					AviLog.Infof("Added CloudKeyCache cache key %v val %v",
						cloud, cloud_obj)
				}
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
		AviLog.Warnf(`DNSProperty Get uri %v returned err %v`, uri, err)
		return ""
	} else {
		resp, ok := rest_response.(map[string]interface{})
		if !ok {
			AviLog.Warnf(`DNSProperty Get uri %v returned %v type %T`, uri,
				rest_response, rest_response)
		} else {
			AviLog.Infof("DNSProperty Get uri %v returned %v ", uri,
				resp["count"])
			results, ok := resp["results"].([]interface{})
			if !ok {
				AviLog.Warnf(`results not of type []interface{}
								 Instead of type %T `, resp["results"])
			}
			for _, dns_intf := range results {
				dns_pol, ok := dns_intf.(map[string]interface{})
				if !ok {
					AviLog.Warnf(`dns_intf not of type map[string]
									 interface{}. Instead of type %T`, dns_intf)
					continue
				}
				if dns_pol["uuid"] == nsDNSIpam {

					dns_profile := dns_pol["internal_profile"]
					dns_profile_pol, dns_found := dns_profile.(map[string]interface{})
					if dns_found {
						dns_ipam := dns_profile_pol["dns_service_domain"].([]interface{})[0].(map[string]interface{})
						// Pick the first dns profile
						AviLog.Infof("Found DNS_IPAM: %v", dns_ipam["domain_name"])
						return dns_ipam["domain_name"].(string)
					}

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
