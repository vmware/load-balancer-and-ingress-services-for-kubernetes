package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
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

	/******* While running unit tests, please comment above paths and uncomment below paths.
	We will add a more permanent solution later.
	*******/
	/*hostruleCRDLocation        = "../../../helm/ako/crds/ako.vmware.com_hostrules.yaml"
	httpruleCRDLocation        = "../../../helm/ako/crds/ako.vmware.com_httprules.yaml"
	aviinfrasettingCRDLocation = "../../../helm/ako/crds/ako.vmware.com_aviinfrasettings.yaml"
	l4ruleCRDLocation          = "../../../helm/ako/crds/ako.vmware.com_l4rules.yaml"
	ssoruleCRDLocation         = "../../../helm/ako/crds/ako.vmware.com_ssorules.yaml"
	l7ruleCRDLocation          = "../../../helm/ako/crds/ako_vmware.com_l7rules.yaml"*/

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

	// Global error variables to capture initialization errors from sync.Once blocks
	hostruleCRDInitError        error
	httpruleCRDInitError        error
	aviinfrasettingCRDInitError error
	l4ruleCRDInitError          error
	ssoruleCRDInitError         error
	l7ruleCRDInitError          error
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

func createHostRuleCRD(clientset apiextension.ApiextensionsV1Interface, log logr.Logger) error {
	hostruleOnce.Do(func() {
		hostruleCRD, hostruleCRDInitError = readCRDFromManifest(hostruleCRDLocation, log)
	})

	if hostruleCRDInitError != nil {
		return hostruleCRDInitError
	}
	if hostruleCRD == nil {
		// This case should ideally be covered by hostruleCRDInitError, but as a safeguard.
		return errors.New(fmt.Sprintf("Global %s CRD is nil after initialization attempt", hostRuleFullCRDName))
	}

	desiredCRD := hostruleCRD // Use the globally initialized CRD

	existingCRD, getErr := clientset.CustomResourceDefinitions().Get(context.TODO(), hostRuleFullCRDName, v1.GetOptions{})

	if getErr == nil && existingCRD != nil {
		// CRD exists, check if update is required by comparing specs
		if reflect.DeepEqual(existingCRD.Spec, desiredCRD.Spec) {
			log.Info(fmt.Sprintf("no updates required for %s CRD", hostRuleFullCRDName))
			// Update global variable with fetched object if no update was needed,
			// to ensure it holds the latest resourceVersion.
			hostruleCRD = existingCRD
		} else {
			// Update the existing CRD's spec with the desired spec
			existingCRD.Spec = desiredCRD.Spec
			_, err := clientset.CustomResourceDefinitions().Update(context.TODO(), existingCRD, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s CRD", hostRuleFullCRDName))
				return err
			}
			log.Info(fmt.Sprintf("successfully updated %s CRD", hostRuleFullCRDName))
			// After successful update, fetch the latest state to update the global variable
			newHostRule, err := clientset.CustomResourceDefinitions().Get(context.TODO(), hostRuleFullCRDName, v1.GetOptions{})
			if err == nil && newHostRule != nil {
				hostruleCRD = newHostRule // Update global variable with the latest fetched object
			}
		}
		return nil
	} else if apierrors.IsNotFound(getErr) {
		// CRD does not exist, create it
		desiredCRD.SetResourceVersion("") // Ensure no resource version is set for creation
		_, err := clientset.CustomResourceDefinitions().Create(context.TODO(), desiredCRD, v1.CreateOptions{})
		if err == nil {
			log.Info(fmt.Sprintf("%s CRD created", hostRuleFullCRDName))
			hostruleCRD = desiredCRD // Update global variable with the newly created object
		} else if apierrors.IsAlreadyExists(err) {
			// This can happen in a race condition, if another controller or process creates it concurrently.
			// In this case, we can just log and proceed, as the next reconciliation will pick it up.
			log.Info(fmt.Sprintf("%s CRD already exists (race condition)", hostRuleFullCRDName))
			// Attempt to get it again to ensure the global variable is set
			existingCRDAfterRace, getAfterRaceErr := clientset.CustomResourceDefinitions().Get(context.TODO(), hostRuleFullCRDName, v1.GetOptions{})
			if getAfterRaceErr == nil && existingCRDAfterRace != nil {
				hostruleCRD = existingCRDAfterRace
			}
			return nil
		}
		return err // Return other errors from create operation
	}
	return getErr // Return other errors from get operation
}

