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

// nolint:unused
package scaletest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/avinetworks/sdk/go/clients"
	"github.com/onsi/gomega"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/scaletest/lib"
)

const (
	SECURE       = "secure"
	INSECURE     = "insecure"
	MULTIHOST    = "multi-host"
	CONTROLLER   = "Controller"
	KUBENODE     = "Node"
	AKONAMESPACE = "avi-system"
)

var (
	testbedFileName          string
	namespace                string
	appName                  string
	serviceNamePrefix        string
	ingressNamePrefix        string
	akoPodName               string
	AviClients               []*clients.AviClient
	numGoRoutines            int
	listOfServicesCreated    []string
	ingressesCreated         []string
	ingressesDeleted         []string
	ingressesUpdated         []string
	ingressHostNames         []string
	ingressSecureHostNames   []string
	ingressInsecureHostNames []string
	initialNumOfPools        = 0
	initialNumOfVSes         = 0
	initialNumOfFQDN         = 0
	initialVSesList          []string
	initialFQDNList          []string
	ingressType              string
	numOfIng                 int
	numOfLBSvc               int
	numOfPodsForLBSvc        = 100
	clusterName              string
	timeout                  string
	dnsVSUUID                string
	testCaseTimeOut          = 1800
	testPollInterval         = "15s"
	mutex                    sync.Mutex
	REBOOTAKO                = false
	REBOOTCONTROLLER         = false
	REBOOTNODE               = false
)

/* Basic Setup including parsing of parameters from command line
And assignment of global variables fetched from the testbed
Aviclients, Deployment and Services used by the test are created */
func Setup() {
	var testbedParams lib.TestbedFields
	var err error
	timeout = os.Args[4]
	testbedFileName = os.Args[5]
	numGoRoutines, err = strconv.Atoi(os.Args[6])
	if err != nil {
		fmt.Println("Setting default value for number of Go routines to 20")
		numGoRoutines = 20
	}
	if numGoRoutines <= 0 {
		fmt.Println("ERROR : Number of Go Routines cannot be zero or negative.")
		os.Exit(0)
	}
	numOfLBSvc, err = strconv.Atoi(os.Args[7])
	if err != nil {
		fmt.Println("ERROR : Number of LB services not provided")
		os.Exit(0)
	}
	numOfIng, err = strconv.Atoi(os.Args[8])
	if err != nil {
		fmt.Println("ERROR : Number of ingresses not provided")
		os.Exit(0)
	}
	testbed, er := os.Open(testbedFileName)
	if er != nil {
		fmt.Println("ERROR : Error opening testbed file ", testbedFileName, " with error : ", er)
		os.Exit(0)
	}
	defer testbed.Close()
	byteValue, err := ioutil.ReadAll(testbed)
	if err != nil {
		fmt.Println("ERROR : Failed to read the testbed file with error : ", err)
		os.Exit(0)
	}
	err = json.Unmarshal(byteValue, &testbedParams)
	if err != nil {
		fmt.Println("ERROR : Failed to unmarshal testbed file as : ", err)
		os.Exit(0)
	}
	namespace = testbedParams.TestParams.Namespace
	appName = testbedParams.TestParams.AppName
	serviceNamePrefix = testbedParams.TestParams.ServiceNamePrefix
	ingressNamePrefix = testbedParams.TestParams.IngressNamePrefix
	clusterName = testbedParams.AkoParam.Clusters[0].ClusterName
	dnsVSUUID = testbedParams.TestParams.DnsVSUUID
	akoPodName = testbedParams.TestParams.AkoPodName
	os.Setenv("CTRL_USERNAME", testbedParams.Vm[0].UserName)
	os.Setenv("CTRL_PASSWORD", testbedParams.Vm[0].Password)
	os.Setenv("CTRL_IPADDRESS", testbedParams.Vm[0].IP)
	lib.KubeInit(testbedParams.AkoParam.Clusters[0].KubeConfigFilePath)
	AviClients, err = lib.SharedAVIClients(2)
	if err != nil {
		fmt.Println("ERROR : Creating Avi Client : ", err)
		os.Exit(0)
	}
	err = lib.CreateApp(appName, namespace, 1)
	if err != nil {
		fmt.Println("ERROR : Creation of Deployment "+appName+" failed due to the error : ", err)
		os.Exit(0)
	}
	listOfServicesCreated, err = lib.CreateService(serviceNamePrefix, appName, namespace, 2)
	if err != nil {
		fmt.Println("ERROR : Creation of Services failed due to the error : ", err)
		os.Exit(0)
	}
}

/* Cleanup of Services and Deployment created for the test */
func Cleanup() {
	err := lib.DeleteService(listOfServicesCreated, namespace)
	if err != nil {
		fmt.Println("ERROR : Cleanup of Services ", listOfServicesCreated, " failed due to the error : ", err)
	}
	err = lib.DeleteApp(appName, namespace)
	if err != nil {
		fmt.Println("ERROR : Cleanup of Deployment "+appName+" failed due to the error : ", err)
	}
}

func ExitWithErrorf(t *testing.T, template string, args ...interface{}) {
	t.Errorf(template, args...)
	os.Exit(1)
}

