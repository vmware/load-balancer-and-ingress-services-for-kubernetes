### SSORule

SSORule CRD can be used to configure OAuth or SAML specific properties for L7 virtual services that are created  by the AKO from Kubernetes Ingress and Openshift Route objects. The SSORule CRD can later be extended to support other SSO protocols that are supported by Avi.  
The SSORule CRD specifies an `fqdn` field, which is used to attach the SSORule to a virtual service. A given SSORule is applied to a virtual service if the virtual service hosts the fqdn specified in the SSORule CRD object.

The SSORule CRD can be used to configure either OAuth/OIDC properties or SAML service provider properties, along with the SSO policy.  
A sample SSORule CRD object with OAuth/OIDC properties looks like this:

```yaml
apiVersion: ako.vmware.com/v1alpha2
kind: SSORule
metadata:
  name: my-ssorule
  namespace: default
spec:
  fqdn: my-ssorule.test.com
  oauthVsConfig:
    cookieName: OAUTH_XYZ
    cookieTimeout: 120
    logoutURI: https://my-ssorule.test.com/oauth/logout
    oauthSettings:
    - appSettings:
        clientID: my-client-id
        clientSecret: my-oauth-secret
        oidcConfig:
          oidcEnable: true
          profile: true
          userinfo: true
        scopes: ["scope1"]
      authProfileRef: okta-oauth
      resourceServer:
        accessType: ACCESS_TOKEN_TYPE_OPAQUE
        introspectionDataTimeout: 60
        opaqueTokenParams:
          serverID: my-server-id
          serverSecret: my-oauth-secret
    redirectURI: https://my-ssorule.test.com/oauth/callback
    postLogoutRedirectURI: https://my-ssorule.test.com/oauth/postLogoutRedirectURI
  ssoPolicyRef: oauth
```

A sample SSORule CRD object with SAML properties looks like this:
```yaml
apiVersion: ako.vmware.com/v1alpha2
kind: SSORule
metadata:
  name: my-ssorule
  namespace: default
spec:
  fqdn: my-ssorule.test.com
  samlSpConfig:
    authnReqAcsType: SAML_AUTHN_REQ_ACS_TYPE_NONE
    cookieName: saml-cookie
    cookieTimeout: 120
    entityID: my-ssorule.test.com
    singleSignonURL: https://my-ssorule.test.com/sso/acs/
    signingSslKeyAndCertificateRef: my-test-app-secret
    useIdpSessionTimeout: false
  ssoPolicyRef: saml-app

```

**NOTE**: SSORule CRD is only supported for Enhanced Virtual Hosting (EVH). The OAuth and SAML settings will only be configured on virtual services when EVH is enabled. When the shard virtual service size is **LARGE** or **MEDIUM** or **SMALL**, the OAuth and SAML settings will only be configured on the EVH child virtual services.

### Specific usage of SSORule CRD

SSORule CRD can be created in a given namespace where the operator desires to have more control. 
The section below walks over the details and associated rules for using each field of the SSORule CRD.

#### SSORule to VS matching using fqdn

A given SSORule is applied to a virtual service if the VS hosts the `fqdn` mentioned in the SSORule CRD. This `fdqn` must exactly match the one the virtual service is hosting.

```yaml
  fqdn: my-ssorule.test.com
```

#### Express OAuth/OIDC Configuration for Virtual Service

The SSORule CRD can be used to configure the OAuth/OIDC properties of an L7 virtual service. The OAuth/OIDC configuration needs to be specified in the `oauthVsConfig` field.

```yaml
  oauthVsConfig:
    cookieName: OAUTH_XYZ
    cookieTimeout: 120
    logoutURI: https://my-ssorule.test.com/oauth/logout
    oauthSettings:
    - appSettings:
        clientID: my-client-id
        clientSecret: my-oauth-secret
        oidcConfig:
          oidcEnable: true
          profile: true
          userinfo: true
        scopes: ["scope1"]
      authProfileRef: okta-oauth
      resourceServer:
        accessType: ACCESS_TOKEN_TYPE_OPAQUE
        introspectionDataTimeout: 60
        opaqueTokenParams:
          serverID: my-server-id
          serverSecret: my-oauth-secret
    redirectURI: https://my-ssorule.test.com/oauth/callback
    postLogoutRedirectURI: https://my-ssorule.test.com/oauth/postLogoutRedirectURI
```
The properties under the `oauthVsConfig` field are discussed in detail below.

