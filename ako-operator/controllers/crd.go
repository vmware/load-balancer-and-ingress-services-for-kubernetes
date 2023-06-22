package controllers

import (
	"context"
	"reflect"

	logr "github.com/go-logr/logr"
	"google.golang.org/protobuf/proto"

	akov1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1alpha1"

	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	CRDGroup   = "ako.vmware.com"
	CRDVersion = "apiextensions.k8s.io/v1"
	Version    = "v1alpha1"
)

var (
	HostRuleFullCRDName        = "hostrules.ako.vmware.com"
	HostRuleCRDSingular        = "hostrule"
	HostRuleCRDPlural          = "hostrules"
	HttpRuleFullCRDName        = "httprules.ako.vmware.com"
	HttpRuleCRDSingular        = "httprule"
	HttpRuleCRDPlural          = "httprules"
	AviInfraSettingFullCRDName = "aviinfrasettings.ako.vmware.com"
	AviInfraSettingCRDSingular = "aviinfrasetting"
	AviInfraSettingCRDPlural   = "aviinfrasettings"
)

func createHostRuleCRD(clientset *apiextension.ApiextensionsV1Client, log logr.Logger) error {

	version := apiextensionv1.CustomResourceDefinitionVersion{
		Name:    Version,
		Served:  true,
		Storage: true,
		Schema: &apiextensionv1.CustomResourceValidation{
			OpenAPIV3Schema: &apiextensionv1.JSONSchemaProps{
				Type: "object",
				Properties: map[string]apiextensionv1.JSONSchemaProps{
					"spec": {
						Type:     "object",
						Required: []string{"virtualhost"},
						Properties: map[string]apiextensionv1.JSONSchemaProps{
							"virtualhost": {
								Type:     "object",
								Required: []string{"fqdn"},
								Properties: map[string]apiextensionv1.JSONSchemaProps{
									"analyticsProfile": {
										Type: "string",
									},
									"applicationProfile": {
										Type: "string",
									},
									"icapProfile": {
										Type: "array",
									},
									"enableVirtualHost": {
										Type: "boolean",
									},
									"errorPageProfile": {
										Type: "string",
									},
									"fqdn": {
										Type: "string",
									},
									"fqdnType": {
										Type: "string",
										Enum: []apiextensionv1.JSON{
											{
												Raw: []byte("\"Exact\""),
											},
											{
												Raw: []byte("\"Contains\""),
											},
											{
												Raw: []byte("\"Wildcard\""),
											},
										},
										Default: &apiextensionv1.JSON{
											Raw: []byte("\"Exact\""),
										},
									},
									"datascripts": {
										Items: &apiextensionv1.JSONSchemaPropsOrArray{
											Schema: &apiextensionv1.JSONSchemaProps{
												Type: "string",
											},
										},
										Type: "array",
									},
									"httpPolicy": {
										Type: "object",
										Properties: map[string]apiextensionv1.JSONSchemaProps{
											"overwrite": {
												Type: "boolean",
											},
											"policySets": {
												Items: &apiextensionv1.JSONSchemaPropsOrArray{
													Schema: &apiextensionv1.JSONSchemaProps{
														Type: "string",
													},
												},
												Type: "array",
											},
										},
									},
									"gslb": {
										Type: "object",
										Properties: map[string]apiextensionv1.JSONSchemaProps{
											"fqdn": {
												Type: "string",
											},
											"includeAliases": {
												Type: "boolean",
												Default: &apiextensionv1.JSON{
													Raw: []byte("false"),
												},
											},
										},
									},
									"wafPolicy": {
										Type: "string",
									},
									"tls": {
										Type:     "object",
										Required: []string{"sslKeyCertificate"},
										Properties: map[string]apiextensionv1.JSONSchemaProps{
											"sslProfile": {
												Type: "string",
											},
											"sslKeyCertificate": {
												Type:     "object",
												Required: []string{"name", "type"},
												Properties: map[string]apiextensionv1.JSONSchemaProps{
													"name": {
														Type: "string",
													},
													"type": {
														Type: "string",
														Enum: []apiextensionv1.JSON{
															{
																Raw: []byte("\"ref\""),
															},
															{
																Raw: []byte("\"secret\""),
															},
														},
													},
													"alternateCertificate": {
														Type:     "object",
														Required: []string{"name", "type"},
														Properties: map[string]apiextensionv1.JSONSchemaProps{
															"name": {
																Type: "string",
															},
															"type": {
																Type: "string",
																Enum: []apiextensionv1.JSON{
																	{
																		Raw: []byte("\"ref\""),
																	},
																	{
																		Raw: []byte("\"secret\""),
																	},
																},
															},
														},
													},
												},
											},
											"termination": {
												Type: "string",
												Enum: []apiextensionv1.JSON{
													{
														Raw: []byte("\"edge\""),
													},
												},
											},
										},
									},
									"analyticsPolicy": {
										Type: "object",
										Properties: map[string]apiextensionv1.JSONSchemaProps{
											"fullClientLogs": {
												Type: "object",
												Properties: map[string]apiextensionv1.JSONSchemaProps{
													"enabled": {
														Type: "boolean",
														Default: &apiextensionv1.JSON{
															Raw: []byte("false"),
														},
													},
													"throttle": {
														Type: "string",
														Enum: []apiextensionv1.JSON{
															{
																Raw: []byte("\"LOW\""),
															},
															{
																Raw: []byte("\"MEDIUM\""),
															},
															{
																Raw: []byte("\"HIGH\""),
															},
															{
																Raw: []byte("\"DISABLED\""),
															},
														},
														Default: &apiextensionv1.JSON{
															Raw: []byte("\"HIGH\""),
														},
													},
												},
											},
											"logAllHeaders": {
												Type: "boolean",
												Default: &apiextensionv1.JSON{
													Raw: []byte("false"),
												},
											},
										},
									},
									"tcpSettings": {
										Type: "object",
										Properties: map[string]apiextensionv1.JSONSchemaProps{
											"listeners": {
												Type: "array",
												Items: &apiextensionv1.JSONSchemaPropsOrArray{
													Schema: &apiextensionv1.JSONSchemaProps{
														Type: "object",
														Properties: map[string]apiextensionv1.JSONSchemaProps{
															"port": {
																Type:    "integer",
																Minimum: proto.Float64(1),
																Maximum: proto.Float64(65535),
															},
															"enableSSL": {
																Type: "boolean",
															},
														},
													},
												},
											},
											"loadBalancerIP": {
												Type: "string",
											},
										},
									},
									"aliases": {
										Type: "array",
										Items: &apiextensionv1.JSONSchemaPropsOrArray{
											Schema: &apiextensionv1.JSONSchemaProps{
												Type: "string",
											},
										},
									},
								},
							},
						},
					},
					"status": {
						Type: "object",
						Properties: map[string]apiextensionv1.JSONSchemaProps{
							"error": {
								Type: "string",
							},
							"status": {
								Type: "string",
							},
						},
					},
				},
			},
		},
		Subresources: &apiextensionv1.CustomResourceSubresources{Status: &apiextensionv1.CustomResourceSubresourceStatus{}},
		AdditionalPrinterColumns: []apiextensionv1.CustomResourceColumnDefinition{
			{
				Description: "virtualhost for which the hostrule is valid",
				JSONPath:    ".spec.virtualhost.fqdn",
				Name:        "Host",
				Type:        "string",
			},
			{
				Description: "status of the hostrule object",
				JSONPath:    ".status.status",
				Name:        "Status",
				Type:        "string",
			},
			{
				JSONPath: ".metadata.creationTimestamp",
				Name:     "Age",
				Type:     "date",
			},
		},
	}
	crd := &apiextensionv1.CustomResourceDefinition{
		TypeMeta:   v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{Name: HostRuleFullCRDName},
		Spec: apiextensionv1.CustomResourceDefinitionSpec{
			Group: CRDGroup,
			Names: apiextensionv1.CustomResourceDefinitionNames{
				Plural:   HostRuleCRDPlural,
				Singular: HostRuleCRDSingular,
				ShortNames: []string{
					HostRuleCRDSingular,
					"hr",
				},
				Kind: reflect.TypeOf(akov1alpha1.HostRule{}).Name(),
			},
			Scope: apiextensionv1.NamespaceScoped,
			Versions: []apiextensionv1.CustomResourceDefinitionVersion{
				version,
			},
			Conversion: &apiextensionv1.CustomResourceConversion{
				Strategy: apiextensionv1.ConversionStrategyType("None"),
			},
		},
		Status: apiextensionv1.CustomResourceDefinitionStatus{},
	}

	_, err := clientset.CustomResourceDefinitions().Create(context.TODO(), crd, v1.CreateOptions{})
	if err == nil {
		log.V(0).Info("hostrules.ako.vmware.com CRD created")
		return nil
	} else if apierrors.IsAlreadyExists(err) {
		log.V(0).Info("hostrules.ako.vmware.com CRD already exists")
		return nil
	}
	return err
}

