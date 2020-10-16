package scaletest

import (
	"fmt"
	"testing"
	"os"
	"sync"
	"strconv"
	"encoding/json"
	"io/ioutil"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/tests/scaletest/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	"github.com/onsi/gomega"
)
var testbedFileName = "/root/load-balancer-and-ingress-services-for-kubernetes/tests/scaletest/configs/st_testbed.json"
var namespace string
var appName string
var serviceNamePrefix string
var ingressNamePrefix string

var numGoRoutines int
var listOfServicesCreated []string
var ingressesCreated []string
var ingressesDeleted []string
var initialNumOfPG = 0
var ingressType string
var clusterName string
var timeout string
var testCaseTimeOut = "1500s"
var testPollInterval = "10s"
var test *testing.T

func Setup(testbedParams lib.TestbedFields){
	namespace = testbedParams.Test.GetNamespace()
	appName = testbedParams.Test.GetAppName()
	serviceNamePrefix = testbedParams.Test.GetServiceNamePrefix()
	ingressNamePrefix = testbedParams.Test.GetIngressNamePrefix()
	clusterName = testbedParams.AkoParam.Clusters[0].GetClusterName()
	os.Setenv("CTRL_USERNAME",testbedParams.Controller[0].UserName)
	os.Setenv("CTRL_PASSWORD",testbedParams.Controller[0].Password)
	os.Setenv("CTRL_IPADDRESS",testbedParams.Controller[0].Ip)
	lib.CreateApp(appName, namespace)
	listOfServicesCreated = lib.CreateService(serviceNamePrefix, appName, namespace, 2)
}

func Cleanup(){
	lib.DeleteService(listOfServicesCreated, namespace)
	lib.DeleteApp(appName,namespace)
}

func PoolVerification() bool{
	AviClients := utils.SharedAVIClients()
	pools := lib.FetchPools(test, AviClients.AviClient[0])
	if ingressType == "multi-host" && (len(pools) < ((len(ingressesCreated)*2)+initialNumOfPG)){
		return false
	}else if (len(pools) < len(ingressesCreated)+initialNumOfPG){
		return false
	}
	var ingressPoolList []string
	var poolList []string
	for i:=0 ;i<len(ingressesCreated); i++ {
		if ingressType != "multi-host"{
			ingressPoolName := clusterName + "--" + ingressesCreated[i] + ".avi.internal-" + namespace + "-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
		}else{
			ingressPoolName := clusterName + "--" + ingressesCreated[i] + "-host1.avi.internal-" + namespace + "-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
			ingressPoolName = clusterName + "--" + ingressesCreated[i] + "-host2.avi.internal-" + namespace + "-" + ingressesCreated[i]
			ingressPoolList = append(ingressPoolList, ingressPoolName)
		}
	}
	for i:=0;i<len(pools);i++{
		poolList = append(poolList, *pools[i].Name)
	}
	diffMap := map[string]int{}
	for _,ingPool := range ingressPoolList{
		diffMap[ingPool] = 1
	}
	for _, pool := range poolList{
		diffMap[pool] = diffMap[pool] + 1
	}
	var diffNum int
	for _, val := range diffMap{
		if val ==1 {
			diffNum = diffNum + 1
		}
	}
	if diffNum == initialNumOfPG {
		return true
	}
	return false
}

func parallelInsecureIngressCreation(wg *sync.WaitGroup, serviceName string, namespace string, numOfIng int, startIndex int){
	defer wg.Done()
	ingresses := lib.CreateInsecureIngress(ingressNamePrefix, serviceName, namespace, numOfIng, startIndex)
	ingressesCreated = append(ingressesCreated, ingresses...)
}

func parallelSecureIngressCreation(wg *sync.WaitGroup, serviceName string, namespace string, numOfIng int, startIndex int){
	defer wg.Done()
	ingresses := lib.CreateSecureIngress(ingressNamePrefix, serviceName, namespace, numOfIng, startIndex)
	ingressesCreated = append(ingressesCreated, ingresses...)
}

func parallelMultiHostIngressCreation(wg *sync.WaitGroup, serviceName []string, namespace string, numOfIng int, startIndex int){
	defer wg.Done()
	ingresses := lib.CreateMultiHostIngress(ingressNamePrefix, serviceName, namespace, numOfIng, startIndex)
	ingressesCreated = append(ingressesCreated, ingresses...)
}

