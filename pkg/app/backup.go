package app

import (
	"fmt"

	myk8s "github.com/jr0d/cert-manager-upgrade/pkg/kubernetes"
)

func Backup(kubeconfig string) error {
	kcs, err := myk8s.GetNativeClientSet(kubeconfig)
	if err != nil {
		return err
	}

	sc, ok, err := myk8s.GetDefaultStorageClass(kcs)
	if err != nil {
		return fmt.Errorf("problem getting storageclass: %v", err)
	}
	if !ok {
		return fmt.Errorf("no default storageclass is defined")
	}

	fmt.Printf("SC:\n%s\n", sc)
	return nil
}