/* Need to be executed for each test case
Fetches the Avi controller state before the testing starts */
func SetupForTesting(t *testing.T) {
	pools := lib.FetchPools(t, AviClients[0])
	initialNumOfPools = len(pools)
	VSes := lib.FetchVirtualServices(t, AviClients[0])
	initialVSesList = []string{}
	for _, vs := range VSes {
		initialVSesList = append(initialVSesList, *vs.Name)
	}
	initialNumOfVSes = len(initialVSesList)
	initialFQDNList = lib.FetchDNSARecordsFQDN(t, dnsVSUUID, AviClients[0])
	initialNumOfFQDN = len(initialFQDNList)
	ingressHostNames = []string{}
	ingressSecureHostNames = []string{}
	ingressInsecureHostNames = []string{}
	ingressesCreated = []string{}
	ingressesDeleted = []string{}
	ingressesUpdated = []string{}
}

/* Used for Controller and Node reboot */
func Reboot(t *testing.T, wg *sync.WaitGroup, nodeType string, controllerIP string, username string, password string) {
	t.Logf("Rebooting %s ... ", nodeType)
	loginID := username + "@" + controllerIP
	cmd := exec.Command("sshpass", "-p", password, "ssh", "-t", loginID, " `echo ", password, " |  sudo -S shutdown --reboot 0 && exit `")
	_, err := cmd.Output()
	if err != nil {
		t.Errorf("Cannot reboot %s because : %v ", nodeType, err.Error())
	} else {
		t.Logf("%s Rebooted", nodeType)
	}
	defer wg.Done()
}

/* Reboots AKO pod */
func RebootAko(t *testing.T, wg *sync.WaitGroup) {
	t.Logf("Rebooting AKO pod %s of namespace %s ...", akoPodName, AKONAMESPACE)
	err := lib.DeletePod(akoPodName, AKONAMESPACE)
	if err != nil {
		ExitWithErrorf(t, "Cannot reboot Ako pod as : %v", err)
	}
	t.Logf("Ako rebooted.")
	defer wg.Done()
}

/* Reboots Controller/Node/Ako if Reboot is set to true */
func CheckReboot(t *testing.T, wg *sync.WaitGroup) {
	if REBOOTAKO == true {
		wg.Add(1)
		go RebootAko(t, wg)
	}
	if REBOOTCONTROLLER == true {
		wg.Add(1)
		go Reboot(t, wg, CONTROLLER, os.Getenv("CTRL_IPADDRESS"), os.Getenv("CTRL_USERNAME"), os.Getenv("CTRL_PASSWORD"))
	}
	if REBOOTNODE == true {
		wg.Add(1)
		var testbedParams lib.TestbedFields
		testbed, err := os.Open(testbedFileName)
		if err != nil {
			t.Fatalf("ERROR : Error opening testbed file %s with error : %s", testbedFileName, err)
			os.Exit(0)
		}
		defer testbed.Close()
		byteValue, err := ioutil.ReadAll(testbed)
		if err != nil {
			t.Fatalf("ERROR : Failed to read the testbed file with error : %s", err)
			os.Exit(0)
		}
		json.Unmarshal(byteValue, &testbedParams)
		go Reboot(t, wg, KUBENODE, testbedParams.AkoParam.Clusters[0].KubeNodes[0].IP, testbedParams.AkoParam.Clusters[0].KubeNodes[0].UserName, testbedParams.AkoParam.Clusters[0].KubeNodes[0].Password)
	}
}

/* Gives all the elements presents in one list but not the other */
func DiffOfLists(list1 []string, list2 []string) []string {
	diffMap := map[string]int{}
	var diffString []string
	for _, l1 := range list1 {
		diffMap[l1] = 1
	}
	for _, l2 := range list2 {
		diffMap[l2] = diffMap[l2] + 1
	}
	var diffNum int
	for key, val := range diffMap {
		if val == 1 {
			diffNum = diffNum + 1
			diffString = append(diffString, key)
		}
	}
	return diffString
}

/* list1 - list2 */
func DiffOfListsOrderBased(list1 []string, list2 []string) []string {
	diffMap := map[string]bool{}
	var diffString []string
	for _, l1 := range list1 {
		diffMap[l1] = true
	}
	for _, l2 := range list2 {
		if _, ok := diffMap[l2]; !ok {
			diffString = append(diffString, l2)
		}
	}
	return diffString
}

/* Verifies if all requires pools are created or not */
func PoolVerification(t *testing.T) bool {
	t.Logf("Verifying pools...")
	pools := lib.FetchPools(t, AviClients[0])
	if ingressType == MULTIHOST && (len(pools) < ((len(ingressesCreated) * 2) + initialNumOfPools)) {
		return false
	} else if len(pools) < len(ingressesCreated)+initialNumOfPools {
		return false
	}
	var ingressPoolList []string
	var poolList []string
	if ingressType == INSECURE {
		for i := 0; i < len(ingressHostNames); i++ {
			ingressPoolName := clusterName + "--" + ingressHostNames[i] + "_-" + namespace + "-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
		}
	} else if ingressType == SECURE {
		for i := 0; i < len(ingressHostNames); i++ {
			ingressPoolName := clusterName + "--" + namespace + "-" + ingressHostNames[i] + "_-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
		}
	} else if ingressType == MULTIHOST {
		for i := 0; i < len(ingressSecureHostNames); i++ {
			ingressPoolName := clusterName + "--" + namespace + "-" + ingressSecureHostNames[i] + "_-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
			ingressPoolName = clusterName + "--" + ingressInsecureHostNames[i] + "_-" + namespace + "-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
		}
	}
	for _, pool := range pools {
		poolList = append(poolList, *pool.Name)
	}
	diffNum := len(DiffOfLists(ingressPoolList, poolList))
	if diffNum == initialNumOfPools {
		return true
	}
	return false
}

