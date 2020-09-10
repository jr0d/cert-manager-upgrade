package config

const (
	TargetVersion                    = "v1.0.1"
	StorageClassAnnotation    		 = "storageclass.kubernetes.io/is-default-class"

	DefaultCertManagerNamespace      = "cert-manager"
	DefaultCertManagerDeploymentName = "controller-cert-manager-kubeaddons"
	DefaultBackupPVName              = "cert-manager-upgrade-backups"
)