func createHttpRuleCRD(clientset *apiextension.ApiextensionsV1Client, log logr.Logger) error {
	version := apiextensionv1.CustomResourceDefinitionVersion{
		Name:    Version,
		Served:  true,
		Storage: true,
		Schema: &apiextensionv1.CustomResourceValidation{
			OpenAPIV3Schema: &apiextensionv1.JSONSchemaProps{
				Type: "object",
				Properties: map[string]apiextensionv1.JSONSchemaProps{
					"spec": {
						Type:     "object",
						Required: []string{"fqdn"},
						Properties: map[string]apiextensionv1.JSONSchemaProps{
							"fqdn": {
								Type: "string",
							},
							"paths": {
								Type: "array",
								Items: &apiextensionv1.JSONSchemaPropsOrArray{
									Schema: &apiextensionv1.JSONSchemaProps{
										Type:     "object",
										Required: []string{"target"},
										Properties: map[string]apiextensionv1.JSONSchemaProps{
											"loadBalancerPolicy": {
												Type: "object",
												Properties: map[string]apiextensionv1.JSONSchemaProps{
													"algorithm": {
														Type: "string",
														Enum: []apiextensionv1.JSON{
															{
																Raw: []byte("\"LB_ALGORITHM_CONSISTENT_HASH\""),
															}, {
																Raw: []byte("\"LB_ALGORITHM_CORE_AFFINITY\""),
															}, {
																Raw: []byte("\"LB_ALGORITHM_FASTEST_RESPONSE\""),
															}, {
																Raw: []byte("\"LB_ALGORITHM_FEWEST_SERVERS\""),
															}, {
																Raw: []byte("\"LB_ALGORITHM_LEAST_CONNECTIONS\""),
															}, {
																Raw: []byte("\"LB_ALGORITHM_LEAST_LOAD\""),
															}, {
																Raw: []byte("\"LB_ALGORITHM_ROUND_ROBIN\""),
															},
														},
													},
													"hash": {
														Type: "string",
														Enum: []apiextensionv1.JSON{
															{
																Raw: []byte("\"LB_ALGORITHM_CONSISTENT_HASH_CALLID\""),
															}, {
																Raw: []byte("\"LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS\""),
															}, {
																Raw: []byte("\"LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS_AND_PORT\""),
															}, {
																Raw: []byte("\"LB_ALGORITHM_CONSISTENT_HASH_URI\""),
															}, {
																Raw: []byte("\"LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER\""),
															}, {
																Raw: []byte("\"LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_STRING\""),
															},
														},
													},
													"hostHeader": {
														Type: "string",
													},
												},
											},
											"target": {
												Type:    "string",
												Pattern: "^\\/.*$",
											},
											"healthMonitors": {
												Type: "array",
												Items: &apiextensionv1.JSONSchemaPropsOrArray{
													Schema: &apiextensionv1.JSONSchemaProps{
														Type: "string",
													},
												},
											},
											"applicationPersistence": {
												Type: "string",
											},
											"tls": {
												Type:     "object",
												Required: []string{"type"},
												Properties: map[string]apiextensionv1.JSONSchemaProps{
													"pkiProfile": {
														Type: "string",
													},
													"destinationCA": {
														Type: "string",
													},
													"sslProfile": {
														Type: "string",
													},
													"type": {
														Type: "string",
														Enum: []apiextensionv1.JSON{
															{
																Raw: []byte("\"reencrypt\""),
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					"status": {
						Type: "object",
						Properties: map[string]apiextensionv1.JSONSchemaProps{
							"error": {
								Type: "string",
							},
							"status": {
								Type: "string",
							},
						},
					},
				},
			},
		},
		Subresources: &apiextensionv1.CustomResourceSubresources{Status: &apiextensionv1.CustomResourceSubresourceStatus{}},
		AdditionalPrinterColumns: []apiextensionv1.CustomResourceColumnDefinition{
			{
				Description: "fqdn associated with the httprule",
				JSONPath:    ".spec.fqdn",
				Name:        "HOST",
				Type:        "string",
			},
			{
				Description: "status of the httprule object",
				JSONPath:    ".status.status",
				Name:        "Status",
				Type:        "string",
			},
			{
				JSONPath: ".metadata.creationTimestamp",
				Name:     "Age",
				Type:     "date",
			},
		},
	}
	crd := &apiextensionv1.CustomResourceDefinition{
		TypeMeta:   v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{Name: HttpRuleFullCRDName},
		Spec: apiextensionv1.CustomResourceDefinitionSpec{
			Group: CRDGroup,
			Names: apiextensionv1.CustomResourceDefinitionNames{
				Plural:   HttpRuleCRDPlural,
				Singular: HttpRuleCRDSingular,
				ShortNames: []string{
					HttpRuleCRDSingular,
				},
				Kind: reflect.TypeOf(akov1alpha1.HTTPRule{}).Name(),
			},
			Scope: apiextensionv1.NamespaceScoped,
			Versions: []apiextensionv1.CustomResourceDefinitionVersion{
				version,
			},
			Conversion: &apiextensionv1.CustomResourceConversion{
				Strategy: apiextensionv1.ConversionStrategyType("None"),
			},
		},
		Status: apiextensionv1.CustomResourceDefinitionStatus{},
	}

	_, err := clientset.CustomResourceDefinitions().Create(context.TODO(), crd, v1.CreateOptions{})
	if err == nil {
		log.V(0).Info("httprules.ako.vmware.com CRD created")
		return nil
	} else if apierrors.IsAlreadyExists(err) {
		log.V(0).Info("httprules.ako.vmware.com CRD already exists")
		return nil
	}
	return err
}

func createAviInfraSettingCRD(clientset *apiextension.ApiextensionsV1Client, log logr.Logger) error {
	version := apiextensionv1.CustomResourceDefinitionVersion{
		Name:    Version,
		Served:  true,
		Storage: true,
		Schema: &apiextensionv1.CustomResourceValidation{
			OpenAPIV3Schema: &apiextensionv1.JSONSchemaProps{
				Description: "AviInfraSetting is used to select specific Avi controller infra attributes.",
				Type:        "object",
				Properties: map[string]apiextensionv1.JSONSchemaProps{
					"spec": {
						Type: "object",
						Properties: map[string]apiextensionv1.JSONSchemaProps{
							"network": {
								Type: "object",
								Properties: map[string]apiextensionv1.JSONSchemaProps{
									"vipNetworks": {
										Type: "array",
										Items: &apiextensionv1.JSONSchemaPropsOrArray{
											Schema: &apiextensionv1.JSONSchemaProps{
												Type:     "object",
												Required: []string{"networkName"},
												Properties: map[string]apiextensionv1.JSONSchemaProps{
													"networkName": {
														Type: "string",
													},
													"cidr": {
														Type: "string",
													},
													"v6cidr": {
														Type: "string",
													},
												},
											},
										},
									},
									"nodeNetworks": {
										Type: "array",
										Items: &apiextensionv1.JSONSchemaPropsOrArray{
											Schema: &apiextensionv1.JSONSchemaProps{
												Type: "object",
												Properties: map[string]apiextensionv1.JSONSchemaProps{
													"networkName": {
														Type: "string",
													},
													"cidrs": {
														Type: "array",
														Items: &apiextensionv1.JSONSchemaPropsOrArray{
															Schema: &apiextensionv1.JSONSchemaProps{
																Type: "string",
															},
														},
													},
												},
												Required: []string{"networkName"},
											},
										},
									},
									"enableRhi": {
										Type: "boolean",
									},
									"enablePublicIP": {
										Type: "boolean",
									},
									"listeners": {
										Type: "array",
										Items: &apiextensionv1.JSONSchemaPropsOrArray{
											Schema: &apiextensionv1.JSONSchemaProps{
												Type: "object",
												Properties: map[string]apiextensionv1.JSONSchemaProps{
													"port": {
														Type:    "integer",
														Minimum: proto.Float64(1),
														Maximum: proto.Float64(65535),
													},
													"enableSSL": {
														Type: "boolean",
													},
													"enableHTTP2": {
														Type: "boolean",
													},
												},
											},
										},
									},
									"bgpPeerLabels": {
										Type: "array",
										Items: &apiextensionv1.JSONSchemaPropsOrArray{
											Schema: &apiextensionv1.JSONSchemaProps{
												Type: "string",
											},
										},
									},
								},
							},
							"seGroup": {
								Type:     "object",
								Required: []string{"name"},
								Properties: map[string]apiextensionv1.JSONSchemaProps{
									"name": {
										Type: "string",
									},
								},
							},
							"l7Settings": {
								Type:     "object",
								Required: []string{"shardSize"},
								Properties: map[string]apiextensionv1.JSONSchemaProps{
									"shardSize": {
										Type: "string",
										Enum: []apiextensionv1.JSON{
											{
												Raw: []byte("\"SMALL\""),
											},
											{
												Raw: []byte("\"MEDIUM\""),
											},
											{
												Raw: []byte("\"LARGE\""),
											},
											{
												Raw: []byte("\"DEDICATED\""),
											},
										},
									},
								},
							},
						},
					},
					"status": {
						Type: "object",
						Properties: map[string]apiextensionv1.JSONSchemaProps{
							"error": {
								Type: "string",
							},
							"status": {
								Type: "string",
							},
						},
					},
				},
			},
		},
		Subresources: &apiextensionv1.CustomResourceSubresources{Status: &apiextensionv1.CustomResourceSubresourceStatus{}},
		AdditionalPrinterColumns: []apiextensionv1.CustomResourceColumnDefinition{
			{
				Description: "status of the nas object",
				JSONPath:    ".status.status",
				Name:        "Status",
				Type:        "string",
			},
			{
				JSONPath: ".metadata.creationTimestamp",
				Name:     "Age",
				Type:     "date",
			},
		},
	}
	crd := &apiextensionv1.CustomResourceDefinition{
		TypeMeta:   v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{Name: AviInfraSettingFullCRDName},
		Spec: apiextensionv1.CustomResourceDefinitionSpec{
			Group: CRDGroup,
			Names: apiextensionv1.CustomResourceDefinitionNames{
				Plural:   AviInfraSettingCRDPlural,
				Singular: AviInfraSettingCRDSingular,
				ShortNames: []string{
					AviInfraSettingCRDSingular,
				},
				Kind: reflect.TypeOf(akov1alpha1.AviInfraSetting{}).Name(),
			},
			Scope: apiextensionv1.ClusterScoped,
			Versions: []apiextensionv1.CustomResourceDefinitionVersion{
				version,
			},
			Conversion: &apiextensionv1.CustomResourceConversion{
				Strategy: apiextensionv1.ConversionStrategyType("None"),
			},
		},
		Status: apiextensionv1.CustomResourceDefinitionStatus{},
	}

	_, err := clientset.CustomResourceDefinitions().Create(context.TODO(), crd, v1.CreateOptions{})
	if err == nil {
		log.V(0).Info("aviinfrasettings.ako.vmware.com CRD created")
		return nil
	} else if apierrors.IsAlreadyExists(err) {
		log.V(0).Info("aviinfrasettings.ako.vmware.com CRD already exists")
		return nil
	}
	return err
}

func createCRDs(cfg *rest.Config, log logr.Logger) error {
	kubeClient, _ := apiextension.NewForConfig(cfg)

	err := createHostRuleCRD(kubeClient, log)
	if err != nil {
		return err
	}
	err = createHttpRuleCRD(kubeClient, log)
	if err != nil {
		return err
	}
	err = createAviInfraSettingCRD(kubeClient, log)
	if err != nil {
		return err
	}
	return nil
}

func deleteCRDs(cfg *rest.Config) error {
	clientset, _ := apiextension.NewForConfig(cfg)
	err := clientset.CustomResourceDefinitions().Delete(context.TODO(), HostRuleFullCRDName, v1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = clientset.CustomResourceDefinitions().Delete(context.TODO(), HttpRuleFullCRDName, v1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = clientset.CustomResourceDefinitions().Delete(context.TODO(), AviInfraSettingFullCRDName, v1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