func createHttpRuleCRD(clientset apiextension.ApiextensionsV1Interface, log logr.Logger) error {
	httpruleOnce.Do(func() {
		httpruleCRD, httpruleCRDInitError = readCRDFromManifest(httpruleCRDLocation, log)
	})

	if httpruleCRDInitError != nil {
		return httpruleCRDInitError
	}
	if httpruleCRD == nil {
		return errors.New(fmt.Sprintf("Global %s CRD is nil after initialization attempt", httpRuleFullCRDName))
	}

	desiredCRD := httpruleCRD

	existingCRD, getErr := clientset.CustomResourceDefinitions().Get(context.TODO(), httpRuleFullCRDName, v1.GetOptions{})

	if getErr == nil && existingCRD != nil {
		if reflect.DeepEqual(existingCRD.Spec, desiredCRD.Spec) {
			log.Info(fmt.Sprintf("no updates required for %s CRD", httpRuleFullCRDName))
			httpruleCRD = existingCRD
		} else {
			existingCRD.Spec = desiredCRD.Spec
			_, err := clientset.CustomResourceDefinitions().Update(context.TODO(), existingCRD, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s CRD", httpRuleFullCRDName))
				return err
			}
			log.Info(fmt.Sprintf("successfully updated %s CRD", httpRuleFullCRDName))
			newHttpRule, err := clientset.CustomResourceDefinitions().Get(context.TODO(), httpRuleFullCRDName, v1.GetOptions{})
			if err == nil && newHttpRule != nil {
				httpruleCRD = newHttpRule
			}
		}
		return nil
	} else if apierrors.IsNotFound(getErr) {
		desiredCRD.SetResourceVersion("")
		_, err := clientset.CustomResourceDefinitions().Create(context.TODO(), desiredCRD, v1.CreateOptions{})
		if err == nil {
			log.Info(fmt.Sprintf("%s CRD created", httpRuleFullCRDName))
			httpruleCRD = desiredCRD
		} else if apierrors.IsAlreadyExists(err) {
			log.Info(fmt.Sprintf("%s CRD already exists (race condition)", httpRuleFullCRDName))
			existingCRDAfterRace, getAfterRaceErr := clientset.CustomResourceDefinitions().Get(context.TODO(), httpRuleFullCRDName, v1.GetOptions{})
			if getAfterRaceErr == nil && existingCRDAfterRace != nil {
				httpruleCRD = existingCRDAfterRace
			}
			return nil
		}
		return err
	}
	return getErr
}

func createAviInfraSettingCRD(clientset apiextension.ApiextensionsV1Interface, log logr.Logger) error {
	aviinfrasettingOnce.Do(func() {
		aviinfrasettingCRD, aviinfrasettingCRDInitError = readCRDFromManifest(aviinfrasettingCRDLocation, log)
	})

	if aviinfrasettingCRDInitError != nil {
		return aviinfrasettingCRDInitError
	}
	if aviinfrasettingCRD == nil {
		return errors.New(fmt.Sprintf("Global %s CRD is nil after initialization attempt", aviInfraSettingFullCRDName))
	}

	desiredCRD := aviinfrasettingCRD

	existingCRD, getErr := clientset.CustomResourceDefinitions().Get(context.TODO(), aviInfraSettingFullCRDName, v1.GetOptions{})

	if getErr == nil && existingCRD != nil {
		if reflect.DeepEqual(existingCRD.Spec, desiredCRD.Spec) {
			log.Info(fmt.Sprintf("no updates required for %s CRD", aviInfraSettingFullCRDName))
			aviinfrasettingCRD = existingCRD
		} else {
			existingCRD.Spec = desiredCRD.Spec
			_, err := clientset.CustomResourceDefinitions().Update(context.TODO(), existingCRD, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s CRD", aviInfraSettingFullCRDName))
				return err
			}
			log.Info(fmt.Sprintf("successfully updated %s CRD", aviInfraSettingFullCRDName))
			newAviinfrasetting, err := clientset.CustomResourceDefinitions().Get(context.TODO(), aviInfraSettingFullCRDName, v1.GetOptions{})
			if err == nil && newAviinfrasetting != nil {
				aviinfrasettingCRD = newAviinfrasetting
			}
		}
		return nil
	} else if apierrors.IsNotFound(getErr) {
		desiredCRD.SetResourceVersion("")
		_, err := clientset.CustomResourceDefinitions().Create(context.TODO(), desiredCRD, v1.CreateOptions{})
		if err == nil {
			log.Info(fmt.Sprintf("%s CRD created", aviInfraSettingFullCRDName))
			aviinfrasettingCRD = desiredCRD
		} else if apierrors.IsAlreadyExists(err) {
			log.Info(fmt.Sprintf("%s CRD already exists (race condition)", aviInfraSettingFullCRDName))
			existingCRDAfterRace, getAfterRaceErr := clientset.CustomResourceDefinitions().Get(context.TODO(), aviInfraSettingFullCRDName, v1.GetOptions{})
			if getAfterRaceErr == nil && existingCRDAfterRace != nil {
				aviinfrasettingCRD = existingCRDAfterRace
			}
			return nil
		}
		return err
	}
	return getErr
}

