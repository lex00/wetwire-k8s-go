package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8042: Private key headers
// This file contains no violations

// Good: Public key is safe
var ConfigMapWithPublicKey = corev1.ConfigMap{
	Data: map[string]string{
		"public.pem": `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0Z3VS5JJcds3xfn/ygWy
-----END PUBLIC KEY-----`,
	},
}

// Good: Certificate is safe
var ConfigMapWithCertificate = corev1.ConfigMap{
	Data: map[string]string{
		"ca.crt": `-----BEGIN CERTIFICATE-----
MIIDQTCCAimgAwIBAgITBmyfz5m/jAo54vB4ikPmljZbyjANBgkqhkiG9w0BAQsF
-----END CERTIFICATE-----`,
	},
}

// Good: Using secret for private keys
var SecretForPrivateKey = corev1.Secret{
	Type: corev1.SecretTypeTLS,
	Data: map[string][]byte{
		"tls.key": []byte("key-data-here"),
		"tls.crt": []byte("cert-data-here"),
	},
}

// Good: ConfigMap with safe data
var ConfigMapSafeData = corev1.ConfigMap{
	Data: map[string]string{
		"config.json": `{"server": "https://example.com"}`,
	},
}
