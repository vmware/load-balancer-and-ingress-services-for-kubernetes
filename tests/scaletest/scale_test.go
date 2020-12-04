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

package scaletest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/avinetworks/sdk/go/clients"
	"github.com/onsi/gomega"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/scaletest/lib"
)

const SECURE = "secure"
const INSECURE = "insecure"
const MULTIHOST = "multi-host"

var testbedFileName string
var namespace string
var appName string
var serviceNamePrefix string
var ingressNamePrefix string
var AviClients []*clients.AviClient
var numGoRoutines int
var listOfServicesCreated []string
var ingressesCreated []string
var ingressesDeleted []string
var initialNumOfPools = 0
var initialNumOfVSes = 0
var ingressType string
var clusterName string
var timeout string
var testCaseTimeOut = 1800
var testPollInterval = "15s"

func Setup() {
	var testbedParams lib.TestbedFields
	timeout = os.Args[3]
	testbedFileName = os.Args[5]
	testbed, err := os.Open(testbedFileName)
	if err != nil {
		fmt.Println("ERROR : Error opening testbed file ", testbedFileName, " with error : ", err)
		os.Exit(0)
	}
	defer testbed.Close()
	byteValue, _ := ioutil.ReadAll(testbed)
	json.Unmarshal(byteValue, &testbedParams)
	numGoRoutines, err = strconv.Atoi(os.Args[4])
	if err != nil {
		numGoRoutines = 5
	}
	if numGoRoutines <= 0 {
		fmt.Println("ERROR : Number of Go Routines cannot be zero or negative.")
		os.Exit(0)
	}
	namespace = testbedParams.TestParams.Namespace
	appName = testbedParams.TestParams.AppName
	serviceNamePrefix = testbedParams.TestParams.ServiceNamePrefix
	ingressNamePrefix = testbedParams.TestParams.IngressNamePrefix
	clusterName = testbedParams.AkoParam.Clusters[0].ClusterName
	os.Setenv("CTRL_USERNAME", testbedParams.Vm[0].UserName)
	os.Setenv("CTRL_PASSWORD", testbedParams.Vm[0].Password)
	os.Setenv("CTRL_IPADDRESS", testbedParams.Vm[0].Ip)

	lib.KubeInit(testbedParams.AkoParam.Clusters[0].KubeConfigFilePath)
	AviClients, err = lib.SharedAVIClients(2)
	if err != nil {
		fmt.Println("ERROR : Creating Avi Client : ", err)
		os.Exit(0)
	}
	err = lib.CreateApp(appName, namespace)
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

func PoolVerification(t *testing.T) bool {
	pools := lib.FetchPools(t, AviClients[0])
	if ingressType == MULTIHOST && (len(pools) < ((len(ingressesCreated) * 2) + initialNumOfPools)) {
		return false
	} else if len(pools) < len(ingressesCreated)+initialNumOfPools {
		return false
	}
	var ingressPoolList []string
	var poolList []string
	for i := 0; i < len(ingressesCreated); i++ {
		if ingressType == INSECURE {
			ingressPoolName := clusterName + "--" + ingressesCreated[i] + ".avi.internal-" + namespace + "-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
		} else if ingressType == SECURE {
			ingressPoolName := clusterName + "--" + namespace + "-" + ingressesCreated[i] + ".avi.internal-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
		} else if ingressType == MULTIHOST {
			ingressPoolName := clusterName + "--" + namespace + "-" + ingressesCreated[i] + "-secure.avi.internal-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
			ingressPoolName = clusterName + "--" + ingressesCreated[i] + "-insecure.avi.internal-" + namespace + "-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
		}
	}
	for i := 0; i < len(pools); i++ {
		poolList = append(poolList, *pools[i].Name)
	}
	diffNum := len(DiffOfLists(ingressPoolList, poolList))
	if diffNum == initialNumOfPools {
		return true
	}
	return false
}

func VSVerification(t *testing.T) bool {
	VSes := lib.FetchVirtualServices(t, AviClients[0])
	var ingressVSList []string
	var VSList []string
	for i := 0; i < len(ingressesCreated); i++ {
		if ingressType != MULTIHOST {
			ingressVSName := clusterName + "--" + ingressesCreated[i] + ".avi.internal"
			ingressVSList = append(ingressVSList, ingressVSName)
		} else {
			ingressVSName := clusterName + "--" + ingressesCreated[i] + "-secure.avi.internal"
			ingressVSList = append(ingressVSList, ingressVSName)
		}
	}
	for i := 0; i < len(VSes); i++ {
		VSList = append(VSList, *VSes[i].Name)
	}
	diffNum := len(DiffOfLists(ingressVSList, VSList))
	if diffNum == initialNumOfVSes {
		return true
	}
	return false
}

func Verify(t *testing.T) bool {
	if ingressType != INSECURE {
		if PoolVerification(t) == true && VSVerification(t) == true {
			t.Logf("Pools and VSes verified")
			return true
		}
	} else {
		if PoolVerification(t) == true {
			t.Logf("Pools verified")
			return true
		}
	}
	return false
}

func parallelInsecureIngressCreation(t *testing.T, wg *sync.WaitGroup, serviceName string, namespace string, numOfIng int, startIndex int) {
	defer wg.Done()
	ingresses, err := lib.CreateInsecureIngress(ingressNamePrefix, serviceName, namespace, numOfIng, startIndex)
	if err != nil {
		t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
	}
	ingressesCreated = append(ingressesCreated, ingresses...)
}

func parallelSecureIngressCreation(t *testing.T, wg *sync.WaitGroup, serviceName string, namespace string, numOfIng int, startIndex int) {
	defer wg.Done()
	ingresses, err := lib.CreateSecureIngress(ingressNamePrefix, serviceName, namespace, numOfIng, startIndex)
	if err != nil {
		t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
	}
	ingressesCreated = append(ingressesCreated, ingresses...)
}

func parallelMultiHostIngressCreation(t *testing.T, wg *sync.WaitGroup, serviceName []string, namespace string, numOfIng int, startIndex int) {
	defer wg.Done()
	ingresses, err := lib.CreateMultiHostIngress(ingressNamePrefix, serviceName, namespace, numOfIng, startIndex)
	if err != nil {
		t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
	}
	ingressesCreated = append(ingressesCreated, ingresses...)
}

func parallelIngressDeletion(t *testing.T, wg *sync.WaitGroup, namespace string, listOfIngressToDelete []string) {
	defer wg.Done()
	ingresses, err := lib.DeleteIngress(namespace, listOfIngressToDelete)
	if err != nil {
		t.Fatalf("Failed to delete ingresses as : %v", err)
	}
	ingressesDeleted = append(ingressesDeleted, ingresses...)
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
		t.Logf("Creating %d %s Ingresses Parallely...", numOfIng, ingressType)
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
		t.Logf("Creating %d %s Ingresses Parallely...", numOfIng, ingressType)
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
		t.Logf("Creating %d %s Ingresses Parallely...", numOfIng, ingressType)
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
	t.Logf("Created %d %s Ingresses Parallely", numOfIng, ingressType)
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

func DeleteIngressesParallel(t *testing.T, numOfIng int, initialNumOfPools int, AviClient *clients.AviClient) {
	var blockSize = numOfIng / numGoRoutines
	var remIng = numOfIng % numGoRoutines
	g := gomega.NewGomegaWithT(t)
	var wg sync.WaitGroup
	ingressesDeleted = []string{}
	t.Logf("Deleting %d %s Ingresses...", numOfIng, ingressType)
	nextStartInd := 0
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
		pools := lib.FetchPools(t, AviClient)
		return len(pools)
	}, testCaseTimeOut, testPollInterval).Should(gomega.Equal(initialNumOfPools))
	t.Logf("Deleted %d Pools", numOfIng)
}