func parallelIngressDeletion(wg *sync.WaitGroup, namespace string, listOfIngressToDelete []string){
	defer wg.Done()
	ingresses := lib.DeleteIngress(namespace, listOfIngressToDelete)
	ingressesDeleted = append(ingressesDeleted, ingresses...)
}

func ParallelIngressHelper(numOfIng int, test *testing.T){
	ingressesCreated = []string{}
	var blockSize = numOfIng / numGoRoutines
	var remIng =  numOfIng % numGoRoutines
	g := gomega.NewGomegaWithT(test)
	AviClients := utils.SharedAVIClients()
	pools := lib.FetchPools(test, AviClients.AviClient[0])
	initialNumOfPG = len(pools)
	var wg sync.WaitGroup
	nextStartInd := 0
	switch  {
	case ingressType == "insecure":
		test.Logf("Creating %d %s Ingresses Parallely...", numOfIng, ingressType)
		for i := 0; i<numGoRoutines; i++ {
			wg.Add(1)
			if(i+1 <= remIng){
				go parallelInsecureIngressCreation(&wg, listOfServicesCreated[0], namespace, blockSize+1, nextStartInd)
				nextStartInd = nextStartInd + blockSize + 1
			}else{
				go parallelInsecureIngressCreation(&wg, listOfServicesCreated[0], namespace, blockSize, nextStartInd)
				nextStartInd = nextStartInd + blockSize 
			}
		}
	case ingressType == "secure":
		test.Logf("Creating %d %s Ingresses Parallely...", numOfIng, ingressType)
		for i := 0; i<numGoRoutines; i++ {
			wg.Add(1)
			if(i+1 <= remIng){
				go parallelSecureIngressCreation(&wg, listOfServicesCreated[0], namespace, blockSize+1, nextStartInd)
				nextStartInd = nextStartInd + blockSize + 1
			}else{
				go parallelSecureIngressCreation(&wg, listOfServicesCreated[0], namespace, blockSize, nextStartInd)
				nextStartInd = nextStartInd + blockSize 
			}
		}
	case ingressType == "multi-host":
		test.Logf("Creating %d %s Ingresses Parallely...", numOfIng, ingressType)
		for i := 0; i<numGoRoutines; i++ {
			wg.Add(1)
			if((i+1) <= remIng){
				go parallelMultiHostIngressCreation(&wg, listOfServicesCreated, namespace, blockSize+1, nextStartInd)
				nextStartInd = nextStartInd + blockSize + 1
			}else{
				go parallelMultiHostIngressCreation(&wg, listOfServicesCreated, namespace, blockSize, nextStartInd)
				nextStartInd = nextStartInd + blockSize 
			}
		}	
	}
	wg.Wait()
	g.Expect(ingressesCreated).To(gomega.HaveLen(numOfIng))
	test.Logf("Created %d %s Ingresses Parallely", numOfIng, ingressType)
	g.Eventually(PoolVerification,testCaseTimeOut,testPollInterval).Should(gomega.BeTrue())
	test.Logf("Created %d Pools", numOfIng)
	ingressesDeleted = []string{}
	test.Logf("Deleting %d %s Ingresses...", numOfIng, ingressType)
	nextStartInd = 0
	for i := 0; i<numGoRoutines; i++ {
		wg.Add(1)
		if((i+1) <= remIng){
			go parallelIngressDeletion(&wg, namespace, ingressesCreated[nextStartInd : nextStartInd + blockSize + 1])
			nextStartInd = nextStartInd + blockSize + 1
		}else{
			go parallelIngressDeletion(&wg, namespace, ingressesCreated[nextStartInd : nextStartInd + blockSize])
			nextStartInd = nextStartInd + blockSize 
		}
	}
	wg.Wait()
	g.Expect(ingressesDeleted).To(gomega.HaveLen(numOfIng))
	test.Logf("Deleted %d %s Ingresses", numOfIng, ingressType)
	g.Eventually(func() int{
		pools = lib.FetchPools(test, AviClients.AviClient[0])
		return len(pools)
	},testCaseTimeOut,testPollInterval).Should(gomega.Equal(initialNumOfPG))
	test.Logf("Deleted %d Pools", numOfIng)
	test.Logf("%d %s Ingress creation deletion along with verification is done", numOfIng, ingressType)
}