#### Express name for authorized session cookie

The `cookieName` property can be used to express the name of HTTP cookie for an authorized session. If this field is not specified, NSX ALB will create the authorized session cookie with a random name.

```yaml
  cookieName: OAUTH_XYZ
```

#### Express timeout for authorized session cookie

The `cookieTimeout` property can be used to specify the HTTP cookie timeout in **minutes** for an authorized session. It supports values from **1** to **1440** and defaults to **60**.

```yaml
  cookieTimeout: 120
```

#### Express Logout URI for OAuth

The `logoutURI` property can be used to express the URI that will trigger OAuth logout.  

```yaml
  logoutURI: https://my-ssorule.test.com/oauth/logout
```

#### Express Application and IDP settings for OAuth/OIDC

The `oauthSettings` property can be used to express the application specific OAuth configuration and Identity Provider (IDP) settings for OAuth/OIDC.  
The application specific properties are specified as a list of `appSettings` properties.

#### Express Client ID

The `clientID` property can be used to express the client ID for the application. It is an application specific identifier registered with the authorization server or Identity Provider (IDP).

```yaml
  clientID: my-client-id
```

#### Express Client Secret

The `clientSecret` field can be used to express the client secret for the application. It is an application specific identifier secret registered with the authorization server, or IDP. Since `clientSecret` is a sensetive field in NSX ALB, AKO requires it to be specified inside a Kubernetes Secret object. So, this `clientSecret` field should be the name of the Kubernetes secret object that specifies the actual client secret value (Base64 encoded) in the **clientSecret** data field.

```yaml
  clientSecret: my-oauth-secret
```

A sample Kubernetes secret object, with the actual value for client and secret (Base64 encoded) in the **clientSecret** data field is shown below.  

```yaml
  apiVersion: v1
  data:
    clientSecret: bXktY2xpZW50LXNlY3JldA==
    serverSecret: bXktc2VydmVyLXNlY3JldA==
  kind: Secret
  metadata:
    name: my-oauth-secret
  type: Opaque
```

#### Express OpenID Connect configuration

The `oidcConfig` property can be used to express OpenID Connect specific configuration.

```yaml
  oidcConfig:
    oidcEnable: true
    profile: true
    userinfo: true
```
`oidcEnable`, if set to **true**, adds OpenID as one of the scopes enabling the OpenID Connect flow.  
`profile`, if set to **true**, will allow fetching profile information by enabling profile scope. It defaults to **true**.   
`userinfo`, if set to **true**, will allow fetching profile information from the Userinfo Endpoint.

#### Express Scope for OAuth

The `scopes` property can be used to express the scope for OAuth to give limited access to the application.

```yaml
  scopes: ["scope1"]
```

#### Express Authentication Profile for OAuth

SSORule CRD can be used to express the authentication profile reference for OAuth. The authentication profile must be created in the AVI Controller before referring to it.

```yaml
  authProfileRef: okta-oauth
```
The Auth Profile should specify all the endpoints and configurations associated with the Identity Provider (IDP) and will be used for validating users.

#### Express Resource Server OAuth config

SSORule CRD can be used to express Resource Server OAuth configuration. The resource server configuration properties are described in detail below.

#### Express Access token type

The `accessType` property can be used to express the access token type. The access token type can either be opaque or JWT.

```yaml
  accessType: ACCESS_TOKEN_TYPE_OPAQUE
```

OR

```yaml
  accessType: ACCESS_TOKEN_TYPE_JWT
```

#### Express Introspection Data Timeout

The `introspectionDataTimeout` property can be used to set the lifetime of the cached introspection data in **minutes**. It supports values from **Zero (0)** to **1440** and defaults to **Zero (0)**. However, it will be set only if the access token type is opaque.

```yaml
  introspectionDataTimeout: 60
```

#### Express Opaque Token Parameters

