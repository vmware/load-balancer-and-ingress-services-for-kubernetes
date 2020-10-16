package lib

type AkoVCenterConfiguration struct {
	VcenterUser     string `json:"user"`
	VcenterPassword string `json:"password"`
	VcenterURL      string `json:"vcenter_url"`
}

func (vcenterconf AkoVCenterConfiguration) GetVCenterUserName() string {
	return vcenterconf.VcenterUser
}

func (vcenterconf AkoVCenterConfiguration) GetVCenterPassword() string {
	return vcenterconf.VcenterPassword
}

func (vcenterconf AkoVCenterConfiguration) GetVCenterURL() string {
	return vcenterconf.VcenterURL
}

type Platform struct {
	PlatformType            string                  `json:"type"`
	AkoVCenterConfiguration AkoVCenterConfiguration `json:"vcenter_configuration"`
}

func (platform Platform) GetPlatformType() string {
	return platform.PlatformType
}
func (platform Platform) GetVCenterConfiguration() AkoVCenterConfiguration {
	return platform.AkoVCenterConfiguration
}

type Cluster struct {
	ClusterId              string   `json:"cluster_id"`
	ClusterName 		   string 	`json:"cluster_name"`
	KubeConfigFilePath     string   `json:"kubeconfig_file"`
	CniPlugin              string   `json:"cniPlugin"`
	CloudName              string   `json:"cloudName"`
	DisableStaticRouteSync string   `json:"disableStaticRouteSync"`
	DefaultIngController   string   `json:"defaultIngController"`
	NetworkName            string   `json:"NetworkName"`
	VrfRefName             string   `json:"vrfRefName"`
	Platform               Platform `json:"platform"`
}

func (cluster Cluster) GetPlatformDetails() Platform {
	return cluster.Platform
}

func (cluster Cluster) GetClusterName() string {
	return cluster.ClusterName
}

func (cluster Cluster) GetNetworkName() string {
	return cluster.NetworkName
}

func (cluster Cluster) GetClusterId() string {
	return cluster.ClusterId
}

func (cluster Cluster) GetKubeConfigFilePath() string {
	return cluster.KubeConfigFilePath
}

func (cluster Cluster) GetCniPlugin() string {
	return cluster.CniPlugin
}

func (cluster Cluster) GetCloudName() string {
	return cluster.CloudName
}

func (cluster Cluster) GetDisableStaticRouteSync() string {
	return cluster.DisableStaticRouteSync
}

func (cluster Cluster) GetDefaultIngController() string {
	return cluster.DefaultIngController
}

func (cluster Cluster) GetVrfRefName() string {
	return cluster.VrfRefName
}

type AkoParams struct {
	NumClusters int       `json:"num_clusters"`
	Clusters    []Cluster `json:"clusters"`
}

func (akoParams AkoParams) GetNumberOfClusters() int {
	return akoParams.NumClusters
}

func (akoParams AkoParams) GetClusterList() []Cluster {
	return akoParams.Clusters
}

func (akoParams AkoParams) GetCluster(clusterNum int) Cluster {
	return akoParams.Clusters[clusterNum]
}

type Test struct{
	Namespace 		  string `json:"namespace"`
	AppName 		  string `json:"appName"`
	ServiceNamePrefix string `json:"serviceNamePrefix"`
	IngressNamePrefix string `json:"ingressNamePrefix"`
}

func (test Test) GetNamespace() string{
	return test.Namespace
}

func (test Test) GetAppName() string{
	return test.AppName
}

func (test Test) GetServiceNamePrefix() string{
	return test.ServiceNamePrefix
}

func (test Test) GetIngressNamePrefix() string{
	return test.IngressNamePrefix
}

type VMNetwork struct {
	Management string `json:"mgmt"`
}

func (vmNet VMNetwork) GetNetworkManagement() string {
	return vmNet.Management
}

type Controller struct {
	DataCenter string    `json:"datacenter"`
	Name       string    `json:"name"`
	Cluster    string    `json:"cluster"`
	Ip         string    `json:"ip"`
	UserName   string 	 `json:"username"`
	Password   string 	 `json:"password"`
	Mask       string    `json:"mask"`
	Network    VMNetwork `json:"networks"`
	CloudName  string    `json:"cloud_name"`
	Host       string    `json:"host"`
	Static     string    `json:"static"`
	Datastore  string    `json:"datastore"`
	Type       string    `json:"type"`
	Gateway    string    `json:"gateway"`
	CpuCores   string    `json:"cpu_cores"`
	PublicIp   string    `json:"public_ip"`
}

func (ctlr Controller) GetVMDataCenter() string {
	return ctlr.DataCenter
}

func (ctlr Controller) GetVMName() string {
	return ctlr.Name
}

func (ctlr Controller) GetVMCluster() string {
	return ctlr.Cluster
}

func (ctlr Controller) GetVMIP() string {
	return ctlr.Ip
}

func (ctlr Controller) GetUserName() string{
	return ctlr.UserName
}

func (ctlr Controller) GetPassword() string{
	return ctlr.Password
}

func (ctlr Controller) GetVMMask() string {
	return ctlr.Mask
}

func (ctlr Controller) GetVMNetwork() VMNetwork {
	return ctlr.Network
}

func (ctlr Controller) GetVMCloudName() string {
	return ctlr.CloudName
}

func (ctlr Controller) GetVMHost() string {
	return ctlr.Host
}

func (ctlr Controller) GetVMStatic() string {
	return ctlr.Static
}

func (ctlr Controller) GetVMDataStore() string {
	return ctlr.Datastore
}

func (ctlr Controller) GetVMType() string {
	return ctlr.Type
}

func (ctlr Controller) GetVMGateway() string {
	return ctlr.Gateway
}

func (ctlr Controller) GetVMCpuCores() string {
	return ctlr.CpuCores
}

func (ctlr Controller) GetVMPulbicIp() string {
	return ctlr.PublicIp
}

type TestbedFields struct {
	AkoParam       AkoParams      `json:"Ako_params"`
	Test		   Test           `json:"Test"`
	Controller     []Controller   `json:"Controller"`
}

func (testbed TestbedFields) GetAkoParam() AkoParams {
	return testbed.AkoParam
}