/* Verifies if all requires DNS A records are created in the DNS VS or not */
func DNSARecordsVerification(t *testing.T, hostNames []string) bool {
	t.Logf("Verifying DNS A Records...")
	FQDNList := lib.FetchDNSARecordsFQDN(t, dnsVSUUID, AviClients[0])
	diffString := DiffOfLists(FQDNList, hostNames)
	if len(diffString) == initialNumOfFQDN {
		return true
	}
	newSharedVSFQDN := DiffOfLists(diffString, initialFQDNList)
	var val int
	for _, fqdn := range newSharedVSFQDN {
		if strings.HasPrefix(fqdn, clusterName+"--shared") == true {
			val++
		}
	}
	if (len(newSharedVSFQDN) - val) == 0 {
		return true
	}
	return false
}

/* Verifies if all requires VSes for secure ingresses are created or not */
func VSVerification(t *testing.T) bool {
	t.Logf("Verifying VSes...")
	VSes := lib.FetchVirtualServices(t, AviClients[0])
	var ingressVSList []string
	var VSList []string
	for _, ing := range ingressesCreated {
		if ingressType != MULTIHOST {
			ingressVSName := clusterName + "--" + ing + ".avi.internal"
			ingressVSList = append(ingressVSList, ingressVSName)
		} else {
			ingressVSName := clusterName + "--" + ing + "-secure.avi.internal"
			ingressVSList = append(ingressVSList, ingressVSName)
		}
	}
	for _, vs := range VSes {
		VSList = append(VSList, *vs.Name)
	}
	diffString := DiffOfLists(ingressVSList, VSList)
	if len(diffString) == initialNumOfVSes {
		return true
	}
	newSharedVSesCreated := DiffOfLists(diffString, initialVSesList)
	var val int = 0
	for _, vs := range newSharedVSesCreated {
		if strings.HasPrefix(vs, clusterName+"--Shared") == true {
			val++
		}
	}
	if (len(newSharedVSesCreated) - val) == 0 {
		return true
	}
	return false
}

/* Calls Pool, VS and DNS A records verification based on the ingress type */
func Verify(t *testing.T) bool {
	if ingressType == SECURE {
		if PoolVerification(t) == true && VSVerification(t) == true && DNSARecordsVerification(t, ingressHostNames) == true {
			t.Logf("Pools, VSes and DNS A Records verified")
			return true
		}
	} else if ingressType == MULTIHOST {
		hostName := append(ingressSecureHostNames, ingressInsecureHostNames...)
		if PoolVerification(t) == true && VSVerification(t) == true && DNSARecordsVerification(t, hostName) == true {
			t.Logf("Pools, VSes and DNS A Records verified")
			return true
		}
	} else if ingressType == INSECURE {
		if PoolVerification(t) == true && DNSARecordsVerification(t, ingressHostNames) == true {
			t.Logf("Pools and DNS A Records verified")
			return true
		}
	}
	return false
}

func parallelInsecureIngressCreation(t *testing.T, wg *sync.WaitGroup, serviceName string, namespace string, numOfIng int, startIndex int) {
	defer wg.Done()
	ingresses, hostNames, err := lib.CreateInsecureIngress(ingressNamePrefix, serviceName, namespace, numOfIng, startIndex)
	if err != nil {
		ExitWithErrorf(t, "Failed to create %s ingresses as : %v", ingressType, err)
	}
	ingressesCreated = append(ingressesCreated, ingresses...)
	ingressHostNames = append(ingressHostNames, hostNames...)
}

func parallelSecureIngressCreation(t *testing.T, wg *sync.WaitGroup, serviceName string, namespace string, numOfIng int, startIndex int) {
	defer wg.Done()
	ingresses, hostNames, err := lib.CreateSecureIngress(ingressNamePrefix, serviceName, namespace, numOfIng, startIndex)
	if err != nil {
		ExitWithErrorf(t, "Failed to create %s ingresses as : %v", ingressType, err)
	}
	ingressesCreated = append(ingressesCreated, ingresses...)
	ingressHostNames = append(ingressHostNames, hostNames...)
}

func parallelMultiHostIngressCreation(t *testing.T, wg *sync.WaitGroup, serviceName []string, namespace string, numOfIng int, startIndex int) {
	defer wg.Done()
	ingresses, secureHostNames, insecureHostNames, err := lib.CreateMultiHostIngress(ingressNamePrefix, serviceName, namespace, numOfIng, startIndex)
	if err != nil {
		ExitWithErrorf(t, "Failed to create %s ingresses as : %v", ingressType, err)
	}
	ingressesCreated = append(ingressesCreated, ingresses...)
	ingressSecureHostNames = append(ingressSecureHostNames, secureHostNames...)
	ingressInsecureHostNames = append(ingressInsecureHostNames, insecureHostNames...)
}

