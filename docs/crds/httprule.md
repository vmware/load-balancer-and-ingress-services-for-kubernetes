### HTTPRule

The HTTPRule CRD is primarily targetted for the developers. While the path matching rules in the Ingress/Route objects would define
traffic routing rules to the microservices, the HTTPRule CRD can be used as a complimentary object to control additional layer 7
properties like: algorithm, hash, tls re-encrypt use cases.

A sample HTTPRule object looks like this:

    apiVersion: ako.vmware.com/v1alpha1
    kind: HTTPRule
    metadata:
       name: my-http-rule
       namespace: purple-l7
    spec:
      fqdn: foo.avi.internal
      paths:
      - target: /foo
        applicationPersistence: cookie-userid-persistence
        healthMonitors:
        - my-health-monitor-1
        - my-health-monitor-2
        loadBalancerPolicy:
          algorithm: LB_ALGORITHM_CONSISTENT_HASH
          hash: LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS
        tls: ## This is a re-encrypt to pool
          type: reencrypt # Mandatory [re-encrypt]
          sslProfile: avi-ssl-profile
          destinationCA:  |-
            -----BEGIN CERTIFICATE-----
            [...]
            -----END CERTIFICATE-----

__NOTE__ : The HTTPRule only applies to paths in the Ingress/Route objects which are specified in the same namespace as the HTTPRule CRD.

### Specific usage of the HTTPRule CRD

The HTTPRule CRD does not have any Avi specific semantics. Hence the developers are free to express their preferences using this CRD
without any knowledge of the Avi objects. Each HTTPRule CRD must be bound to a FQDN (both secure or insecure) to subscribe to rules for a specific hostpath combinations.

#### Express loadbalancer alogrithm

The loadbalancer policies are a predefined set of values which the user can choose from. Presently the following values are supported for
loadbalancer policy:

      - LB_ALGORITHM_CONSISTENT_HASH
      - LB_ALGORITHM_CORE_AFFINITY
      - LB_ALGORITHM_FASTEST_RESPONSE
      - LB_ALGORITHM_FEWEST_SERVERS
      - LB_ALGORITHM_LEAST_CONNECTIONS
      - LB_ALGORITHM_LEAST_LOAD
      - LB_ALGORITHM_ROUND_ROBIN

The way one could configure the loadbalancer policy for a given ingress path is as follows:

      - target: /foo 
        loadBalancerPolicy:
          algorithm: LB_ALGORITHM_FEWEST_SERVERS
          
This rule is applied all paths matching `/foo` and subsets of `/foo/xxx`

More about pool algorithm can be found [here](https://avinetworks.com/docs/18.1/load-balancing-algorithms/).

The `hash` field is used when the algorithm is chosen as `LB_ALGORITHM_CONSISTENT_HASH`. Otherwise it's not applicable. 
Similarly a `hostHeader` field is used only when the `hash` is chosen as `LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER`.

A sample setting with these fields would look like this:

      - target: /foo 
        loadBalancerPolicy:
          algorithm: LB_ALGORITHM_CONSISTENT_HASH
          hash: LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER
          hostHeader: foo
 
If the `hostHeader` is specified in any other case it's ignored.
If the algorithm isn't `LB_ALGORITHM_CONSISTENT_HASH` then the `hash` field is ignored.

#### Express application persistence profile
HTTPRule CRD can be used to express application persistence profile references. The application persistence profile reference should have been created in the Avi Controller prior to this CRD creation.

      applicationPersistence: cookie-userid-persistence

The application persistence profile can be used to maintain stickyness to a server instance based on cookie values, headers etc. for a desired duration of time.

#### Express health monitors
HTTPRule CRD can be used to express health monitor references. The health monitor reference should have been created in the Avi Controller prior to this CRD creation.

      healthMonitors:
      - my-health-monitor-1
      - my-health-monitor-2

The health monitors can be used to verify server health. A server (kubernetes pods in this case) will be marked UP only when all the health monitors return successful responses. Health monitors provided here overwrite the default health monitor configuration set by AKO i.e. `System-TCP` for HTTP/TCP traffic and `System-UDP` for UDP traffic based on the ingress/service configuration.

#### Reencrypt traffic to the services

While AKO can terminate TLS traffic, it also provides and option where the users can choose to re-encrypt the traffic between the Avi SE and the
backend application server. The following option is provided for `reencrypt`:

        tls: ## This is a re-encrypt to pool
          type: reencrypt # Mandatory [re-encrypt]
          sslProfile: avi-ssl-profile
          destinationCA:  |-
            -----BEGIN CERTIFICATE-----
            [...]
            -----END CERTIFICATE-----
          
`sslProfile`, additionally, can be used to determine the set of SSL versions and ciphers to accept for SSL/TLS terminated connections. If the `sslProfile` is not defined, AKO defaults to sslProfile `System-Standard` defined in Avi.

In case of reencrypt, if `destinationCA` is specified in the HTTPRule CRD, as shown in the example, a corresponding PKI profile is created for that pool (host path combination).

#### Status Messages

The status messages are used to give instanteneous feedback to the users about the whether a HTTPRule CRD was `Accepted` or `Rejected`.


##### Accepted HTTPRule

    $ kubectl get httprule
    NAME            HOSTRULE                     STATUS     AGE
    my-http-rules   default/secure-waf-policy    Accepted   5h34m


