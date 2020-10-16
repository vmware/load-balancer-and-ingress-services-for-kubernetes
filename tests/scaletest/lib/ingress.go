package lib

import (
	"flag"
	"path/filepath"
	"strconv"
	"testing"
	coreV1 "k8s.io/api/core/v1"
	appsV1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

var ingressResource = schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "ingresses"}
var kubeClient dynamic.Interface
var coreV1Client corev1.CoreV1Interface
var appsV1Client appsv1.AppsV1Interface

func CreateApp(appName string, namespace string){
	deploymentSpec := &appsV1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name: appName,
			Namespace: namespace,
		},
		Spec: appsV1.DeploymentSpec{
			ProgressDeadlineSeconds: func() *int32 { i := int32(600); return &i }(),
			RevisionHistoryLimit: func() *int32 { i := int32(10); return &i }(),
			Replicas: func() *int32 { i := int32(2); return &i }(),
			Selector: &metaV1.LabelSelector{
				MatchLabels: map[string]string{
					"app": appName,
				},
			},
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Labels: map[string]string{
						"app": appName,
					},
				},
				Spec: coreV1.PodSpec{
					Containers: []coreV1.Container{
						{
							Name:  appName,
							Image: "avinetworks/server-os",
							ImagePullPolicy: func() coreV1.PullPolicy { str  := coreV1.PullPolicy("IfNotPresent"); return str}(),
							Ports: []coreV1.ContainerPort{
								{
									Name:          "http",
									Protocol:      coreV1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
							TerminationMessagePath: "/dev/termination-log",
							TerminationMessagePolicy: func() coreV1.TerminationMessagePolicy { str := coreV1.TerminationMessagePolicy("File"); return str}(),
						},
					},
					DNSPolicy: func() coreV1.DNSPolicy{ str := coreV1.DNSPolicy("ClusterFirst"); return str}(),
					RestartPolicy: func() coreV1.RestartPolicy { str := coreV1.RestartPolicy("Always"); return str}(),
					SchedulerName: "default-scheduler",
					TerminationGracePeriodSeconds: func() *int64 { i := int64(30); return &i }(),
				},
			},
		},
	}

	_, err := appsV1Client.Deployments(namespace).Create(deploymentSpec)
	if err != nil {
		panic(err)
	}
}

func DeleteApp(appName string, namespace string){
	err := appsV1Client.Deployments(namespace).Delete(appName, &metaV1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
}

func CreateService(serviceNamePrefix string, appName string, namespace string, num int) []string{
	var listOfServicesCreated []string
	for i := 1; i<=num; i++{
		serviceName := serviceNamePrefix + strconv.Itoa(i)
		serviceSpec := &coreV1.Service{
			ObjectMeta: metaV1.ObjectMeta{
				Name: serviceName,
				Namespace: namespace,
			},
			Spec: coreV1.ServiceSpec{
				Selector: map[string]string{
					"app": appName,
				},
				Ports: []coreV1.ServicePort{
					{
						Port: 8080,
						TargetPort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 8080,
						},
					},
				},
			},
		}
		_, err := coreV1Client.Services(namespace).Create(serviceSpec)
		if err != nil {
			panic(err)
		}
		listOfServicesCreated = append(listOfServicesCreated, serviceName)
	}
	return listOfServicesCreated
}