func parallelIngressDeletion(t *testing.T, wg *sync.WaitGroup, namespace string, listOfIngressToDelete []string) {
	defer wg.Done()
	ingresses, err := lib.DeleteIngress(namespace, listOfIngressToDelete)
	if err != nil {
		ExitWithErrorf(t, "Failed to delete ingresses as : %v", err)
	}
	ingressesDeleted = append(ingressesDeleted, ingresses...)
}

func parallelIngressUpdation(t *testing.T, wg *sync.WaitGroup, namespace string, listofIngressToUpdate []string) {
	defer wg.Done()
	ingresses, err := lib.UpdateIngress(namespace, listofIngressToUpdate)
	if err != nil {
		ExitWithErrorf(t, "Failed to update ingresses as : %v", err)
	}
	ingressesUpdated = append(ingressesUpdated, ingresses...)
}

func CreateIngressesParallel(t *testing.T, numOfIng int, initialNumOfPools int) {
	ingressesCreated = []string{}
	var blockSize = numOfIng / numGoRoutines
	var remIng = numOfIng % numGoRoutines
	g := gomega.NewGomegaWithT(t)
	var wg sync.WaitGroup
	nextStartInd := 0
	switch {
	case ingressType == INSECURE:
		t.Logf("Creating %d %s Ingresses Parallelly...", numOfIng, ingressType)
		CheckReboot(t, &wg)
		for i := 0; i < numGoRoutines; i++ {
			wg.Add(1)
			if i+1 <= remIng {
				go parallelInsecureIngressCreation(t, &wg, listOfServicesCreated[0], namespace, blockSize+1, nextStartInd)
				nextStartInd = nextStartInd + blockSize + 1
			} else {
				go parallelInsecureIngressCreation(t, &wg, listOfServicesCreated[0], namespace, blockSize, nextStartInd)
				nextStartInd = nextStartInd + blockSize
			}
		}
	case ingressType == SECURE:
		t.Logf("Creating %d %s Ingresses Parallelly...", numOfIng, ingressType)
		CheckReboot(t, &wg)
		for i := 0; i < numGoRoutines; i++ {
			wg.Add(1)
			if i+1 <= remIng {
				go parallelSecureIngressCreation(t, &wg, listOfServicesCreated[0], namespace, blockSize+1, nextStartInd)
				nextStartInd = nextStartInd + blockSize + 1
			} else {
				go parallelSecureIngressCreation(t, &wg, listOfServicesCreated[0], namespace, blockSize, nextStartInd)
				nextStartInd = nextStartInd + blockSize
			}
		}
	case ingressType == MULTIHOST:
		t.Logf("Creating %d %s Ingresses Parallelly...", numOfIng, ingressType)
		CheckReboot(t, &wg)
		for i := 0; i < numGoRoutines; i++ {
			wg.Add(1)
			if (i + 1) <= remIng {
				go parallelMultiHostIngressCreation(t, &wg, listOfServicesCreated, namespace, blockSize+1, nextStartInd)
				nextStartInd = nextStartInd + blockSize + 1
			} else {
				go parallelMultiHostIngressCreation(t, &wg, listOfServicesCreated, namespace, blockSize, nextStartInd)
				nextStartInd = nextStartInd + blockSize
			}
		}
	}
	wg.Wait()
	g.Expect(ingressesCreated).To(gomega.HaveLen(numOfIng))
	t.Logf("Created %d %s Ingresses Parallelly", numOfIng, ingressType)
	t.Logf("Verifiying Avi objects ...")
	pollInterval, _ := time.ParseDuration(testPollInterval)
	waitTimeIncr, _ := strconv.Atoi(testPollInterval[:len(testPollInterval)-1])
	// Verifies for Avi objects creation by checking every 'waitTime' seconds for 'testCaseTimeOut' seconds
	for waitTime := 0; waitTime < testCaseTimeOut; {
		if Verify(t) == true {
			t.Logf("Created %d Ingresses and associated Avi objects\n", numOfIng)
			return
		}
		time.Sleep(pollInterval)
		waitTime = waitTime + waitTimeIncr
	}
	t.Fatalf("Error : Verification failed\n")
}

func DeleteIngressesParallel(t *testing.T, numOfIng int, initialNumOfPools int) {
	var blockSize = numOfIng / numGoRoutines
	var remIng = numOfIng % numGoRoutines
	g := gomega.NewGomegaWithT(t)
	var wg sync.WaitGroup
	ingressesDeleted = []string{}
	t.Logf("Deleting %d %s Ingresses...", numOfIng, ingressType)
	nextStartInd := 0
	CheckReboot(t, &wg)
	for i := 0; i < numGoRoutines; i++ {
		wg.Add(1)
		if (i + 1) <= remIng {
			go parallelIngressDeletion(t, &wg, namespace, ingressesCreated[nextStartInd:nextStartInd+blockSize+1])
			nextStartInd = nextStartInd + blockSize + 1
		} else {
			go parallelIngressDeletion(t, &wg, namespace, ingressesCreated[nextStartInd:nextStartInd+blockSize])
			nextStartInd = nextStartInd + blockSize
		}
	}
	wg.Wait()
	g.Expect(ingressesDeleted).To(gomega.HaveLen(numOfIng))
	t.Logf("Deleted %d %s Ingresses", numOfIng, ingressType)
	t.Logf("Verifiying Avi objects ...")
	g.Eventually(func() int {
		pools := lib.FetchPools(t, AviClients[0])
		return len(pools)
	}, testCaseTimeOut, testPollInterval).Should(gomega.Equal(initialNumOfPools))
	t.Logf("Deleted %d Ingresses and associated Avi objects\n", numOfIng)
}