func createL4RuleCRD(clientset apiextension.ApiextensionsV1Interface, log logr.Logger) error {
	l4ruleOnce.Do(func() {
		l4ruleCRD, l4ruleCRDInitError = readCRDFromManifest(l4ruleCRDLocation, log)
	})

	if l4ruleCRDInitError != nil {
		return l4ruleCRDInitError
	}
	if l4ruleCRD == nil {
		return errors.New(fmt.Sprintf("Global %s CRD is nil after initialization attempt", l4RuleFullCRDName))
	}

	desiredCRD := l4ruleCRD

	existingCRD, getErr := clientset.CustomResourceDefinitions().Get(context.TODO(), l4RuleFullCRDName, v1.GetOptions{})

	if getErr == nil && existingCRD != nil {
		if reflect.DeepEqual(existingCRD.Spec, desiredCRD.Spec) {
			log.Info(fmt.Sprintf("no updates required for %s CRD", l4RuleFullCRDName))
			l4ruleCRD = existingCRD
		} else {
			existingCRD.Spec = desiredCRD.Spec
			_, err := clientset.CustomResourceDefinitions().Update(context.TODO(), existingCRD, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s CRD", l4RuleFullCRDName))
				return err
			}
			log.Info(fmt.Sprintf("successfully updated %s CRD", l4RuleFullCRDName))
			newl4rule, err := clientset.CustomResourceDefinitions().Get(context.TODO(), l4RuleFullCRDName, v1.GetOptions{})
			if err == nil && newl4rule != nil {
				l4ruleCRD = newl4rule
			}
		}
		return nil
	} else if apierrors.IsNotFound(getErr) {
		desiredCRD.SetResourceVersion("")
		_, err := clientset.CustomResourceDefinitions().Create(context.TODO(), desiredCRD, v1.CreateOptions{})
		if err == nil {
			log.Info(fmt.Sprintf("%s CRD created", l4RuleFullCRDName))
			l4ruleCRD = desiredCRD
		} else if apierrors.IsAlreadyExists(err) {
			log.Info(fmt.Sprintf("%s CRD already exists (race condition)", l4RuleFullCRDName))
			existingCRDAfterRace, getAfterRaceErr := clientset.CustomResourceDefinitions().Get(context.TODO(), l4RuleFullCRDName, v1.GetOptions{})
			if getAfterRaceErr == nil && existingCRDAfterRace != nil {
				l4ruleCRD = existingCRDAfterRace
			}
			return nil
		}
		return err
	}
	return getErr
}

func createSSORuleCRD(clientset apiextension.ApiextensionsV1Interface, log logr.Logger) error {
	ssoruleOnce.Do(func() {
		ssoruleCRD, ssoruleCRDInitError = readCRDFromManifest(ssoruleCRDLocation, log)
	})

	if ssoruleCRDInitError != nil {
		return ssoruleCRDInitError
	}
	if ssoruleCRD == nil {
		return errors.New(fmt.Sprintf("Global %s CRD is nil after initialization attempt", ssoRuleFullCRDName))
	}

	desiredCRD := ssoruleCRD

	existingCRD, getErr := clientset.CustomResourceDefinitions().Get(context.TODO(), ssoRuleFullCRDName, v1.GetOptions{})

	if getErr == nil && existingCRD != nil {
		if reflect.DeepEqual(existingCRD.Spec, desiredCRD.Spec) {
			log.Info(fmt.Sprintf("no updates required for %s CRD", ssoRuleFullCRDName))
			ssoruleCRD = existingCRD
		} else {
			existingCRD.Spec = desiredCRD.Spec
			_, err := clientset.CustomResourceDefinitions().Update(context.TODO(), existingCRD, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s CRD", ssoRuleFullCRDName))
				return err
			}
			log.Info(fmt.Sprintf("successfully updated %s CRD", ssoRuleFullCRDName))
			newSSORule, err := clientset.CustomResourceDefinitions().Get(context.TODO(), ssoRuleFullCRDName, v1.GetOptions{})
			if err == nil && newSSORule != nil {
				ssoruleCRD = newSSORule
			}
		}
		return nil
	} else if apierrors.IsNotFound(getErr) {
		desiredCRD.SetResourceVersion("")
		_, err := clientset.CustomResourceDefinitions().Create(context.TODO(), desiredCRD, v1.CreateOptions{})
		if err == nil {
			log.Info(fmt.Sprintf("%s CRD created", ssoRuleFullCRDName))
			ssoruleCRD = desiredCRD
		} else if apierrors.IsAlreadyExists(err) {
			log.Info(fmt.Sprintf("%s CRD already exists (race condition)", ssoRuleFullCRDName))
			existingCRDAfterRace, getAfterRaceErr := clientset.CustomResourceDefinitions().Get(context.TODO(), ssoRuleFullCRDName, v1.GetOptions{})
			if getAfterRaceErr == nil && existingCRDAfterRace != nil {
				ssoruleCRD = existingCRDAfterRace
			}
			return nil
		}
		return err
	}
	return getErr
}

