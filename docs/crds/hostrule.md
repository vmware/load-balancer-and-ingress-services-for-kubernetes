### HostRule 

HostRule CRD is primarily targeted to be used by the Operator. This CRD can be used to express additional virtual host
properties. The virtual host FQDN is matched from either Kubernetes Ingress or OpenShift Route based objects. 

***Note***
With AKO 1.11.1, HostRule is transitioned to v1beta1 version. There are no schema changes between version v1alpha1 and v1beta1. AKO 1.11.1 supports both v1alpha1 and v1beta1 but recommendation is to create new CRD objects in v1beta1 version and transition existing objects to v1beta1 version. AKO will deprecate v1alpha1 version in future releases.

A sample HostRule CRD looks like this:

    apiVersion: ako.vmware.com/v1beta1
    kind: HostRule
    metadata:
      name: my-host-rule
      namespace: red
    spec:
      virtualhost:
        fqdn: foo.region1.com # mandatory
        fqdnType: Exact
        enableVirtualHost: true
        tls: # optional
          sslKeyCertificate:
            name: avi-ssl-key-cert
            type: ref
            alternateCertificate:
              name: avi-ssl-key-cert2
              type: ref
          sslProfile: avi-ssl-profile
          termination: edge
        gslb:
          fqdn: foo.com
          includeAliases: false
        httpPolicy: 
          policySets:
          - avi-secure-policy-ref
          overwrite: false
        datascripts:
        - avi-datascript-redirect-app1
        wafPolicy: avi-waf-policy
        applicationProfile: avi-app-ref
        networkSecurityPolicy: avi-network-security-policy-ref
        icapProfile: 
        - avi-icap-ref
        analyticsProfile: avi-analytics-ref
        errorPageProfile: avi-errorpage-ref
        analyticsPolicy: # optional
          fullClientLogs:
            enabled: true
            throttle: HIGH
          logAllHeaders: true
        tcpSettings:
          listeners:
          - port: 8081
          - port: 6443
            enableSSL: true
          loadBalancerIP: 10.10.10.1
        aliases: # optional
        -  bar.com
        -  baz.com
        l7Rule: my-l7-rule-name


### Specific usage of HostRule CRD

HostRule CRD can be created in a given namespace where the operator desires to have more control. 
The section below walks over the details and associated rules of using each field of the HostRule CRD.

#### HostRule to VS matching using fqdn/fqdnType

A given HostRule is applied to a virtualservice if the VS hosts the `fqdn` mentioned in the HostRule CRD. This `fdqn` must exactly match with the one the virtualservice is hosting. However, in order to simplify the user experience, and provide for an easy way to apply a HostRule to individual or multiple virtualservices, the `fqdnType` field can be employed, that provides greated flexibility when it comes to specifying the string in the `fqdn` field.

One of 3 following matching criterias can be specified with `fqdnType`
- `Exact`: Matches the string character by character to the VS FQDNs, in an exact match fashion.
- `Wildcard`: Matches the string to multiple VS FQDNs, and matches the FQDNs with the provided string as the suffix. The string must start with a '*' to qualify for wildcard matching.

        fqdn: *.alb.vmware.com
        fqdnType: Wildcard

- `Contains`: Matches the string to multiple VS FQDNs, and matches the FQDNs with the provided string as a substring of any possible FQDNs programmed by AKO.

	      fqdn: Shared-VS-L7-1
        fqdnType: Contains

The `fqdnType` field defaults to `Exact`.

#### Enable/Disable Virtual Host

HostRule CRD can be used to enable/disable corresponding virtual services created by AKO on Avi. This removes any virtual host related configuration from the data plane (Avi service engines) in addition to disabling traffic on the virtual host/fqdn.

        enableVirtualHost: false

This property can be applied only for secure FQDNs and cannot be applied for insecure routes. The default value is `true`.

#### Express HTTP policy object refs.

HostRule CRD can be used to express httppolicyset references. These httppolicyset objects should be pre-created in the Avi controller.

        httpPolicy: 
          policySets:
          - avi-secure-policy-ref
          overwrite: false

The httppolicyset currently is only applicable for secure FQDNs and cannot be applied for insecure routes.
The order of evaluation of the httpolicyset rules is in the same order they appear in the CRD definition. The list of httpolicyset rules are
always intepreted as an `AND` operation.

