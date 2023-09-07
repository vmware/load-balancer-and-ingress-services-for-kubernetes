# Upgrade Notes

This document mentions Notes around upgrading to AKO 1.6.1 from any prior AKO releases. These steps are in addition to the `helm upgrade` steps mentioned in the [installation guide](../install/helm.md). It is highly recommended to go through this document before upgrading to AKO 1.6.1

## Refactored VIP Network inputs
From AKO 1.5.1, fields `subnetIP` and `subnetPrefix` in `values.yaml` have been deprecated (used during helm installation), and allows specifying the information via the `cidr` field within `vipNetworkList` ([explained here in detail](../values.md#networksettingsvipnetworklist)). This is done in favor of providing consistency to the VIP Network related user interfaces exposed via values.yaml / AviInfraSetting CRD / AKO Config (ako-operator).

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
      or
         NetworkSettings:
          vipNetworkList:
            - networkUUID: "dvportgroup-2167-cloud-d4b24fc7-a435-408d-af9f-150229a6fea6f"
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
      or
        spec:
          network:
            vipNetworks:
              - networkUUID: dvportgroup-3167-cloud-d4b24fc7-a435-408d-af9f-150229a6fea6f

Existing AviInfraSetting CRDs would need an explicit update after applying the updated CRD schema yamls. CRD schema yamls are applied during the [helm upgrade](../install/helm.md#upgrade-ako-using-helm) step. <br/>

The workflow to update existing AviInfraSettings while performing a `helm upgrade` would be:
0. Make sure that your AviInfraSetting CRDs have the required configurations, if not please save it in yaml files.
1. Follow __Step 1__ in the helm upgrade guide. This would update the CRD schema yamls that would enable you to provide `vipNetworks` in the new format.
2. Updating the CRD schema would remove the NOW invalid `spec.network` configuration in existing AviInfraSettings. Update the AviInfraSettings to follow the new schema as shown above and apply the changed yamls.
3. Proceed with __Step 2__ of the helm upgrade guide.