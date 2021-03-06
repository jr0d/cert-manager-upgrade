package config

type c struct {
	TargetVersion             string
	CertManagerNamespace      string
	CertManagerDeploymentName string
	BackupPvName              string
	PreserveCRDs              bool
	DryRun                    bool
	PreserveBackups           bool
	SkipFixSecrets            bool
}

var (
	AppConfig         = c{}
	ResourcesToBackup = [...]string{
		"issuers", "clusterissuers", "certificates", "certificaterequests",
	}
	ResourceKindMap = map[string]string{
		"Issuer":             "issuers",
		"ClusterIssuer":      "clusterissuers",
		"Certificate":        "certificates",
		"CertificateRequest": "certificaterequests",
	}
	MetaDataFieldsToRemove = [...]string{
		"creationTimestamp",
		"generation",
		"resourceVersion",
		"selfLink",
		"uid",
	}
)
