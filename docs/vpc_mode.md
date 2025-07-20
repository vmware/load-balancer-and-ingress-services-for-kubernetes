# AKO in VPC Mode for NSX-T Cloud

This feature allows AKO to operate in NSX VPC based environments.

## Overview

NSX Virtual Private Cloud (VPC) is an abstraction to simplify the consumption of NSX-T Data Center networking and security services. It provides a tenant-centric view of the NSX-T Data Center and provisions the subnet, IPAM, and DHCP automatically to simplify the overall consumption and configuration.

When running in VPC mode, AKO will configure VIP networks automatically unlike the non-VPC mode where the VIP networks and their respective IPAM are user-configured. This eliminates the need to configure `networkSettings.vipNetworkList` in `values.yaml`. AKO will always allocate Virtual IPs (VIPs) from the Public IP Address Pool available within the specified VPC.


### Configuration

#### Steps to enable VPC Mode in AKO

1. You need to set the `vpcMode` flag to `true` in `values.yaml`.

```yaml
akoSettings:
  vpcMode: true
```

2. Additionally, you must configure the `NetworkSettings.nsxtT1LR` with the path of the NSX-T VPC in the format `/orgs/<ord-id>/projects/<project-id>/vpcs/<vpc-id>` in `values.yaml`. AKO uses this information to populate the virtualservice's and pool's T1Lr attribute.

```yaml
NetworkSettings:
  nsxtT1LR: "/orgs/<ord-id>/projects/<project-id>/vpcs/<vpc-id>"
```

3. You also need to disable `NetworkSettings.vipNetworkList` in `values.yaml`.

```yaml
NetworkSettings:
  vipNetworkList: []
```