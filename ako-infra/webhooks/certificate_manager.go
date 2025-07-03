/*
Copyright 2024 VMware, Inc.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhooks

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

const (
	// Certificate validity periods
	CACertValidityDays     = 365 * 10 // 10 years for CA
	ServerCertValidityDays = 365      // 1 year for server cert
	CertRotationThreshold  = 30       // Rotate when less than 30 days remaining

	// Certificate key sizes
	RSAKeySize = 2048

	// Secret keys
	CACertKey     = "ca.crt"
	CAKeyKey      = "ca.key"
	ServerCertKey = "tls.crt"
	ServerKeyKey  = "tls.key"
)

// VKSWebhookCertificateManager manages self-signed certificates for VKS webhook
type VKSWebhookCertificateManager struct {
	client        kubernetes.Interface
	namespace     string
	secretName    string
	serviceName   string
	certDir       string
	webhookConfig string
}

// NewVKSWebhookCertificateManager creates a new certificate manager
func NewVKSWebhookCertificateManager(client kubernetes.Interface, namespace, secretName, serviceName, certDir, webhookConfig string) *VKSWebhookCertificateManager {
	return &VKSWebhookCertificateManager{
		client:        client,
		namespace:     namespace,
		secretName:    secretName,
		serviceName:   serviceName,
		certDir:       certDir,
		webhookConfig: webhookConfig,
	}
}

// EnsureCertificates ensures valid certificates exist, creating or rotating them as needed
func (m *VKSWebhookCertificateManager) EnsureCertificates(ctx context.Context) error {
	utils.AviLog.Infof("VKS webhook certificate manager: ensuring certificates for service %s", m.serviceName)

	// Check if certificates need generation or rotation
	needsGeneration, err := m.needsCertificateGeneration(ctx)
	if err != nil {
		return fmt.Errorf("failed to check certificate status: %w", err)
	}

	if needsGeneration {
		utils.AviLog.Infof("VKS webhook certificate manager: generating new certificates")
		if err := m.generateCertificates(ctx); err != nil {
			return fmt.Errorf("failed to generate certificates: %w", err)
		}
	}

	// Write certificates to filesystem
	if err := m.writeCertificatesToFileSystem(ctx); err != nil {
		return fmt.Errorf("failed to write certificates to filesystem: %w", err)
	}

	// Update webhook configuration with CA bundle
	if err := m.updateWebhookConfiguration(ctx); err != nil {
		return fmt.Errorf("failed to update webhook configuration: %w", err)
	}

	utils.AviLog.Infof("VKS webhook certificate manager: certificates ready")
	return nil
}

// StartCertificateRotation starts a background goroutine for certificate rotation
func (m *VKSWebhookCertificateManager) StartCertificateRotation(ctx context.Context, rotationInterval time.Duration) {
	ticker := time.NewTicker(rotationInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				utils.AviLog.Infof("VKS webhook certificate manager: stopping certificate rotation")
				return
			case <-ticker.C:
				if err := m.EnsureCertificates(ctx); err != nil {
					utils.AviLog.Errorf("VKS webhook certificate manager: failed to rotate certificates: %v", err)
				}
			}
		}
	}()
	utils.AviLog.Infof("VKS webhook certificate manager: started certificate rotation with interval %v", rotationInterval)
}

// needsCertificateGeneration checks if certificates need to be generated or rotated
func (m *VKSWebhookCertificateManager) needsCertificateGeneration(ctx context.Context) (bool, error) {
	// Check if secret exists
	secret, err := m.client.CoreV1().Secrets(m.namespace).Get(ctx, m.secretName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			utils.AviLog.Infof("VKS webhook certificate manager: secret %s not found, will generate certificates", m.secretName)
			return true, nil
		}
		return false, err
	}

	// Check if secret has all required keys
	requiredKeys := []string{CACertKey, CAKeyKey, ServerCertKey, ServerKeyKey}
	for _, key := range requiredKeys {
		if _, exists := secret.Data[key]; !exists {
			utils.AviLog.Infof("VKS webhook certificate manager: secret missing key %s, will regenerate certificates", key)
			return true, nil
		}
	}

	// Check certificate validity
	serverCertPEM := secret.Data[ServerCertKey]
	block, _ := pem.Decode(serverCertPEM)
	if block == nil {
		utils.AviLog.Infof("VKS webhook certificate manager: invalid server certificate PEM, will regenerate")
		return true, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		utils.AviLog.Infof("VKS webhook certificate manager: failed to parse server certificate, will regenerate: %v", err)
		return true, nil
	}

	// Check if certificate is expiring soon
	timeUntilExpiry := time.Until(cert.NotAfter)
	if timeUntilExpiry < CertRotationThreshold*24*time.Hour {
		utils.AviLog.Infof("VKS webhook certificate manager: certificate expires in %v, will rotate", timeUntilExpiry)
		return true, nil
	}

	utils.AviLog.Infof("VKS webhook certificate manager: certificates are valid for %v more", timeUntilExpiry)
	return false, nil
}

// generateCertificates generates a new CA and server certificate
func (m *VKSWebhookCertificateManager) generateCertificates(ctx context.Context) error {
	// Generate CA private key
	caKey, err := rsa.GenerateKey(rand.Reader, RSAKeySize)
	if err != nil {
		return fmt.Errorf("failed to generate CA private key: %w", err)
	}

	// Generate CA certificate
	caCert, err := m.generateCACertificate(caKey)
	if err != nil {
		return fmt.Errorf("failed to generate CA certificate: %w", err)
	}

	// Generate server private key
	serverKey, err := rsa.GenerateKey(rand.Reader, RSAKeySize)
	if err != nil {
		return fmt.Errorf("failed to generate server private key: %w", err)
	}

	// Generate server certificate
	serverCert, err := m.generateServerCertificate(serverKey, caCert, caKey)
	if err != nil {
		return fmt.Errorf("failed to generate server certificate: %w", err)
	}

	// Encode certificates and keys to PEM format
	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCert.Raw})
	caKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caKey)})
	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCert.Raw})
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})

	// Create or update secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.secretName,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":      "ako",
				"app.kubernetes.io/component": "vks-webhook",
				"app.kubernetes.io/part-of":   "ako-vks-integration",
			},
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			CACertKey:     caCertPEM,
			CAKeyKey:      caKeyPEM,
			ServerCertKey: serverCertPEM,
			ServerKeyKey:  serverKeyPEM,
		},
	}

	// Try to update existing secret first
	_, err = m.client.CoreV1().Secrets(m.namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new secret
			_, err = m.client.CoreV1().Secrets(m.namespace).Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create certificate secret: %w", err)
			}
			utils.AviLog.Infof("VKS webhook certificate manager: created certificate secret %s", m.secretName)
		} else {
			return fmt.Errorf("failed to update certificate secret: %w", err)
		}
	} else {
		utils.AviLog.Infof("VKS webhook certificate manager: updated certificate secret %s", m.secretName)
	}

	return nil
}

// generateCACertificate generates a self-signed CA certificate
func (m *VKSWebhookCertificateManager) generateCACertificate(caKey *rsa.PrivateKey) (*x509.Certificate, error) {
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "VKS Webhook CA",
			Organization: []string{"VMware"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, CACertValidityDays),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// generateServerCertificate generates a server certificate signed by the CA
func (m *VKSWebhookCertificateManager) generateServerCertificate(serverKey *rsa.PrivateKey, caCert *x509.Certificate, caKey *rsa.PrivateKey) (*x509.Certificate, error) {
	// Generate DNS names for the webhook service
	dnsNames := []string{
		m.serviceName,
		fmt.Sprintf("%s.%s", m.serviceName, m.namespace),
		fmt.Sprintf("%s.%s.svc", m.serviceName, m.namespace),
		fmt.Sprintf("%s.%s.svc.cluster.local", m.serviceName, m.namespace),
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("%s.%s.svc", m.serviceName, m.namespace),
			Organization: []string{"VMware"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 0, ServerCertValidityDays),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    dnsNames,
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)}, // localhost for testing
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// writeCertificatesToFileSystem writes certificates to the filesystem for the webhook server
func (m *VKSWebhookCertificateManager) writeCertificatesToFileSystem(ctx context.Context) error {
	// Get secret
	secret, err := m.client.CoreV1().Secrets(m.namespace).Get(ctx, m.secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get certificate secret: %w", err)
	}

	// Ensure certificate directory exists
	if err := os.MkdirAll(m.certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Write server certificate
	certPath := filepath.Join(m.certDir, "tls.crt")
	if err := os.WriteFile(certPath, secret.Data[ServerCertKey], 0644); err != nil {
		return fmt.Errorf("failed to write server certificate: %w", err)
	}

	// Write server private key
	keyPath := filepath.Join(m.certDir, "tls.key")
	if err := os.WriteFile(keyPath, secret.Data[ServerKeyKey], 0600); err != nil {
		return fmt.Errorf("failed to write server private key: %w", err)
	}

	utils.AviLog.Infof("VKS webhook certificate manager: wrote certificates to %s", m.certDir)
	return nil
}

// updateWebhookConfiguration updates the MutatingWebhookConfiguration with the CA bundle
func (m *VKSWebhookCertificateManager) updateWebhookConfiguration(ctx context.Context) error {
	// Get secret
	secret, err := m.client.CoreV1().Secrets(m.namespace).Get(ctx, m.secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get certificate secret: %w", err)
	}

	// Get CA certificate
	caCertPEM := secret.Data[CACertKey]
	if len(caCertPEM) == 0 {
		return fmt.Errorf("CA certificate is empty")
	}

	// Get webhook configuration
	webhookConfig, err := m.client.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, m.webhookConfig, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			utils.AviLog.Infof("VKS webhook certificate manager: webhook configuration %s not found, skipping CA bundle update", m.webhookConfig)
			return nil
		}
		return fmt.Errorf("failed to get webhook configuration: %w", err)
	}

	// Update CA bundle for all webhooks
	updated := false
	for i := range webhookConfig.Webhooks {
		if !equalByteSlices(webhookConfig.Webhooks[i].ClientConfig.CABundle, caCertPEM) {
			webhookConfig.Webhooks[i].ClientConfig.CABundle = caCertPEM
			updated = true
		}
	}

	if updated {
		_, err = m.client.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(ctx, webhookConfig, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update webhook configuration: %w", err)
		}
		utils.AviLog.Infof("VKS webhook certificate manager: updated CA bundle in webhook configuration %s", m.webhookConfig)
	} else {
		utils.AviLog.Infof("VKS webhook certificate manager: CA bundle already up to date in webhook configuration %s", m.webhookConfig)
	}

	return nil
}

// GetCABundle returns the base64 encoded CA certificate for Helm values
func (m *VKSWebhookCertificateManager) GetCABundle(ctx context.Context) (string, error) {
	secret, err := m.client.CoreV1().Secrets(m.namespace).Get(ctx, m.secretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get certificate secret: %w", err)
	}

	caCertPEM := secret.Data[CACertKey]
	if len(caCertPEM) == 0 {
		return "", fmt.Errorf("CA certificate is empty")
	}

	return base64.StdEncoding.EncodeToString(caCertPEM), nil
}

// ValidateCertificates validates that the certificates are properly configured
func (m *VKSWebhookCertificateManager) ValidateCertificates(ctx context.Context) error {
	// Check filesystem certificates
	certPath := filepath.Join(m.certDir, "tls.crt")
	keyPath := filepath.Join(m.certDir, "tls.key")

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return fmt.Errorf("server certificate file not found: %s", certPath)
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("server private key file not found: %s", keyPath)
	}

	// Load and validate certificate
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("failed to load certificate pair: %w", err)
	}

	// Parse certificate to check validity
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check if certificate is still valid
	now := time.Now()
	if now.Before(x509Cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid (valid from %v)", x509Cert.NotBefore)
	}

	if now.After(x509Cert.NotAfter) {
		return fmt.Errorf("certificate has expired (expired on %v)", x509Cert.NotAfter)
	}

	// Check DNS names
	expectedDNS := fmt.Sprintf("%s.%s.svc", m.serviceName, m.namespace)
	found := false
	for _, dnsName := range x509Cert.DNSNames {
		if dnsName == expectedDNS {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("certificate does not contain expected DNS name %s, found: %v", expectedDNS, x509Cert.DNSNames)
	}

	utils.AviLog.Infof("VKS webhook certificate manager: certificates validated successfully")
	return nil
}

// equalByteSlices compares two byte slices for equality
func equalByteSlices(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
