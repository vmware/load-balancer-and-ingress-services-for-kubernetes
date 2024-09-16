package controllers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	logr "github.com/go-logr/logr"

	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	myscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const (
	hostruleCRDLocation        = "/var/crds/ako.vmware.com_hostrules.yaml"
	httpruleCRDLocation        = "/var/crds/ako.vmware.com_httprules.yaml"
	aviinfrasettingCRDLocation = "/var/crds/ako.vmware.com_aviinfrasettings.yaml"
	l4ruleCRDLocation          = "/var/crds/ako.vmware.com_l4rules.yaml"
	ssoruleCRDLocation         = "/var/crds/ako.vmware.com_ssorules.yaml"
	l7ruleCRDLocation          = "/var/crds/ako_vmware.com_l7rules.yaml"
	hostRuleFullCRDName        = "hostrules.ako.vmware.com"
	httpRuleFullCRDName        = "httprules.ako.vmware.com"
	aviInfraSettingFullCRDName = "aviinfrasettings.ako.vmware.com"
	l4RuleFullCRDName          = "l4rules.ako.vmware.com"
	ssoRuleFullCRDName         = "ssorules.ako.vmware.com"
	l7RuleFullCRDName          = "l7rules.ako.vmware.com"
)

var (
	hostruleCRD         *apiextensionv1.CustomResourceDefinition
	httpruleCRD         *apiextensionv1.CustomResourceDefinition
	aviinfrasettingCRD  *apiextensionv1.CustomResourceDefinition
	l4ruleCRD           *apiextensionv1.CustomResourceDefinition
	ssoruleCRD          *apiextensionv1.CustomResourceDefinition
	l7ruleCRD           *apiextensionv1.CustomResourceDefinition
	hostruleOnce        sync.Once
	httpruleOnce        sync.Once
	aviinfrasettingOnce sync.Once
	l4ruleOnce          sync.Once
	ssoruleOnce         sync.Once
	l7ruleOnce          sync.Once
)

func readCRDFromManifest(crdLocation string, log logr.Logger) (*apiextensionv1.CustomResourceDefinition, error) {
	crdYaml, err := os.ReadFile(crdLocation)
	if err != nil {
		log.Error(err, fmt.Sprintf("unable to read file : %s", crdLocation))
		return nil, err
	}
	crdObj := parseK8sYaml([]byte(crdYaml), log)
	if len(crdObj) == 0 {
		return nil, errors.New(fmt.Sprintf("Error while parsing yaml file at %s", crdLocation))
	}

	// convert the runtime.Object to unstructured.Unstructured
	crdUnstructuredObjMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(crdObj[0])
	if err != nil {
		log.Error(err, "unable to convert parsed crd obj to unstrectured map")
		return nil, err
	}

	var crd apiextensionv1.CustomResourceDefinition
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(crdUnstructuredObjMap, &crd); err != nil {
		return nil, err
	}
	return &crd, nil
}

func createHostRuleCRD(clientset *apiextension.ApiextensionsV1Client, log logr.Logger) error {
	var err error
	hostruleOnce.Do(func() {
		hostruleCRD, err = readCRDFromManifest(hostruleCRDLocation, log)
	})
	if err != nil {
		return err
	} else if hostruleCRD == nil {
		return errors.New(fmt.Sprintf("Failure while reading %s CRD manifest", hostRuleFullCRDName))
	}

	existingHostRule, err := clientset.CustomResourceDefinitions().Get(context.TODO(), hostRuleFullCRDName, v1.GetOptions{})
	if err == nil && existingHostRule != nil {
		if hostruleCRD.GetResourceVersion() == existingHostRule.GetResourceVersion() {
			log.Info(fmt.Sprintf("no updates required for %s CRD", hostRuleFullCRDName))
		} else {
			hostruleCRD.SetResourceVersion(existingHostRule.GetResourceVersion())
			_, err = clientset.CustomResourceDefinitions().Update(context.TODO(), hostruleCRD, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s CRD", hostRuleFullCRDName))
				return err
			} else {
				log.Info(fmt.Sprintf("successfully updated %s CRD", hostRuleFullCRDName))
			}
		}
		return nil
	}

	_, err = clientset.CustomResourceDefinitions().Create(context.TODO(), hostruleCRD, v1.CreateOptions{})
	if err == nil {
		log.Info(fmt.Sprintf("%s CRD created", hostRuleFullCRDName))
		return nil
	} else if apierrors.IsAlreadyExists(err) {
		log.Info(fmt.Sprintf("%s CRD already exists", hostRuleFullCRDName))
		return nil
	}
	return err
}

