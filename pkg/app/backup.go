package app

import (
	"log"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/jr0d/cert-manager-upgrade/pkg/config"
	myk8s "github.com/jr0d/cert-manager-upgrade/pkg/kubernetes"
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
	if err != nil {
		return err
	}

	if !res {
		log.Printf("v1alpha1 is not present, nothing to do.")
		return nil
	}

	resources, err := myk8s.GetCertManagerResources(dyn)
	if err != nil {
		return err
	}

	log.Printf("backing up %d items", len(resources))
	if err := myk8s.BackupResources(c, resources); err != nil {
		return err
	}

	if !config.AppConfig.PreserveCRDs {
		log.Printf("deleting CRDs...")
		if err := myk8s.DeleteCRDs(cfg); err != nil {
			return err
		}
	}

	if !config.AppConfig.SkipFixSecrets {
		log.Printf("fixing secrets for direct injection...")
		if err := myk8s.FixWebhookSecret(c); err != nil {
			return err
		}
	}

	return nil
}
