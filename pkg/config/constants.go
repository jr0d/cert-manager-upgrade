package config

const (
	TargetVersion = "v1.0.1"

	DefaultCertManagerNamespace = "cert-manager"

	CertManagerGroupV1Alpha1   = "certmanager.k8s.io"
	CertManagerVersionV1Alpha1 = "v1alpha1"

	CertManagerGroupV1Alpha2   = "cert-manager.io"
	CertManagerVersionV1Alpha2 = "v1alpha2"

	CertManagerGroupV1   = "cert-manager.io"
	CertManagerVersionV1 = "v1"

	InvalidCAInjectorAnnotation = "certmanager.k8s.io/allow-direct-injection"
	ValidCAInjectorAnnotation   = "cert-manager.io/allow-direct-injection"
)