func ParallelIngressHelper(t *testing.T, numOfIng int) {
	pools := lib.FetchPools(t, AviClients[0])
	initialNumOfPools = len(pools)
	VSes := lib.FetchVirtualServices(t, AviClients[0])
	initialNumOfVSes = len(VSes)
	CreateIngressesParallel(t, numOfIng, initialNumOfPools)
	DeleteIngressesParallel(t, numOfIng, initialNumOfPools, AviClients[0])
	t.Logf("%d %s Ingress creation deletion along with verification is done", numOfIng, ingressType)
}

func CreateIngressesSerial(t *testing.T, numOfIng int, initialNumOfPools int) {
	g := gomega.NewGomegaWithT(t)
	var err error
	switch {
	case ingressType == INSECURE:
		t.Logf("Creating %d %s Ingresses Serially...", numOfIng, ingressType)
		ingressesCreated, err = lib.CreateInsecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, numOfIng)
		if err != nil {
			t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
		}
	case ingressType == SECURE:
		t.Logf("Creating %d %s Ingresses Serially...", numOfIng, ingressType)
		ingressesCreated, err = lib.CreateSecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, numOfIng)
		if err != nil {
			t.Fatalf("Failed to create %s ingresses as : %v", ingressType, err)
		}
	case ingressType == MULTIHOST:
		t.Logf("Creating %d %s Ingresses Serially...", numOfIng, ingressType)
		ingressesCreated, err = lib.CreateMultiHostIngress(ingressNamePrefix, listOfServicesCreated, namespace, numOfIng)
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

