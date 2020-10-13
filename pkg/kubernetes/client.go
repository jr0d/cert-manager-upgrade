package kubernetes

import (
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetConfig(kubeconfig string) (*restclient.Config, error) {
	var cfg *restclient.Config
	var err error
	if kubeconfig == "" {
		cfg, err = restclient.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}
