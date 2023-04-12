## Overview
This document explains IPv6 support in AKO. End to end support is limited to L7 currently (AKO 1.9.3).

### Use cases
AKO supports IPv6 and dual stack. Following table explains what all is supported in AKO.
| Use case   |      Dual stack support      |  comments |
|----------|---------------|-------|
| Frontend VIP |  SUPPORTED | Customers can choose to have either v6only or dual VIPs (v4, v6) for the Virtual services. IPv6 vip for LBSvc is not supported (L4Policy set does not support IPv6 on AVI) |
| Backend (Pod IPs) |    V6ONLY   |   AKO will add either v4 or v6 addresses to the pools based on AKO configuration ('ipFamily'). AKO does not support a mix of v4 and v6 addresses |
| K8s Nodes | V6ONLY |    AKO configures routes to the pod IPs via Node IP. AKO will choose v4 or a v6 IP based on ipFamily attribute in AKO config. For NodePort mode, AKO will choose either v4 or v6 IPs of Nodes based on the ipFamily attribute in AKO config |
| Avi Controller IP | V4ONLY | AKO only supports v4 IPs to talk to Avi controller mgmt interface. AKO pod needs IPv4 connectivity to Avi controller mgmt interface |
| K8s API server | V4ONLY | AKO only supports IPv4 to talk to K8s API server. v6 is not supported |

### Support Matrix
| Case   |      Support      |
|----------|---------------|
| Cloud   |      vCenter      |
| CNI   |      Calico, Antrea      |
| ServiceType   |      ClusterIP, NodePort      |
| Kubernetes   |      Supported      |
| Openshift   |      Not Supported      |
| NodeportLocal   |      Not Supported      |

### Configuration
#### Frontend Support for IPv6
In values.yaml, under NetworkSettings.vipNetworklist user can now specify v6cidr for networks
The field can be set as follows:

    NetworkSettings:
      vipNetworkList:
      - networkName: net1
        cidr: 100.1.1.0/24
        v6cidr: 2002::1234:abcd:ffff:c0a8:101/64

v6cidr is an optional field and can be specified independent of cidr. When v6cidr is specified, AKO will enable auto allocation for IPv6 IPs for VIP. AKO allows VIPs to have both v4 and v6 IPs.

#### Backend Support for IPv6
ipFamily (values: V4, V6; default: V4 ): The ipFamily field in values.yaml determines whether AKO will look for IPv6 or IPv4 IPs for pool servers. to change ipFamily, AKO needs reboot

When ipFamily is V6, AKO looks for IPv6 address for nodes to add to static routes. 

For Calico CNI, AKO will read node ip from internal IP or node annotation projectcalico.org/IPv6Address.

For Antrea CNI, AKO will read node IP from internal IP or node annotation node.antrea.io/transport-addresses, if transportInterface is specified in antrea config.