func SerialIngressHelper(t *testing.T, numOfIng int) {
	pools := lib.FetchPools(t, AviClients[0])
	initialNumOfPools = len(pools)
	VSes := lib.FetchVirtualServices(t, AviClients[0])
	initialNumOfVSes = len(VSes)
	CreateIngressesSerial(t, numOfIng, initialNumOfPools)
	DeleteIngressesSerial(t, numOfIng, initialNumOfPools, AviClients[0])
	t.Logf("%d %s Ingress serially creation deletion along with verification is done", numOfIng, ingressType)
}

func TestMain(t *testing.M) {
	Setup()
	t.Run()
	Cleanup()
}

func TestParallel200InsecureIngresses(t *testing.T) {
	ingressType = INSECURE
	ParallelIngressHelper(t, 200)
}

func TestParallel200SecureIngresses(t *testing.T) {
	ingressType = SECURE
	ParallelIngressHelper(t, 200)
}

func TestParallel200MultiHostIngresses(t *testing.T) {
	ingressType = MULTIHOST
	ParallelIngressHelper(t, 200)
}

func TestSerial200InsecureIngresses(t *testing.T) {
	ingressType = INSECURE
	SerialIngressHelper(t, 200)
}

func TestSerial200SecureIngresses(t *testing.T) {
	ingressType = SECURE
	SerialIngressHelper(t, 200)
}

func TestSerial200MultiHostIngresses(t *testing.T) {
	ingressType = MULTIHOST
	SerialIngressHelper(t, 200)
}

func TestParallel500InsecureIngresses(t *testing.T) {
	ingressType = INSECURE
	ParallelIngressHelper(t, 500)
}

func TestParallel500SecureIngresses(t *testing.T) {
	ingressType = SECURE
	ParallelIngressHelper(t, 500)
}

func TestParallel500MultiHostIngresses(t *testing.T) {
	ingressType = MULTIHOST
	ParallelIngressHelper(t, 500)
}

func TestSerial500InsecureIngresses(t *testing.T) {
	ingressType = INSECURE
	SerialIngressHelper(t, 500)
}

func TestSerial500SecureIngresses(t *testing.T) {
	ingressType = SECURE
	SerialIngressHelper(t, 500)
}

func TestSerial500MultiHostIngresses(t *testing.T) {
	ingressType = MULTIHOST
	SerialIngressHelper(t, 500)
}