func UpdateIngressesParallel(t *testing.T, numOfIng int) {
	var blockSize = numOfIng / numGoRoutines
	var remIng = numOfIng % numGoRoutines
	g := gomega.NewGomegaWithT(t)
	var wg sync.WaitGroup
	ingressesUpdated = []string{}
	t.Logf("Updating %d %s Ingresses...", numOfIng, ingressType)
	nextStartInd := 0
	CheckReboot(t, &wg)
	for i := 0; i < numGoRoutines; i++ {
		wg.Add(1)
		if (i + 1) <= remIng {
			go parallelIngressUpdation(t, &wg, namespace, ingressesCreated[nextStartInd:nextStartInd+blockSize+1])
			nextStartInd = nextStartInd + blockSize + 1
		} else {
			go parallelIngressUpdation(t, &wg, namespace, ingressesCreated[nextStartInd:nextStartInd+blockSize])
			nextStartInd = nextStartInd + blockSize
		}
	}
	wg.Wait()
	g.Expect(ingressesUpdated).To(gomega.HaveLen(numOfIng))
	t.Logf("Updated %d Ingresses\n", numOfIng)
}

func CreateIngressesSerial(t *testing.T, numOfIng int, initialNumOfPools int) {
	g := gomega.NewGomegaWithT(t)
	ingressesCreated = []string{}
	var err error
	switch {
	case ingressType == INSECURE:
		t.Logf("Creating %d %s Ingresses Serially...", numOfIng, ingressType)
		ingressesCreated, ingressHostNames, err = lib.CreateInsecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, numOfIng)
		if err != nil {
			t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
		}
	case ingressType == SECURE:
		t.Logf("Creating %d %s Ingresses Serially...", numOfIng, ingressType)
		ingressesCreated, ingressHostNames, err = lib.CreateSecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, numOfIng)
		if err != nil {
			t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
		}
	case ingressType == MULTIHOST:
		t.Logf("Creating %d %s Ingresses Serially...", numOfIng, ingressType)
		ingressesCreated, ingressSecureHostNames, ingressInsecureHostNames, err = lib.CreateMultiHostIngress(ingressNamePrefix, listOfServicesCreated, namespace, numOfIng)
		if err != nil {
			t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
		}
	}
	g.Expect(ingressesCreated).To(gomega.HaveLen(numOfIng))
	t.Logf("Created %d %s Ingresses Serially", numOfIng, ingressType)
	t.Logf("Verifiying Avi objects ...")
	pollInterval, _ := time.ParseDuration(testPollInterval)
	waitTimeIncr, _ := strconv.Atoi(testPollInterval[:len(testPollInterval)-1])
	for waitTime := 0; waitTime < testCaseTimeOut; {
		if Verify(t) == true {
			return
		}
		time.Sleep(pollInterval)
		waitTime = waitTime + waitTimeIncr
	}
	t.Fatalf("Error : Verification failed\n")

}

func DeleteIngressesSerial(t *testing.T, numOfIng int, initialNumOfPools int, AviClient *clients.AviClient) {
	g := gomega.NewGomegaWithT(t)
	ingressesDeleted = []string{}
	t.Logf("Deleting %d %s Ingresses Serially...", numOfIng, ingressType)
	ingressesDeleted, err := lib.DeleteIngress(namespace, ingressesCreated)
	if err != nil {
		t.Fatalf("Failed to delete ingresses as : %v", err)
	}
	g.Expect(ingressesDeleted).To(gomega.HaveLen(numOfIng))
	t.Logf("Deleted %d %s Ingresses Serially", numOfIng, ingressType)
	t.Logf("Verifiying Avi objects ...")
	g.Eventually(func() int {
		pools := lib.FetchPools(t, AviClient)
		return len(pools)
	}, testCaseTimeOut, testPollInterval).Should(gomega.Equal(initialNumOfPools))
	t.Logf("Deleted %d Pools", numOfIng)
}