func SerialIngressHelper(numOfIng int, test *testing.T){
	g := gomega.NewGomegaWithT(test)
	AviClients := utils.SharedAVIClients()
	pools := lib.FetchPools(test, AviClients.AviClient[0])
	initialNumOfPG = len(pools)
	switch  {
	case ingressType == "insecure":
		test.Logf("Creating %d %s Ingresses Serially...", numOfIng, ingressType)
		ingressesCreated = lib.CreateInsecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, numOfIng)
	case ingressType == "secure":
		test.Logf("Creating %d %s Ingresses Serially...", numOfIng, ingressType)
		ingressesCreated = lib.CreateSecureIngress(ingressNamePrefix, listOfServicesCreated[0], namespace, numOfIng)
	case ingressType == "multi-host":
		test.Logf("Creating %d %s Ingresses Serially...", numOfIng, ingressType)
		ingressesCreated = lib.CreateMultiHostIngress(ingressNamePrefix, listOfServicesCreated, namespace, numOfIng)
	}
	g.Expect(ingressesCreated).To(gomega.HaveLen(numOfIng))
	test.Logf("Created %d %s Ingresses Serially", numOfIng, ingressType)	
	g.Eventually(PoolVerification,testCaseTimeOut,testPollInterval).Should(gomega.BeTrue())
	test.Logf("Created %d Pools", numOfIng)
	test.Logf("Deleting %d %s Ingresses Serially...", numOfIng, ingressType)
	ingressesDeleted = lib.DeleteIngress(namespace, ingressesCreated)
	g.Expect(ingressesDeleted).To(gomega.HaveLen(numOfIng))
	test.Logf("Deleted %d %s Ingresses Serially", numOfIng, ingressType)
	g.Eventually(func() int{
		pools = lib.FetchPools(test, AviClients.AviClient[0])
		return len(pools)
	},testCaseTimeOut,testPollInterval).Should(gomega.Equal(initialNumOfPG))
	test.Logf("Deleted %d Pools", numOfIng)
	test.Logf("%d %s Ingress serially creation deletion along with verification is done", numOfIng, ingressType)
}

func TestMain(test *testing.M){
	lib.KubeInit()
	var testbedParams lib.TestbedFields 
	testbed, err := os.Open(testbedFileName)
	if err != nil {
		fmt.Println("ERROR : Error opening testbed file ", testbedFileName)
		return
    }
    defer testbed.Close()
    byteValue, _ := ioutil.ReadAll(testbed)
	json.Unmarshal(byteValue, &testbedParams)
	numGoRoutines, err = strconv.Atoi(os.Args[5])
	if err!= nil  {
		numGoRoutines = 5
	}
	if numGoRoutines == 0{
		fmt.Println("ERROR : Number of Go Routines cannot be zero.")
		return
	}
	timeout = os.Args[4]
	Setup(testbedParams)
	test.Run()
	Cleanup()
}

func TestParallel200InsecureIngresses(test *testing.T){
	ingressType = "insecure"
	ParallelIngressHelper(200, test)
}

func TestParallel200SecureIngresses(test *testing.T){
	ingressType = "secure"
	ParallelIngressHelper(200, test)
}

func TestParallel200MultiHostIngresses(test *testing.T){
	ingressType = "multi-host"
	ParallelIngressHelper(200, test)
}

func TestSerial200InsecureIngresses(test *testing.T){
	ingressType = "insecure"
	SerialIngressHelper(200, test)
}

func TestSerial200SecureIngresses(test *testing.T){
	ingressType = "secure"
	SerialIngressHelper(200, test)
}

func TestSerial200MultiHostIngresses(test *testing.T){
	ingressType = "multi-host"
	SerialIngressHelper(200, test)
}

func TestParallel500InsecureIngresses(test *testing.T){
	ingressType = "insecure"
	ParallelIngressHelper(500, test)
}

func TestParallel500SecureIngresses(test *testing.T){
	ingressType = "secure"
	ParallelIngressHelper(500, test)
}

func TestParallel500MultiHostIngresses(test *testing.T){
	ingressType = "multi-host"
	ParallelIngressHelper(500, test)
}

func TestSerial500InsecureIngresses(test *testing.T){
	ingressType = "insecure"
	SerialIngressHelper(500, test)
}

func TestSerial500SecureIngresses(test *testing.T){
	ingressType = "secure"
	SerialIngressHelper(500, test,)
}

func TestSerial500MultiHostIngresses(test *testing.T){
	ingressType = "multi-host"
	SerialIngressHelper(500, test)
}