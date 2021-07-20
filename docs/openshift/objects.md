### Openshift Route
In Openshift cluster, AKO can be used to configure routes. Ingress configuration is not supported. 

### Insecure Route

    apiVersion: v1
    kind: Route
    metadata:
      name: route1
    spec:
      host: routehost1.avi.internal
      path: /foo
      to:
        kind: Service
        name: avisvc1

For insecure route, AKO creates a Shared VS, Poolgroup and Datascript like insecure ingress. For poolName, route configuration differs from ingress configuration. Service name is appended at the end of poolname, as illustrated bellow.

#### Shared VS pool names for route

The formula to derive the Shared VS pool name for route is as follows:

    poolname = clusterName + "--" + hostname + "-" + namespace + "-" + routeName + "-" + serviceName


### Insecure Route with Alternate Backends

A route can also be associated with multiple services denoted by alternateBackends. The requests that are handled by each service is governed by the service weight.

    apiVersion: v1
    kind: Route
    metadata:
      name: route1
    spec:
      host: routehost1.avi.internal
      path: /foo
      to:
        kind: Service
        name: avisvc1
        weight: 20
      alternateBackends:
      - kind: Service
        name: avisvc2
        weight: 10


For each backend of a route, a new pool is added. All such pools are added with same priority label - 'hostname/path'. For the above example, two pools would be added with priority - 'routehost1.avi.internal/foo'. Ratio for a pool is same as the weight specified for the service in the route.


### Secure Route with Edge Termination

    apiVersion: v1
    kind: Route
    metadata:
      name: secure-route1
    spec:
      host: secure1.avi.internal
      path: /bar
      to:
        kind: Service
        name: avisvc1
      tls:
        termination: edge
        key: |-
        -----BEGIN RSA PRIVATE KEY-----
        ...
        ...
        -----END RSA PRIVATE KEY-----

        certificate: |-
        -----BEGIN CERTIFICATE-----
        ...
        ...
        -----END CERTIFICATE-----

Secure route is configured in Avi like secure ingress. An SNI VS is created for each hostname and for each hostpath one poolgroup is created. However, for alternate backends, multiple pools are added in each poolgroup. Also, unlike Secure Ingress, no redirect policy is configured for Secure Route for insecure traffic.

##### SNI pool names for route

The formula to derive the SNI virtualservice's pools for route is as follows:

    poolname = clusterName + "--" + namespace + "-" + hostname + "_" + path + "-" + routeName + "-" serviceName


### Secure Route with termination reencrypt

    apiVersion: v1
      kind: Route
    metadata:
      name: secure-route1
    spec:
      host: secure1.avi.internal
      to:
        kind: Service
        name: service-name 
      tls:
        termination: reencrypt        
        key: |-
          -----BEGIN PRIVATE KEY-----
          [...]
          -----END PRIVATE KEY-----
        certificate: |-
          -----BEGIN CERTIFICATE-----
          [...]
          -----END CERTIFICATE-----
        destinationCACertificate: |-
          -----BEGIN CERTIFICATE-----
          [...]
          -----END CERTIFICATE-----

In case of reencrypt, an SNI VS is created for each hostname and for each host/path combination corresponds to a PoolGroup in Avi. Ssl is enabled in each pool for such Virtualservices with SSL profile set to System-Standard. In additon, if destinationCACertificate is specified, a PKI profile with the destinationCACertificate is created for each pool. 


### Secure Route insecureEdgeTerminationPolicy Redirect:

    apiVersion: v1
      kind: Route
    metadata:
      name: secure-route1
    spec:
      host: secure1.avi.internal
      to:
        kind: Service
        name: service-name 
      tls:
        termination: edge
        insecureEdgeTerminationPolicy: redirect      
        key: |-
          -----BEGIN PRIVATE KEY-----
          [...]
          -----END PRIVATE KEY-----
        certificate: |-
          -----BEGIN CERTIFICATE-----
          [...]
          -----END CERTIFICATE-----