AKO currently uses httppolicyset objects on the SNI virtualservices to route traffic based on host/path matches. These rules are always at
a lower index than the httppolicyset objects specified in the CRD object. In case, a user would want to overwrite all httppolicyset objects
on a SNI virtualservice with the ones specified in the HostRule CRD, the `overwrite` flag can be set to `true`. The default value  for  `overwrite` is `false`.


#### Express WAF policy object refs.

HostRule CRD can be used to express WAF policy references. The WAF policy object should have been created in the Avi Controller prior to this
CRD creation.

        wafPolicy: avi-waf-policy

 This property can be applied only for secure FQDNs and cannot be applied for insecure routes.
 WAF policies are useful when deep layer 7 packet filtering is required.
 
#### Express custom application profiles

HostRule CRD can be used to express application profile references. The application profile reference should have been created in the Avi Controller 
prior to this CRD creation. The application profile should be of `TYPE` of `APPLICATION_PROFILE_TYPE_HTTP`.

        applicationProfile: avi-app-ref
 
 This property can be applied only for secure FQDNs and cannot be applied for insecure routes.
 The application profiles can be used for various HTTP/HTTP2 protocol settings.

#### Express custom ICAP profile

HostRule CRD can be used to express a single ICAP profile reference per host. The ICAP profile reference should have been created in the Avi Controller prior to this CRD creation.

        icapProfile: 
        - avi-icap-ref
 
 This property can be applied for both secure and insecure hosts via EVH parent and child Virtual Services, SNI child Virtual Services and dedicated VS's.
 The [ICAP profile](https://avinetworks.com/docs/22.1/icap/) can be used for transporting HTTP traffic to 3rd party services for processes such as content sanitization and antivirus scanning.

#### Express custom analytics profiles

HostRule CRD can be used to express analytics profile references. The analytics profile reference should have been created in the Avi Controller prior to this CRD creation.

        analyticsProfile: avi-analytics-ref

 This property can be applied only for secure FQDNs and cannot be applied for insecure routes. The analytics profiles can be used for various Network/HTTP/Healthscore analytics settings, log processing etc.


#### Express custom error page profiles

HostRule CRD can be used to express error page profile references. The error page profile reference should have been created in the Avi Controller prior to this CRD creation.

        errorPageProfile: avi-errorpage-ref

 This property can be applied only for secure FQDNs and cannot be applied for insecure routes. The error page profiles can be used to send a custom error page to the client generated by the proxy.


#### Express datascripts

HostRule CRD can be used to express error datascript references. The datascript references should have been created in the Avi Controller prior to this CRD creation.

        datascripts:
        - avi-datascript-redirect-app1

This property can be applied only for secure FQDNs and cannot be applied for insecure routes. The datascripts can be used to apply custom scripts to data traffic. The order of evaluation of the datascripts is in the same order they appear in the CRD definition.


#### Express TLS configuration

If the kubernetes operator wants to control the TLS termination from a privileged namespace then the HostRule CRD can be created in such a namespace.

        tls:
          sslKeyCertificate:
            name: avi-ssl-key-cert
            type: ref
            alternateCertificate:
              name: avi-ssl-key-cert2
              type: ref
          sslProfile: avi-ssl-profile
          termination: edge

The `name` field refers to an Avi object if `type` specifies the value as `ref`. Alternatively, we also support a kubernetes
`Secret` to be specified where the sslkeyandcertificate object is created by AKO using the Secret. 

        tls:
          sslKeyCertificate:
            name: k8s-app-secret
            type: secret
          termination: edge

An `alternateCertificate` option is provided in case the application needs to be configured to provide multiple server certificates, typically when trying to configure both RSA and ECC signed certificates. Avi Controller allows a Virtual Service to be configured with two certificates at a time, one each of RSA and ECC. This enables Avi Controller to negotiate the optimal algorithm or cipher with the client. If the client supports ECC, in that case the ECC algorithm is preferred, and RSA is used as a fallback in cases where the clients do not support ECC.

`sslProfile`, additionally, can be used to determine the set of SSL versions and ciphers to accept for SSL/TLS terminated connections. If the `sslProfile` is not defined, AKO defaults to the sslProfile `System-Standard-PFS` defined in Avi.

Currently only one of type of termination is supported viz. `edge`. In the future, we should be able to support other types of termination policies.

#### Configure GSLB FQDN

A GSLB FQDN can be specified within the HostRule CRD. This is only used if AKO is used with AMKO and not otherwise.

        gslb:
          fqdn: foo.com
          includeAliases: false

This additional FQDN inherits all the properties of the root FQDN specified under the the `virtualHost` section.
Use this flag if you would want traffic with a GSLB FQDN to get routed to a site local FQDN. For example, in the above CRD, the client request from a GSLB
DNS will arrive with the host header as foo.com to the VIP hosting foo.region1.com in region1. This CRD property would ensure that the request is routed appropriately to the backend service of `foo.region1.com`

This knob is currently only supported with the SNI model and not with Enhanced Virtual Hosting model.

The `includeAliases` is used by AMKO. Whenever a GSLB FQDN is provided and the `useCustomGlobalFqdn` is set to true in AMKO, a GSLB Service is created for the GSLB FQDN instead of the local FQDN(hostname). [Refer this](https://github.com/vmware/global-load-balancing-services-for-kubernetes/blob/master/docs/local_and_global_fqdn.md)

When this flag is set to `false` the Domain Name of the GSLB Service is set to the GSLB FQDN. 

When this flag is set to `true` in addition to the GSLB FQDN, AMKO adds the FQDNs mentioned under [aliases](#aliases) to domain names of the GSLB Service. 

#### Configure Analytics Policy

The HostRule CRD can be used to configure analytics policies such as enable/disable non-significant logs, throttle the number of non-significant logs per second on each SE, enable/disable logging of all headers, etc.

        analyticsPolicy:
          fullClientLogs:
            enabled: true
            throttle: HIGH
          logAllHeaders: true

The `throttle` will be in effect only when `enabled` is set to `true`. The possible values of `throttle` are DISABLED (0), LOW (50), MEDIUM (30) and HIGH (10).

The AKO sets the duration of logging the non-significant logs to infinity by default. It is the responsibility of the user to disable the non-significant logs when it is no longer required.

#### Configure TCP Settings

The TCP Settings section is responsible for configuring Parent virtualservice specific parameters using the HostRule CRD. 
The `tcpSettings` block, in addition to any other parameters provided in the HostRule, is only applied to Parent VSes and dedicated VSes. The `tcpSettings` block does not have any effect on child VSes.

In order to consume TCP setting configurations for parent VSes, the HostRule must be matched to a Shared/Dedicated VS FQDN, using the existing `fqdn` field in HostRule. 
Where dedicated VSes are created corresponding to a single application, Shared VSes would host multiple application FQDNs. Therefore, in order to apply a HostRule to a dedicated VS, users can simply provide the application FQDN in the HostRule `fqdn` field. For Shared VSes however, users can either provide the AKO programmed Shared VS FQDN (TODO: Provide link), or utilize the `fqdnType: Contains` parameter with the Shared VS name itself.

        fqdn: foo.com     # dedicated VS
        fqdnType: Exact
        tcpSettings:
          listeners:
          - port: 6443
            enableSSL: true


        fqdn: Shared-VS-L7-1.admin.avi.com    # AKO configured Shared VS fqdn
        fqdnType: Exact
        tcpSettings:
          loadBalancerIP: 10.10.10.1


        fqdn: Shared-VS-L7-1      # bound for clusterName--Shared-VS-L7-1
        fqdnType: Contains
        tcpSettings:
          loadBalancerIP: 10.10.10.1

##### Custom Ports

In order to overwrite the ports opened for VSes created by AKO, users can provide the port details under the `listeners` setting. The ports mentioned under this section overwrites the default open ports, 80 and 443 (SSL enabled). This is applicable only for Shared or Dedicated virtual services.

        tcpSettings:
          listeners:
          - port: 80
          - port: 8081
          - port: 6443
            enableSSL: true


**Note**: It is required that one of the ports that are mentioned in the setting has `enableSSL` field set to `true`.

##### L7 Static IP

The `loadBalancerIP` field can be used to provide a valid preferred IPv4 address for L7 virtual services created for the Shared or Dedicated VS. The preferred IP must be part of the IPAM configured for the Cloud, and must not overlap with any other IP addresses already in use. In case of any misconfigurations whatsoever, AKO would fail to configure the virtual service appropriately throwing an ERROR log for the same.

        tcpSettings:
          loadBalancerIP: 10.10.10.199

**Note**: The HostRule CRD is not aware of the misconfigurations while it is being created, therefore the HostRule will be `Accepted` nonetheless.

#### L7Rule 

L7rule field can be used to specify the name of [L7Rule](./l7rule.md) CRD. It is used to modify select VS Properties which are not part of HostRule CRD.

**Note**: This property is available only in HostRule `v1beta1` schema definition.

#### <a id="aliases"> Configure aliases for FQDN

The Aliases field adds the ability to have multiple FQDNs configured under a specific route/ingress for the child VS instead of creating the route/ingress multiple times.

        aliases:
        - bar.com
        - baz.com

This list of FQDNs inherits all the properties of the root FQDN specified under the `virtualHost` section.
Traffic would arrive with the host header as bar.com to the VIP hosting foo.region1.com and this CRD property would ensure that the request is routed appropriately to the backend service of `foo.region1.com`.

Aliases field must contain unique FQDNs and must not contain GSLB FQDN or the root FQDN. Users must ensure that the `fqdnType` is set as `Exact` before setting this field.

#### Express custom network security policy object ref
HostRule CRD can be used to express network security policy object references. The network security policy object should have been created in the Avi Controller prior to this CRD creation.
The `networkSecurityPolicy` setting, in addition to any other parameters provided in the HostRule, is only applied to Parent VSes and dedicated VSes. The `networkSecurityPolicy` setting does not have any effect on child VSes.

         networkSecurityPolicy: avi-network-security-policy-ref

***Note***
1. This property is available only in HostRule `v1beta1` schema definition.
2. The HostRule CRD is not aware of the misconfigurations if it is applied to Child VS while it is being created, therefore the HostRule will be `Accepted` nonetheless. AKO will print warning message regarding this.

#### Status Messages

The status messages are used to give instantaneous feedback to the users about the reference objects specified in the HostRule CRD.

Following are some of the sample status messages:

##### Accepted HostRule object

    $ kubectl get hr
    NAME                 HOST                  STATUS     AGE
    secure-waf-policy    foo.avi.internal      Accepted   3d3h

A HostRule is accepted only when all the reference objects specified inside it exist in the Avi Controller.

##### A Rejected HostRule object

    $ kubectl get hr
    NAME                     HOST                  STATUS     AGE
    secure-waf-policy-alt    foo.avi.internal      Rejected   2d23h
    
The detailed rejection reason can be obtained from the status:

    status:
    error: duplicate fqdn foo.avi.internal found in default/secure-waf-policy-alt
    status: Rejected
    
#### Conditions and Caveats

##### Converting insecure FQDNs to secure

The HostRule CRD can be used to convert an insecure host fqdn to a secure one. This is done by specifying a `tls` section in the CRD object.
Whatever `sslKeyCertificate` is provided for the FQDN, will override all sslkeyandcertificates generated for the FQDN. This maybe useful if:

* The operator wants to convert an insecure ingress FQDN to secure.

* The operator wants to override any existing secrets for a given host fqdn and define tls termination semantics. 

##### Certificate precedence

If the ingress object specifies a Secret for SNI termination and the HostRule CRD also specifies a sslKeyCertificate for the same `virtualhost` then the
sslkeycertificate in the HostRule CRD will take precedence over the Secret object associated with the Ingress.

##### HostRule deletion

If a HostRule is deleted, all the settings for the FQDNs are withdrawn from the Avi controller.

##### HostRule admission

A HostRule CRD is only admitted if all the objects referenced in it, exist in the Avi Controller. If after admission the object references are
deleted out-of-band, then AKO does not re-validate the associated HostRule CRD objects. The user needs to manually edit or delete the object
for new changes to take effect.

##### Duplicate FQDN rules

Two HostRule CRDs cannot be used for the same FQDN information across namespaces. If AKO finds a duplicate FQDN in more than one HostRules, AKO honors the first HostRule that gets created and rejects the others. In case of AKO reboots, the CRD that gets honored might not be the same as the one honored earlier.
