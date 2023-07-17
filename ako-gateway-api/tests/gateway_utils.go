package tests

import (
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func SetGatewayName(gw *gatewayv1beta1.Gateway, name string) {
	gw.Name = name
}
func UnsetGatewayName(gw *gatewayv1beta1.Gateway) {
	gw.Name = ""
}

func SetGatewayGatewayClass(gw *gatewayv1beta1.Gateway, name string) {
	gw.Spec.GatewayClassName = gatewayv1beta1.ObjectName(name)
}
func UnsetGatewayGatewayClass(gw *gatewayv1beta1.Gateway) {
	gw.Spec.GatewayClassName = ""
}

func AddGatewayListener(gw *gatewayv1beta1.Gateway, name string, port int32, protocol gatewayv1beta1.ProtocolType, isTLS bool) {

	listner := gatewayv1beta1.Listener{
		Name:     gatewayv1beta1.SectionName(name),
		Port:     gatewayv1beta1.PortNumber(port),
		Protocol: protocol,
	}
	if isTLS {
		SetListenerTLS(&listner, gatewayv1beta1.TLSModeTerminate, "secret-example", "default")
	}
	gw.Spec.Listeners = append(gw.Spec.Listeners, listner)
}

func SetListenerTLS(l *gatewayv1beta1.Listener, tlsMode gatewayv1beta1.TLSModeType, secretName, secretNS string) {
	l.TLS = &gatewayv1beta1.GatewayTLSConfig{Mode: &tlsMode}
	namespace := gatewayv1beta1.Namespace(secretNS)
	kind := gatewayv1beta1.Kind("Secret")
	l.TLS.CertificateRefs = []gatewayv1beta1.SecretObjectReference{
		{
			Name:      gatewayv1beta1.ObjectName(secretName),
			Namespace: &namespace,
			Kind:      &kind,
		},
	}
}
func UnsetListenerTLS(l *gatewayv1beta1.Listener) {
	l.TLS = &gatewayv1beta1.GatewayTLSConfig{}
}

func SetListenerHostname(l *gatewayv1beta1.Listener, hostname string) {
	l.Hostname = (*gatewayv1beta1.Hostname)(&hostname)
}
func UnsetListenerHostname(l *gatewayv1beta1.Listener) {
	var hname gatewayv1beta1.Hostname
	l.Hostname = &hname
}