In additon to the secure sni VS, for this type of route, AKO creates a redirect policy on the shared parent of the SNI child for this specific secure hostname.
This allows the client to automatically redirect the http requests to https if they are accessed on the insecure port (80).

### Secure Route insecureEdgeTerminationPolicy Allow:

    apiVersion: v1
      kind: Route
    metadata:
      name: secure-route1
    spec:
      host: secure1.avi.internal
      to:
        kind: Service
        name: service-name 
      tls:
        termination: edge
        insecureEdgeTerminationPolicy: Allow      
        key: |-
          -----BEGIN PRIVATE KEY-----
          [...]
          -----END PRIVATE KEY-----
        certificate: |-
          -----BEGIN CERTIFICATE-----
          [...]
          -----END CERTIFICATE-----


If insecureEdgeTerminationPolicy is Allow, then AKO creates an SNI VS for the hostname; also a pool is created for the same hostname which is added as member is Poolgroup of the parent Shared VS. This enables the host to be accessed via both http(80) and https(443) port. 

### Passthrough Route:

With passthrough routes, secure traffic is sent to the backend pods without TLS termination in AVI. A set of shared L4 Virtual Services are created by AKO to handle all tls passthrough routes. Number of shards can be configured in helm with the flag passthroughShardSize in values.yaml. These Virtual Services would listen on port 443 and have one L4 ssl datascript each. Name of the VS would be of the format clustername--'Shared-Passthrough'-shardnumner. Number of shards can be configured using the flag `passthroughShardSize` while installation using helm.

    apiVersion: v1
      kind: Route
    metadata:
      name: passthrough-route1
    spec:
      host: pass1.avi.internal
      to:
        kind: Service
        name: service-name 
      tls:
        termination: edge
        insecureEdgeTerminationPolicy: Allow   

For each passthrough host, one unique Poolgroup is created with name clustername-fqdn and the Poolgroup is attached to the datascript of the VS derived by the sharding logic. In this case, a Poolgroup with name clustername-pass1.avi.internal is created.

For each backend of a tls passthrough route, one pool is created with ratio as per the route spec and is attached to the corresponding PoolGroup.

If insecureEdgeTerminationPolicy is redirect, another Virtual Service is created for each shared L4 VS, to handle insecure traffic on port 80. HTTP Request polices would be added in this VS for each fqdn with insecureEdgeTermination policy set to redirect. Both the Virtual Services listening on port 443 and 80 have a common VSvip. This allows DNS VS to resolve the hostname to one IP address consistently. The name of the insecure shared VS would be of the format clustername--'Shared-Passthrough'-shard-number-'insecure'. 

For passthrough routes, insecureEdgeTerminationPolicy: Allow is not supported in openshift.


##### Multi-Port Service Support in Openshift

A service in openshift can have multiple ports. In order for a `Service` to have multiple ports, openshift mandates them to have a `name`. To use such a service, the user `must` specify the targetPort within port in route spec. The value of targetPort can be interger value of the target port or name of the port. If the backend service has only one port, then `port` field in route can be skipped, but it can not be skipped if the service has multiple ports. For example, consider the following Service:

        apiVersion: v1
        kind: Service
        metadata:
          labels:
            run: avisvc
        spec:
          ports:
          - name: myport1
            port: 80
            protocol: TCP
            targetPort: 80
          - name: myport2
            port: 8080
            protocol: TCP
            targetPort: 8080
          selector:
            app: my-app
          type: ClusterIP


In order to use this service in a route, the route spec can look like one of the following:

        apiVersion: v1
        kind: Route
        metadata:
          name: route1
        spec:
          host: routehost1.avi.internal
          path: /foo
          port:
            targetPort: 8080
          to:
            kind: Service
            name: avisvc1


        apiVersion: v1
        kind: Route
        metadata:
          name: route1
        spec:
          host: routehost1.avi.internal
          path: /foo
          port:
            targetPort: myport2
          to:
            kind: Service
            name: avisvc1