func HybridCreation(t *testing.T, wg *sync.WaitGroup, numOfIng int, deletionStartPoint int) {
	for i := deletionStartPoint; i < numOfIng; i++ {
		mutex.Lock()
		wg.Add(1)
		var ingresses []string
		var err error
		switch {
		case ingressType == INSECURE:
			ingresses, _, err = lib.CreateInsecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, 1, i)
			if err != nil {
				ExitWithErrorf(t, "Failed to create %s ingresses as : %v", ingressType, err)
			}
		case ingressType == SECURE:
			ingresses, _, err = lib.CreateSecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, 1, i)
			if err != nil {
				ExitWithErrorf(t, "Failed to create %s ingresses as : %v", ingressType, err)
			}
		case ingressType == MULTIHOST:
			ingresses, _, _, err = lib.CreateMultiHostIngress(ingressNamePrefix, listOfServicesCreated, namespace, 1, i)
			if err != nil {
				ExitWithErrorf(t, "Failed to create %s ingresses as : %v", ingressType, err)
			}
		}
		t.Logf("Created ingresses %s", ingresses)
		ingressesCreated = append(ingressesCreated, ingresses...)
		mutex.Unlock()
		defer wg.Done()

	}
	defer wg.Done()
}

func HybridUpdation(t *testing.T, wg *sync.WaitGroup, numOfIng int) {
	for len(ingressesUpdated) < numOfIng {
		mutex.Lock()
		wg.Add(1)
		tempStr := DiffOfLists(ingressesCreated, ingressesDeleted)
		toUpdateIngresses := DiffOfListsOrderBased(ingressesUpdated, tempStr)
		if len(toUpdateIngresses) > 0 {
			updatedIngresses, err := lib.UpdateIngress(namespace, toUpdateIngresses)
			if err != nil {
				ExitWithErrorf(t, "Error updating ingresses as : %v ", err)
				return
			}
			t.Logf("Updated ingresses %s", updatedIngresses)
			ingressesUpdated = append(ingressesUpdated, updatedIngresses...)
		}
		mutex.Unlock()
		defer wg.Done()
	}
	defer wg.Done()
}

func HybridDeletion(t *testing.T, wg *sync.WaitGroup, numOfIng int) {
	for len(ingressesDeleted) < numOfIng {
		mutex.Lock()
		wg.Add(1)
		toDeleteIngresses := DiffOfLists(ingressesCreated, ingressesDeleted)
		if len(toDeleteIngresses) > 0 {
			deletedIngresses, err := lib.DeleteIngress(namespace, toDeleteIngresses)
			if err != nil {
				ExitWithErrorf(t, "Error deleting ingresses as : %v ", err)
			}
			t.Logf("Deleted ingresses %s", deletedIngresses)
			ingressesDeleted = append(ingressesDeleted, deletedIngresses...)
		}
		mutex.Unlock()
		defer wg.Done()
	}
	defer wg.Done()
}

/* Creates some(deletionStartPoint) ingresses first, followed by creation, updation and deletion of ingresses parallelly */
func HybridExecution(t *testing.T, numOfIng int, deletionStartPoint int) {
	g := gomega.NewGomegaWithT(t)
	var wg sync.WaitGroup
	var err error
	ingressesCreated = []string{}
	ingressesUpdated = []string{}
	ingressesDeleted = []string{}
	switch {
	case ingressType == INSECURE:
		t.Logf("Creating %d %s Ingresses...", deletionStartPoint, ingressType)
		ingressesCreated, _, err = lib.CreateInsecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, deletionStartPoint)
		if err != nil {
			t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
		}
	case ingressType == SECURE:
		t.Logf("Creating %d %s Ingresses...", deletionStartPoint, ingressType)
		ingressesCreated, _, err = lib.CreateSecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, deletionStartPoint)
		if err != nil {
			t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
		}
	case ingressType == MULTIHOST:
		t.Logf("Creating %d %s Ingresses...", deletionStartPoint, ingressType)
		ingressesCreated, _, _, err = lib.CreateMultiHostIngress(ingressNamePrefix, listOfServicesCreated, namespace, deletionStartPoint)
		if err != nil {
			t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
		}
	}
	wg.Add(3)
	go HybridCreation(t, &wg, numOfIng, deletionStartPoint)
	go HybridUpdation(t, &wg, numOfIng/2)
	go HybridDeletion(t, &wg, numOfIng)
	wg.Wait()
	g.Expect(ingressesCreated).To(gomega.HaveLen(numOfIng))
	g.Expect(ingressesDeleted).To(gomega.HaveLen(numOfIng))
}

func CreateIngressParallelWithAkoReboot(t *testing.T) {
	REBOOTAKO = true
	CreateIngressesParallel(t, numOfIng, initialNumOfPools)
	REBOOTAKO = false
}

func CreateIngressParallelWithControllerReboot(t *testing.T) {
	REBOOTCONTROLLER = true
	CreateIngressesParallel(t, numOfIng, initialNumOfPools)
	REBOOTCONTROLLER = false
}

func CreateIngressParallelWithNodeReboot(t *testing.T) {
	REBOOTNODE = true
	CreateIngressesParallel(t, numOfIng, initialNumOfPools)
	REBOOTNODE = false
}

func DeleteIngressParallelWithAkoReboot(t *testing.T) {
	REBOOTAKO = true
	DeleteIngressesParallel(t, numOfIng, initialNumOfPools)
	REBOOTAKO = false
}

func DeleteIngressParallelWithControllerReboot(t *testing.T) {
	REBOOTCONTROLLER = true
	DeleteIngressesParallel(t, numOfIng, initialNumOfPools)
	REBOOTCONTROLLER = false
}