func DeleteService(serviceNameList []string, namespace string){
	for i:=0;i<len(serviceNameList);i++ {
		err := coreV1Client.Services(namespace).Delete(serviceNameList[i], &metaV1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
	} 
}

func CreateInsecureIngress(ingressNamePrefix string, serviceName string, namespace string, num int, startIndex ...int) []string{
	var listOfIngressCreated []string
	var startInd int
	if(len(startIndex)==0){
		startInd = 0
	} else{
		startInd = startIndex[0]
	}
	for i := startInd; i < num + startInd; i++{ 
		ingressName := ingressNamePrefix + strconv.Itoa(i)
		ingress := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"name": ingressName,
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"rules": []map[string]interface{}{
						{
							"host" : ingressName + ".avi.internal",
							"http" : map[string]interface{}{
								"paths" : []map[string]interface{}{
									{
										"backend" : map[string]interface{}{
											"serviceName": serviceName,
											"servicePort": 8080,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		_, err := kubeClient.Resource(ingressResource).Namespace(namespace).Create(ingress, metaV1.CreateOptions{})
		if err != nil {
			panic(err)
		}
		listOfIngressCreated = append(listOfIngressCreated, ingressName) 

	}
	return listOfIngressCreated
}

func CreateSecureIngress(ingressNamePrefix string, serviceName string, namespace string, num int, startIndex ...int) []string{
	var listOfIngressCreated []string
	var startInd int
	if(len(startIndex)==0){
		startInd = 0
	} else{
		startInd = startIndex[0]
	}
	for i := startInd; i < num + startInd; i++{ 
		ingressName := ingressNamePrefix + strconv.Itoa(i)
		ingress := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"name": ingressName,
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"tls": []map[string]interface{}{
						{
							"secretName": "ingress-host-tls",
							"hosts": []interface{}{
								"secure-ingress.avi.internal",
							},
						},
					},
					"rules": []map[string]interface{}{
						{
							"host" : ingressName + ".avi.internal",
							"http" : map[string]interface{}{
								"paths" : []map[string]interface{}{
									{
										"backend" : map[string]interface{}{
											"serviceName": serviceName,
											"servicePort": 8080,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		_, err := kubeClient.Resource(ingressResource).Namespace(namespace).Create(ingress, metaV1.CreateOptions{})
		if err != nil {
			panic(err)
		}
		listOfIngressCreated = append(listOfIngressCreated, ingressName) 

	}
	return listOfIngressCreated
}

func CreateMultiHostIngress(ingressNamePrefix string, listOfServices []string, namespace string, num int, startIndex ...int) []string{
	var listOfIngressCreated []string
	var startInd int
	if(len(startIndex)==0){
		startInd = 0
	} else{
		startInd = startIndex[0]
	}
	for i := startInd; i < num + startInd; i++{ 
		ingressName := ingressNamePrefix + "-multi-host-" + strconv.Itoa(i)
		ingress := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"name": ingressName,
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"tls": []map[string]interface{}{
						{
							"secretName": "ingress-host-tls",
							"hosts": []interface{}{
								"secure-ingress.avi.internal",
							},
						},
					},
					"rules": []map[string]interface{}{
						{
							"host" : ingressName + "-host1.avi.internal",
							"http" : map[string]interface{}{
								"paths" : []map[string]interface{}{
									{
										"backend" : map[string]interface{}{
											"serviceName": listOfServices[0],
											"servicePort": 8080,
										},
									},
								},
							},
						},
						{
							"host" : ingressName + "-host2.avi.internal",
							"http" : map[string]interface{}{
								"paths" : []map[string]interface{}{
									{
										"backend" : map[string]interface{}{
											"serviceName": listOfServices[1],
											"servicePort": 8080,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		_, err := kubeClient.Resource(ingressResource).Namespace(namespace).Create(ingress, metaV1.CreateOptions{})
		if err != nil {
			panic(err)
		}
		listOfIngressCreated = append(listOfIngressCreated, ingressName) 

	}
	return listOfIngressCreated
}

func DeleteIngress(namespace string, listOfIngressToDelete []string) []string{
	var listOfDeletedIngresses []string
	for i:=0; i<len(listOfIngressToDelete);i++{
		ingressName := listOfIngressToDelete[i]
		deletePolicy := metaV1.DeletePropagationForeground
		deleteOptions := &metaV1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}
		if err := kubeClient.Resource(ingressResource).Namespace(namespace).Delete(ingressName, deleteOptions); err != nil {
			panic(err)
		}
		listOfDeletedIngresses = append(listOfDeletedIngresses, ingressName) 
	}
	return listOfDeletedIngresses
}

func ListIngress( t *testing.T, namespace string){
	t.Logf("Listing ingress in namespace %q:\n", coreV1.NamespaceDefault)
	list, err := kubeClient.Resource(ingressResource).Namespace(namespace).List(metaV1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		t.Logf(" * %s \n", d.GetName())
	}
}

func KubeInit() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	kubeClient, err = dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	coreV1Client = clientset.CoreV1()
	appsV1Client = clientset.AppsV1()
}