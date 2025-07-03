# AKO in VPC Mode for NSX-T Cloud

This feature allows AKO to operate in VPC mode with NSX-T Cloud. 

## Overview

In the VPC mode, data networks will be configured automatically unlike the non-VPC mode where the data networks and their respective IPAM are user-configured. When in VPC mode, NSX provisions the subnet, IPAM, and DHCP automatically to simplify the overall consumption and configuration.

Because NSX-T manages the data networks in VPC Mode, you do not need to specify `networkSettings.vipNetworkList` in `values.yaml`. AKO will automatically place Virtual IPs (VIPs) on the public subnet available within the specified VPC.


### Configuration

To enable VPC Mode, you would need to enable it in `values.yaml`. Set the `vpcMode` flag to `true`.

```yaml
akoSettings:
  vpcMode: true
```

Additionally, you must configure the `NetworkSettings.nsxtT1LR` with the path of the NSX-T VPC in the format `/orgs/<ord-id>/projects/<project-id>/vpcs/<vpc-id>` in `values.yaml`. AKO uses this information to populate the virtualservice's and pool's T1Lr attribute.

```yaml
NetworkSettings:
  nsxtT1LR: "/orgs/<ord-id>/projects/<project-id>/vpcs/<vpc-id>"
```

Optionally, you may disable `NetworkSettings.vipNetworkList` in `values.yaml`.

```yaml
NetworkSettings:
  vipNetworkList: []
```