func DeleteIngressParallelWithNodeReboot(t *testing.T) {
	REBOOTNODE = true
	DeleteIngressesParallel(t, numOfIng, initialNumOfPools)
	REBOOTNODE = false
}

func UpdateIngressParallelWithAkoReboot(t *testing.T) {
	REBOOTAKO = true
	UpdateIngressesParallel(t, numOfIng)
	REBOOTAKO = false
}

func UpdateIngressParallelWithControllerReboot(t *testing.T) {
	REBOOTCONTROLLER = true
	UpdateIngressesParallel(t, numOfIng)
	REBOOTCONTROLLER = false
}

func UpdateIngressParallelWithNodeReboot(t *testing.T) {
	REBOOTNODE = true
	UpdateIngressesParallel(t, numOfIng)
	REBOOTNODE = false
}

func CreateServiceTypeLBWithApp(t *testing.T, numPods int, numOfServices int, appNameLB string, serviceNamePrefixLB string, aviObjPrefix string) []string {
	g := gomega.NewGomegaWithT(t)
	t.Logf("Creating a %v deployment with %v replicas", appNameLB, numPods)
	err := lib.CreateApp(appNameLB, namespace, numPods)
	if err != nil {
		t.Fatalf("ERROR : Could not create deployment for service type LB support as %v", err)
	}

	t.Logf("Creating %v services of type LB", numOfServices)
	servicesCreated, port, err := lib.CreateLBService(serviceNamePrefixLB, appNameLB, namespace, numOfServices)
	if err != nil {
		t.Fatalf("ERROR : Could not create %d Services of type LB as %v", numOfServices, err)
	}
	t.Logf("Verifying AVI object creation...")
	g.Eventually(func() bool {
		var VSList []string
		var poolList []string
		/* Verifying pool creation*/
		for _, svc := range servicesCreated {
			VSList = append(VSList, aviObjPrefix+svc)
			poolList = append(poolList, aviObjPrefix+svc+"--"+port)
		}
		pools := lib.FetchPools(t, AviClients[0])
		var aviPoolList []string
		for _, pool := range pools {
			aviPoolList = append(aviPoolList, *pool.Name)
		}
		diffNum := len(DiffOfLists(poolList, aviPoolList))
		if diffNum != initialNumOfPools {
			return false
		}
		/* Verification of servers on pools */
		for _, pool := range pools {
			if strings.HasPrefix(*pool.Name, aviObjPrefix+serviceNamePrefixLB) == true {
				if len(pool.Servers) != numPods {
					return false
				}
			}
		}
		/* Verifying VS creation */
		VSes := lib.FetchVirtualServices(t, AviClients[0])
		var svcLBVSList []string
		for _, vs := range VSes {
			svcLBVSList = append(svcLBVSList, *vs.Name)
		}
		diffNum = len(DiffOfLists(VSList, svcLBVSList))
		if diffNum != initialNumOfVSes {
			return false
		}
		t.Logf("Verified pools, servers on pools and VSes")
		return true
	}, testCaseTimeOut, testPollInterval).Should(gomega.Equal(true))
	return servicesCreated
}

func DeleteLBDeployment(t *testing.T, deploymentName string, serviceNamePrefixLB string, aviObjPrefix string) {
	g := gomega.NewGomegaWithT(t)
	t.Logf("Deleting deployment %v...", deploymentName)
	err := lib.DeleteApp(deploymentName, namespace)
	if err != nil {
		t.Fatalf("Error deleting the deployment of LB service as : %v", err)
	}
	t.Logf("Deleted deployment %v", deploymentName)
	t.Logf("Verifying AVI object deletion")
	g.Eventually(func() bool {
		pools := lib.FetchPools(t, AviClients[0])
		for _, pool := range pools {
			if strings.HasPrefix(*pool.Name, aviObjPrefix+serviceNamePrefixLB) == true {
				if len(pool.Servers) != 0 {
					return false
				}
			}
		}
		return true
	}, testCaseTimeOut, testPollInterval).Should(gomega.Equal(true))
}

func DeleteServiceTypeLB(t *testing.T, serviceList []string) {
	t.Logf("Deleting LB services...")
	g := gomega.NewGomegaWithT(t)
	err := lib.DeleteService(serviceList, namespace)
	if err != nil {
		t.Fatalf("ERROR : Deleting services of type LB %v", err)
	}
	t.Logf("Deleted LB services...")
	t.Logf("Verifying AVI object deletion")
	g.Eventually(func() int {
		pools := lib.FetchPools(t, AviClients[0])
		return len(pools)
	}, testCaseTimeOut, testPollInterval).Should(gomega.Equal(initialNumOfPools))
	t.Logf("Pools verified")
	g.Eventually(func() int {
		VSes := lib.FetchVirtualServices(t, AviClients[0])
		return len(VSes)
	}, testCaseTimeOut, testPollInterval).Should(gomega.Equal(initialNumOfVSes))
	t.Logf("VSes verified")
}

