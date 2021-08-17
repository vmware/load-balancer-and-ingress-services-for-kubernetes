# Upgrade Notes

This document mentions Notes around upgrading to AKO 1.5.1 from any prior AKO releases. These steps are in addition to the `helm upgrade` steps mentioned in the [installation guide](../install/helm.md). It is highly recommended to go through this document before upgrading to AKO 1.5.1

## Refactored VIP Network inputs
AKO 1.5.1 deprecates `subnetIP` and `subnetPrefix` in `values.yaml`  (used during helm installation), and allows specifying the information via the `cidr` field within `vipNetworkList` ([explained here in detail](values.md#networksettingsvipnetworklist)). This is done in favor of providing consistency to the VIP Network related user interfaces exposed via values.yaml / AviInfraSetting CRD / AKO Config (ako-operator).

### helm values.yaml
The following schemas show the changed structure.

  **BEFORE:**
 
        NetworkSettings:
          vipNetworkList:
            - networkName: "net1"
          subnetIP: "10.10.10.0"
          subnetPrefix: "24"
  
  **AFTER:**

        NetworkSettings:
          vipNetworkList:
            - networkName: "net1"
              cidr: "10.10.10.0/24"


### AviInfraSetting CRD
The following schemas show the changed structure of `spec.network` field.

  **BEFORE:**

        spec:
          network:
            names:
              - vip-network-10-10-10-0-24
  **AFTER:**

        spec:
          network:
            vipNetworks:
              - networkName: vip-network-10-10-10-0-24

Existing AviInfraSetting CRDs would need an explicit update after applying the updated CRD schema yamls. CRD schema yamls are applied during the [Helm upgrade](../install/helm.md) step. <br/>

The workflow to update existing AviInfraSettings while performing a `helm upgrade` would be:
0. Make sure that your AviInfraSetting CRDs have the required configurations, if not please save it in yaml files.
1. Follow __Step 1__ in the [helm upgrade](../install/helm.md) guide. This would update the CRD schema yamls that would enable you to provide `vipNetworks` in the new format.
2. Updating the CRD schema would remove the NOW invalid `spec.network` configuration in existing AviInfraSettings. Update the AviInfraSettings to follow the new schema as shown above and apply the changed yamls.
3. Proceed with __Step 2__ of the helm upgrade guide.