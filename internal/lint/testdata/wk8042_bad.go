package testdata

import (
	corev1 "k8s.io/api/core/v1"
)

// WK8042: Private key headers
// This file contains violations

// Bad: RSA private key in ConfigMap
var ConfigMapWithRSAKey = corev1.ConfigMap{
	Data: map[string]string{
		"key.pem": `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF6CKqf7oNmMcXjM1u4N+bLSxQ
-----END RSA PRIVATE KEY-----`,
	},
}

// Bad: Private key in ConfigMap
var ConfigMapWithPrivateKey = corev1.ConfigMap{
	Data: map[string]string{
		"server.key": `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7
-----END PRIVATE KEY-----`,
	},
}

// Bad: EC private key
var ConfigMapWithECKey = corev1.ConfigMap{
	Data: map[string]string{
		"ec.key": `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIGlRvzRLFII3j3d5xVJGVpGu/JmVnW8F3gGVbqjF1DSoAoGCCqGSM49
-----END EC PRIVATE KEY-----`,
	},
}

// Bad: OpenSSH private key
var ConfigMapWithSSHKey = corev1.ConfigMap{
	Data: map[string]string{
		"id_rsa": `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
-----END OPENSSH PRIVATE KEY-----`,
	},
}
