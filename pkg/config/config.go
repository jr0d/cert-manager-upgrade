package config

type c struct {
	TargetVersion string
	CertManagerNamespace string
	CertManagerDeploymentName string
	BackupPvName string
}

var (
	AppConfig = c{}
	ResourcesToBackup = [...]string{
		"issuer", "clusterissuer", "certificates", "certificaterequests",
	}
)