func LBService(t *testing.T) {
	appNameLB := "lb-" + appName
	serviceNamePrefixLB := "lb-" + serviceNamePrefix
	aviObjPrefix := clusterName + "--" + namespace + "-"
	serviceList := CreateServiceTypeLBWithApp(t, numOfPodsForLBSvc, numOfLBSvc, appNameLB, serviceNamePrefixLB, aviObjPrefix)
	DeleteLBDeployment(t, appNameLB, serviceNamePrefixLB, aviObjPrefix)
	DeleteServiceTypeLB(t, serviceList)

}
func TestMain(t *testing.M) {
	Setup()
	defer Cleanup()
	os.Exit(t.Run())
}

func TestInsecureParallelCreationUpdationDeletionWithoutReboot(t *testing.T) {
	SetupForTesting(t)
	ingressType = INSECURE
	CreateIngressesParallel(t, numOfIng, initialNumOfPools)
	UpdateIngressesParallel(t, numOfIng)
	DeleteIngressesParallel(t, numOfIng, initialNumOfPools)
}

func TestInsecureParallelCreationUpdationDeletionWithAkoReboot(t *testing.T) {
	SetupForTesting(t)
	ingressType = INSECURE
	CreateIngressParallelWithAkoReboot(t)
	UpdateIngressParallelWithAkoReboot(t)
	DeleteIngressParallelWithAkoReboot(t)
}

func TestInsecureParallelCreationUpdationDeletionWithNodeReboot(t *testing.T) {
	SetupForTesting(t)
	ingressType = INSECURE
	CreateIngressParallelWithNodeReboot(t)
	UpdateIngressParallelWithNodeReboot(t)
	DeleteIngressParallelWithNodeReboot(t)
}

func TestInsecureParallelCreationUpdationDeletionWithControllerReboot(t *testing.T) {
	SetupForTesting(t)
	ingressType = INSECURE
	CreateIngressParallelWithControllerReboot(t)
	UpdateIngressParallelWithControllerReboot(t)
	DeleteIngressParallelWithControllerReboot(t)
}

func TestSecureParallelCreationUpdationDeletionWithoutReboot(t *testing.T) {
	SetupForTesting(t)
	ingressType = SECURE
	CreateIngressesParallel(t, numOfIng, initialNumOfPools)
	UpdateIngressesParallel(t, numOfIng)
	DeleteIngressesParallel(t, numOfIng, initialNumOfPools)
}

func TestSecureParallelCreationUpdationDeletionWithAkoReboot(t *testing.T) {
	SetupForTesting(t)
	ingressType = SECURE
	CreateIngressParallelWithAkoReboot(t)
	UpdateIngressParallelWithAkoReboot(t)
	DeleteIngressParallelWithAkoReboot(t)
}

func TestSecureParallelCreationUpdationDeletionWithNodeReboot(t *testing.T) {
	SetupForTesting(t)
	ingressType = SECURE
	CreateIngressParallelWithNodeReboot(t)
	UpdateIngressParallelWithNodeReboot(t)
	DeleteIngressParallelWithNodeReboot(t)
}

func TestSecureParallelCreationUpdationDeletionWithControllerReboot(t *testing.T) {
	SetupForTesting(t)
	ingressType = SECURE
	CreateIngressParallelWithControllerReboot(t)
	UpdateIngressParallelWithControllerReboot(t)
	DeleteIngressParallelWithControllerReboot(t)
}

func TestMultiHostParallelCreationUpdationDeletionWithoutReboot(t *testing.T) {
	SetupForTesting(t)
	ingressType = MULTIHOST
	CreateIngressesParallel(t, numOfIng, initialNumOfPools)
	UpdateIngressesParallel(t, numOfIng)
	DeleteIngressesParallel(t, numOfIng, initialNumOfPools)
}

func TestMultiHostParallelCreationUpdationDeletionWithAkoReboot(t *testing.T) {
	SetupForTesting(t)
	ingressType = MULTIHOST
	CreateIngressParallelWithAkoReboot(t)
	UpdateIngressParallelWithAkoReboot(t)
	DeleteIngressParallelWithAkoReboot(t)
}

func TestMultiHostParallelCreationUpdationDeletionWithNodeReboot(t *testing.T) {
	SetupForTesting(t)
	ingressType = MULTIHOST
	CreateIngressParallelWithNodeReboot(t)
	UpdateIngressParallelWithNodeReboot(t)
	DeleteIngressParallelWithNodeReboot(t)
}

func TestMultiHostParallelCreationUpdationDeletionWithControllerReboot(t *testing.T) {
	SetupForTesting(t)
	ingressType = MULTIHOST
	CreateIngressParallelWithControllerReboot(t)
	UpdateIngressParallelWithControllerReboot(t)
	DeleteIngressParallelWithControllerReboot(t)
}

func TestInsecureHybridExecution(t *testing.T) {
	SetupForTesting(t)
	ingressType = INSECURE
	HybridExecution(t, numOfIng, numOfIng/2)
}

func TestSecureHybridExecution(t *testing.T) {
	SetupForTesting(t)
	ingressType = SECURE
	HybridExecution(t, numOfIng, numOfIng/2)
}

func TestMultiHostHybridExecution(t *testing.T) {
	SetupForTesting(t)
	ingressType = MULTIHOST
	HybridExecution(t, numOfIng, numOfIng/2)
}

func TestServiceTypeLB(t *testing.T) {
	SetupForTesting(t)
	LBService(t)
}