func createHttpRuleCRD(clientset *apiextension.ApiextensionsV1Client, log logr.Logger) error {
	var err error
	httpruleOnce.Do(func() {
		httpruleCRD, err = readCRDFromManifest(httpruleCRDLocation, log)
	})
	if err != nil {
		return err
	} else if httpruleCRD == nil {
		return errors.New(fmt.Sprintf("Failure while reading %s CRD manifest", httpRuleFullCRDName))
	}

	existingHttpRule, err := clientset.CustomResourceDefinitions().Get(context.TODO(), httpRuleFullCRDName, v1.GetOptions{})
	if err == nil && existingHttpRule != nil {
		if httpruleCRD.GetResourceVersion() == existingHttpRule.GetResourceVersion() {
			log.Info(fmt.Sprintf("no updates required for %s CRD", httpRuleFullCRDName))
		} else {
			httpruleCRD.SetResourceVersion(existingHttpRule.GetResourceVersion())
			_, err = clientset.CustomResourceDefinitions().Update(context.TODO(), httpruleCRD, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s CRD", httpRuleFullCRDName))
				return err
			} else {
				log.Info(fmt.Sprintf("successfully updated %s CRD", httpRuleFullCRDName))
			}
		}
		return nil
	}

	_, err = clientset.CustomResourceDefinitions().Create(context.TODO(), httpruleCRD, v1.CreateOptions{})
	if err == nil {
		log.Info(fmt.Sprintf("%s CRD created", httpRuleFullCRDName))
		return nil
	} else if apierrors.IsAlreadyExists(err) {
		log.Info(fmt.Sprintf("%s CRD already exists", httpRuleFullCRDName))
		return nil
	}
	return err
}

func createAviInfraSettingCRD(clientset *apiextension.ApiextensionsV1Client, log logr.Logger) error {
	var err error
	aviinfrasettingOnce.Do(func() {
		aviinfrasettingCRD, err = readCRDFromManifest(aviinfrasettingCRDLocation, log)
	})
	if err != nil {
		return err
	} else if aviinfrasettingCRD == nil {
		return errors.New(fmt.Sprintf("Failure while reading %s CRD manifest", aviInfraSettingFullCRDName))
	}

	existingAviinfrasetting, err := clientset.CustomResourceDefinitions().Get(context.TODO(), aviInfraSettingFullCRDName, v1.GetOptions{})
	if err == nil && existingAviinfrasetting != nil {
		if aviinfrasettingCRD.GetResourceVersion() == existingAviinfrasetting.GetResourceVersion() {
			log.Info(fmt.Sprintf("no updates required for %s CRD", aviInfraSettingFullCRDName))
		} else {
			aviinfrasettingCRD.SetResourceVersion(existingAviinfrasetting.GetResourceVersion())
			_, err = clientset.CustomResourceDefinitions().Update(context.TODO(), aviinfrasettingCRD, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s CRD", aviInfraSettingFullCRDName))
				return err
			} else {
				log.Info(fmt.Sprintf("successfully updated %s CRD", aviInfraSettingFullCRDName))
			}
		}
		return nil
	}

	_, err = clientset.CustomResourceDefinitions().Create(context.TODO(), aviinfrasettingCRD, v1.CreateOptions{})
	if err == nil {
		log.Info(fmt.Sprintf("%s CRD created", aviInfraSettingFullCRDName))
		return nil
	} else if apierrors.IsAlreadyExists(err) {
		log.Info(fmt.Sprintf("%s CRD already exists", aviInfraSettingFullCRDName))
		return nil
	}
	return err
}

func createL4RuleCRD(clientset *apiextension.ApiextensionsV1Client, log logr.Logger) error {
	var err error
	l4ruleOnce.Do(func() {
		l4ruleCRD, err = readCRDFromManifest(l4ruleCRDLocation, log)
	})
	if err != nil {
		return err
	} else if l4ruleCRD == nil {
		return errors.New(fmt.Sprintf("Failure while reading %s CRD manifest", l4RuleFullCRDName))
	}

	existingl4rule, err := clientset.CustomResourceDefinitions().Get(context.TODO(), l4RuleFullCRDName, v1.GetOptions{})
	if err == nil && existingl4rule != nil {
		if l4ruleCRD.GetResourceVersion() == existingl4rule.GetResourceVersion() {
			log.Info(fmt.Sprintf("no updates required for %s CRD", l4RuleFullCRDName))
		} else {
			l4ruleCRD.SetResourceVersion(existingl4rule.GetResourceVersion())
			_, err = clientset.CustomResourceDefinitions().Update(context.TODO(), l4ruleCRD, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s CRD", l4RuleFullCRDName))
				return err
			} else {
				log.Info(fmt.Sprintf("successfully updated %s CRD", l4RuleFullCRDName))
			}
		}
		return nil
	}

	_, err = clientset.CustomResourceDefinitions().Create(context.TODO(), l4ruleCRD, v1.CreateOptions{})
	if err == nil {
		log.Info(fmt.Sprintf("%s CRD created", l4RuleFullCRDName))
		return nil
	} else if apierrors.IsAlreadyExists(err) {
		log.Info(fmt.Sprintf("%s CRD already exists", l4RuleFullCRDName))
		return nil
	}
	return err
}

