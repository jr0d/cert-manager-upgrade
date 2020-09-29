package app

import (
	"fmt"
	myk8s "github.com/jr0d/cert-manager-upgrade/pkg/kubernetes"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"log"
)

func Backup(kubeconfig string) error {
	cfg, err := myk8s.GetConfig(kubeconfig)
	if err != nil {
		return err
	}

	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	res, err := myk8s.HasV1Alpha1(c)
	if  err != nil {
		return err
	}

	if !res {
		fmt.Printf("v1alpha1 is not present, nothing to do.")
	}

	resources, err := myk8s.GetCertManagerResources(dyn)
	if err != nil {
		return err
	}

	log.Printf("backing up %d items", len(resources))
	if err := myk8s.BackupResources(c, resources); err != nil {
		return err
	}
	return nil
}