The `opaqueTokenParams` property should be specified if the `accessType` property is **ACCESS_TOKEN_TYPE_OPAQUE**. It can be used to express the validation parameters to be used when the access token type is opaque.

```yaml
  opaqueTokenParams:
    serverID: my-server-id
    serverSecret: my-oauth-secret
```

The `serverID` property can be used to express the server ID for the resource server. It is the resource server specific identifier registered with the authorization server or Identity Provider(IDP), and is used to validate against the introspection endpoint when the access token type is opaque.  

The `serverSecret` field can be used to express the server secret for the resource server. It is a resource server specific identifier secret registered with the authorization server, or IDP. Since `serverSecret` is a sensetive field in NSX ALB, AKO requires it to be specified inside a Kubernetes Secret object. So, this `serverSecret` field should be the name of the Kubernetes secret object that specifies the actual server secret value (Base64 encoded) in the **serverSecret** data field. The server and client secrets can be specified in the same or different Kubernetes objects, as already shown in [Express Client Secret.](#express-client-secret)

#### Express Redirect URI for OAuth

SSORule CRD can be used to express the redirect URI that is specified in the request to the authorization server or Identity Provider (IDP). The redirect URI is the callback entry point of the application where the authorization server sends the user once the application has been successfully authorized and granted an authorization code.

```yaml
  redirectURI: https://my-ssorule.test.com/oauth/callback
```

#### Express Post Logout Redirect URI for OAuth

SSORule CRD can be used to express the post-logout redirect URI for the application. It is the URI to which the Identity Provider (IDP) will redirect after application logs out of the IDP.

```yaml
  postLogoutRedirectURI: https://my-ssorule.test.com/oauth/postLogoutRedirectURI
```

#### Express application specific SAML configuration

SSORule CRD can be used to express the application specific configuration, i.e., SAML service provider configuration for the L7 virtual service. The SAML service provider configuration needs to be specified in the `samlSpConfig` field.

```yaml
  samlSpConfig:
    acsIndex: 64
    authnReqAcsType: SAML_AUTHN_REQ_ACS_TYPE_INDEX
    cookieName: saml-cookie
    cookieTimeout: 120
    entityID: my-ssorule.test.com
    singleSignonURL: https://my-ssorule.test.com/sso/acs/
    signingSslKeyAndCertificateRef: my-test-app-secret
    useIdpSessionTimeout: false
```

The properties under the `samlSpConfig` field are discussed in detail below.

#### Express Assertion Consumer Service Index

The `acsIndex` property can be used to express the index to be used in the AssertionConsumerServiceIndex attribute of the authentication request. It will be set only if the `authnReqAcsType` is set to **SAML_AUTHN_REQ_ACS_TYPE_INDEX**. It supports values from **Zero (0)** to **64**.

```yaml
  acsIndex: 64
```

#### Express Assertion Consumer Service Type for Authentication Request

The `authnReqAcsType` property can be used to express the assertion consumer service type for authentication requests. It will determine the ACS attributes that will be set in the authentication request. It supports the following three values:

When `authnReqAcsType` is set to **SAML_AUTHN_REQ_ACS_TYPE_NONE**, no ACS attributes will be set in the SAML authentication request.

```yaml
  authnReqAcsType: SAML_AUTHN_REQ_ACS_TYPE_NONE
```

When `authnReqAcsType` is set to **SAML_AUTHN_REQ_ACS_TYPE_URL**, the AssertionConsumerServiceURL attribute will be set in the SAML authentication request. The ACS URL should be equal to the single signon URL set for the virtual service.

```yaml
  authnReqAcsType: SAML_AUTHN_REQ_ACS_TYPE_URL
```

When `authnReqAcsType` is set to **SAML_AUTHN_REQ_ACS_TYPE_INDEX**, the AssertionConsumerServiceIndex attribute of the SAML authentication request will be set with the value specified in the `acsIndex` property.

```yaml
  authnReqAcsType: SAML_AUTHN_REQ_ACS_TYPE_INDEX
```

#### Express name for authenticated session cookie

The `cookieName` property can be used to express the name of HTTP cookie for an authenticated session. If this field is not specified, NSX ALB will create the authenticated session cookie with a random name.

```yaml
  cookieName: saml-cookie
```

#### Express timeout for authenticated session cookie

The `cookieTimeout` property can be used to specify the timeout for HTTP cookie in **minutes** for an authenticated session. It supports values from **1** to **1440** and defaults to **60**.

```yaml
  cookieTimeout: 120
```

#### Express Entity ID for SAML Service Provider Application

The `entityID` property can be used to express the globally unique entity ID for the SAML Service Provider application. The SAML application entity ID on the Identity Provider (IDP) should match this.

```yaml
  entityID: my-ssorule.test.com
```

#### Express SslKeyAndCertificateRef for SAML Service Provider application

The `signingSslKeyAndCertificateRef` property can be used to express **SslKeyAndCertificate** reference for the SAML Service Provider application. The **SslKeyAndCertificate** must be created in the AVI Controller before referring to it. The service provider will use this SSL certificate to sign requests going to the Identity Provider (IDP) and decrypt the assertions coming from IDP.

```yaml
  signingSslKeyAndCertificateRef: my-test-app-secret
```

#### Express Single Signon URL for SAML

The `singleSignonURL` property can be used to express the single signon URL for the application. It specifies the endpoint to receive the authentication response and the destination endpoint to be configured for the application on the IDP. If the `authnReqAcsType` is set to **SAML_AUTHN_REQ_ACS_TYPE_URL**, this endpoint will be sent in the AssertionConsumerServiceURL attribute of the authentication request.

```yaml
  singleSignonURL: https://my-ssorule.test.com/sso/acs/
```

#### Express IDP control for service provider session timeout

The `useIdpSessionTimeout` property can be used to enable the Identity Provider (IDP) to control how long the Service Provider (SP) session can exist through the SessionNotOnOrAfter field in the AuthNStatement of the SAML Response.

```yaml
  useIdpSessionTimeout: false
```

#### Express SSO Policy for the Virtual Service

The SSORule CRD can be used to express the SSO policy reference for the virtual service. The SSO Policy must be created in the AVI Controller before referring to it. The SSO policy can be of type OAUTH or SAML, depending on the SSO protocol that needs to be configured for the virtual service.

```yaml
  ssoPolicyRef: my-sso-policy
```

**NOTE**: The `oauthVsConfig` and `samlSpConfig` are **oneOf** fields, and only one can be set since a virtual service can specify either OAuth or SAML as the SSO protocol. 

#### Status Messages

The status messages are used to give instantaneous feedback to the users about the reference objects specified in the SSORule CRD.

Following are some of the sample status messages:

##### Accepted SSORule object

    $ kubectl get ssorule
    NAME         STATUS     AGE
    my-sso-rule   Accepted   3d5s

An SSORule is accepted only when all the reference objects specified inside it exist in the AVI Controller.

##### Rejected SSORule object

    $ kubectl get ssorule
    NAME            STATUS     AGE
    my-sso-rule-alt  Rejected   2d23h

The detailed rejection reason can be obtained from the status:

```yaml
  status:
    error: authprofile "okta-oauth" not found on controller
    status: Rejected
```

#### Conditions and Caveats

##### SSORule is only supported for Enhanced Virtual Hosting (EVH)

SSORule CRD is only supported for Enhanced Virtual Hosting (EVH). The OAuth and SAML settings will only be configured on virtual services when EVH is enabled. When shard virtual service size is **LARGE** or **MEDIUM** or **SMALL**, the OAuth and SAML settings will only be configured on the EVH child virtual services.

##### SSORule deletion

If an SSORule is deleted, all the settings for the FQDNs are withdrawn from the Avi controller.

##### SSORule admission

An SSORule CRD is only admitted if all the objects referenced in it exist in the AVI Controller. If, after admission, the object references are
deleted out-of-band, then AKO does not re-validate the associated SSORule CRD objects. The user needs to manually edit or delete the object, for new changes to take effect.

##### Duplicate FQDN rules

Two SSORule CRDs cannot be used for the same FQDN information across namespaces. If AKO finds a duplicate FQDN in more than one SSORule, AKO honours the first SSORule that gets created and rejects the others. In the case of AKO reboots, the CRD that gets honoured might not be the same as the one honoured earlier.