func createSSORuleCRD(clientset *apiextension.ApiextensionsV1Client, log logr.Logger) error {
	var err error
	ssoruleOnce.Do(func() {
		ssoruleCRD, err = readCRDFromManifest(ssoruleCRDLocation, log)
	})
	if err != nil {
		return err
	} else if ssoruleCRD == nil {
		return errors.New(fmt.Sprintf("Failure while reading %s CRD manifest", ssoRuleFullCRDName))
	}

	existingSSORule, err := clientset.CustomResourceDefinitions().Get(context.TODO(), ssoRuleFullCRDName, v1.GetOptions{})
	if err == nil && existingSSORule != nil {
		if ssoruleCRD.GetResourceVersion() == existingSSORule.GetResourceVersion() {
			log.Info(fmt.Sprintf("no updates required for %s CRD", ssoRuleFullCRDName))
		} else {
			ssoruleCRD.SetResourceVersion(existingSSORule.GetResourceVersion())
			_, err = clientset.CustomResourceDefinitions().Update(context.TODO(), ssoruleCRD, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s CRD", ssoRuleFullCRDName))
				return err
			} else {
				log.Info(fmt.Sprintf("successfully updated %s CRD", ssoRuleFullCRDName))
			}
		}
		return nil
	}

	_, err = clientset.CustomResourceDefinitions().Create(context.TODO(), ssoruleCRD, v1.CreateOptions{})
	if err == nil {
		log.Info(fmt.Sprintf("%s CRD created", ssoRuleFullCRDName))
		return nil
	} else if apierrors.IsAlreadyExists(err) {
		log.Info(fmt.Sprintf("%s CRD already exists", ssoRuleFullCRDName))
		return nil
	}
	return err
}

func createL7RuleCRD(clientset *apiextension.ApiextensionsV1Client, log logr.Logger) error {
	var err error
	l7ruleOnce.Do(func() {
		l7ruleCRD, err = readCRDFromManifest(l7ruleCRDLocation, log)
	})
	if err != nil {
		return err
	} else if l7ruleCRD == nil {
		return errors.New(fmt.Sprintf("Failure while reading %s CRD manifest", l7RuleFullCRDName))
	}

	existingl7rule, err := clientset.CustomResourceDefinitions().Get(context.TODO(), l7RuleFullCRDName, v1.GetOptions{})
	if err == nil && existingl7rule != nil {
		if l7ruleCRD.GetResourceVersion() == existingl7rule.GetResourceVersion() {
			log.Info(fmt.Sprintf("no updates required for %s CRD", l7RuleFullCRDName))
		} else {
			l7ruleCRD.SetResourceVersion(existingl7rule.GetResourceVersion())
			_, err = clientset.CustomResourceDefinitions().Update(context.TODO(), l7ruleCRD, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s CRD", l7RuleFullCRDName))
				return err
			} else {
				log.Info(fmt.Sprintf("successfully updated %s CRD", l7RuleFullCRDName))
			}
		}
		return nil
	}

	_, err = clientset.CustomResourceDefinitions().Create(context.TODO(), l7ruleCRD, v1.CreateOptions{})
	if err == nil {
		log.Info(fmt.Sprintf("%s CRD created", l7RuleFullCRDName))
		return nil
	} else if apierrors.IsAlreadyExists(err) {
		log.Info(fmt.Sprintf("%s CRD already exists", l7RuleFullCRDName))
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
	err = createL4RuleCRD(kubeClient, log)
	if err != nil {
		return err
	}
	err = createSSORuleCRD(kubeClient, log)
	if err != nil {
		return err
	}
	err = createL7RuleCRD(kubeClient, log)
	if err != nil {
		return err
	}
	return nil
}

func deleteCRDs(cfg *rest.Config, log logr.Logger) error {
	clientset, _ := apiextension.NewForConfig(cfg)
	var err error
	func(crdFullNames ...string) {
		for _, crdFullName := range crdFullNames {
			err = clientset.CustomResourceDefinitions().Delete(context.TODO(), crdFullName, v1.DeleteOptions{})
			if err != nil {
				return
			}
			log.Info(fmt.Sprintf("%s crd deleted successfully", crdFullName))
		}
	}(hostRuleFullCRDName, httpRuleFullCRDName, aviInfraSettingFullCRDName, l4RuleFullCRDName, ssoRuleFullCRDName, l7RuleFullCRDName)
	if err != nil {
		return err
	}
	return nil
}

func parseK8sYaml(fileR []byte, log logr.Logger) []runtime.Object {
	sch := runtime.NewScheme()
	_ = myscheme.AddToScheme(sch)
	_ = apiextensionv1.AddToScheme(sch)
	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode
	fileAsString := string(fileR[:])
	sepYamlfiles := strings.Split(fileAsString, "---")
	retVal := make([]runtime.Object, 0, len(sepYamlfiles))
	for _, f := range sepYamlfiles {
		if f == "\n" || f == "" {
			// ignore empty cases
			continue
		}
		obj, _, err := decode([]byte(f), nil, nil)

		if err != nil {
			log.Error(err, fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
			continue
		}
		retVal = append(retVal, obj)
	}
	return retVal
}
