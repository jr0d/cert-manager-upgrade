module github.com/jr0d/cert-manager-upgrade

go 1.15

require (
	github.com/spf13/cobra v1.0.0
	k8s.io/api v0.16.6
	k8s.io/apiextensions-apiserver v0.0.0-00010101000000-000000000000
	k8s.io/apimachinery v0.16.6
	k8s.io/client-go v0.16.6
)

replace (
	k8s.io/api => k8s.io/api v0.16.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.6
	k8s.io/apiserver => k8s.io/apiserver v0.16.6
	k8s.io/client-go => k8s.io/client-go v0.16.6
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.4.0
)