func createL7RuleCRD(clientset apiextension.ApiextensionsV1Interface, log logr.Logger) error {
	l7ruleOnce.Do(func() {
		l7ruleCRD, l7ruleCRDInitError = readCRDFromManifest(l7ruleCRDLocation, log)
	})

	if l7ruleCRDInitError != nil {
		return l7ruleCRDInitError
	}
	if l7ruleCRD == nil {
		return errors.New(fmt.Sprintf("Global %s CRD is nil after initialization attempt", l7RuleFullCRDName))
	}

	desiredCRD := l7ruleCRD

	existingCRD, getErr := clientset.CustomResourceDefinitions().Get(context.TODO(), l7RuleFullCRDName, v1.GetOptions{})

	if getErr == nil && existingCRD != nil {
		if reflect.DeepEqual(existingCRD.Spec, desiredCRD.Spec) {
			log.Info(fmt.Sprintf("no updates required for %s CRD", l7RuleFullCRDName))
			l7ruleCRD = existingCRD
		} else {
			existingCRD.Spec = desiredCRD.Spec
			_, err := clientset.CustomResourceDefinitions().Update(context.TODO(), existingCRD, v1.UpdateOptions{})
			if err != nil {
				log.Error(err, fmt.Sprintf("Error while updating %s CRD", l7RuleFullCRDName))
				return err
			}
			log.Info(fmt.Sprintf("successfully updated %s CRD", l7RuleFullCRDName))
			newl7rule, err := clientset.CustomResourceDefinitions().Get(context.TODO(), l7RuleFullCRDName, v1.GetOptions{})
			if err == nil && newl7rule != nil {
				l7ruleCRD = newl7rule
			}
		}
		return nil
	} else if apierrors.IsNotFound(getErr) {
		desiredCRD.SetResourceVersion("")
		_, err := clientset.CustomResourceDefinitions().Create(context.TODO(), desiredCRD, v1.CreateOptions{})
		if err == nil {
			log.Info(fmt.Sprintf("%s CRD created", l7RuleFullCRDName))
			l7ruleCRD = desiredCRD
		} else if apierrors.IsAlreadyExists(err) {
			log.Info(fmt.Sprintf("%s CRD already exists (race condition)", l7RuleFullCRDName))
			existingCRDAfterRace, getAfterRaceErr := clientset.CustomResourceDefinitions().Get(context.TODO(), l7RuleFullCRDName, v1.GetOptions{})
			if getAfterRaceErr == nil && existingCRDAfterRace != nil {
				l7ruleCRD = existingCRDAfterRace
			}
			return nil
		}
		return err
	}
	return getErr
}

func createCRDs(cfg *rest.Config, log logr.Logger) error {
	kubeClient, _ := apiextension.NewForConfig(cfg)

	// Use a slice to store errors and return all of them if any occur
	var allErrors []error

	if err := createHostRuleCRD(kubeClient, log); err != nil {
		allErrors = append(allErrors, err)
	}
	if err := createHttpRuleCRD(kubeClient, log); err != nil {
		allErrors = append(allErrors, err)
	}
	if err := createAviInfraSettingCRD(kubeClient, log); err != nil {
		allErrors = append(allErrors, err)
	}
	if err := createL4RuleCRD(kubeClient, log); err != nil {
		allErrors = append(allErrors, err)
	}
	if err := createSSORuleCRD(kubeClient, log); err != nil {
		allErrors = append(allErrors, err)
	}
	if err := createL7RuleCRD(kubeClient, log); err != nil {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("failed to create/update one or more CRDs: %v", allErrors)
	}
	return nil
}

func deleteCRDs(cfg *rest.Config, log logr.Logger) error {
	clientset, _ := apiextension.NewForConfig(cfg)
	var allErrors []error

	crdFullNames := []string{
		hostRuleFullCRDName,
		httpRuleFullCRDName,
		aviInfraSettingFullCRDName,
		l4RuleFullCRDName,
		ssoRuleFullCRDName,
		l7RuleFullCRDName,
	}

	for _, crdFullName := range crdFullNames {
		err := clientset.CustomResourceDefinitions().Delete(context.TODO(), crdFullName, v1.DeleteOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Info(fmt.Sprintf("%s CRD not found, skipping deletion", crdFullName))
			} else {
				log.Error(err, fmt.Sprintf("Error while deleting %s CRD", crdFullName))
				allErrors = append(allErrors, err)
			}
		} else {
			log.Info(fmt.Sprintf("%s CRD deleted successfully", crdFullName))
		}
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("failed to delete one or more CRDs: %v", allErrors)
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
