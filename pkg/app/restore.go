package app

import (
	"fmt"
	"github.com/jr0d/cert-manager-upgrade/pkg/config"
	"log"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	upgrade "github.com/jr0d/cert-manager-upgrade/pkg"
	myk8s "github.com/jr0d/cert-manager-upgrade/pkg/kubernetes"
)

func Restore(kubeconfig string) error {
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

	res, err := myk8s.HasV1(c)

	if err != nil {
		return err
	}

	if !res {
		return fmt.Errorf(
			"api server does not have cert-manager v1 support, before attempting restore")
	}

	backups, err := myk8s.FetchBackups(c)
	if err != nil {
		return err
	}

	var errors []error
	converted := make([]runtime.Object, 0, len(backups))
	for _, backup := range backups {
		o, err := upgrade.Convert(backup)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		converted = append(converted, o)
	}

	if len(errors) > 0 {
		for _, e := range errors {
			log.Printf("error converting resource: %v", e)
		}
		return fmt.Errorf("error encountered while converting some resources, stopping")
	}

	// restore
	for _, obj := range converted {
		err := myk8s.ApplyObject(dyn, obj)
		if err != nil {
			// stop on fist failure
			return fmt.Errorf("error applying object: %v", err)
		}
	}

	// validate
	// TODO

	// destroy backups
	if !config.AppConfig.PreserveBackups {
		if err := myk8s.DeleteBackups(c); err != nil {
			return fmt.Errorf("error deleting backups: %v", err)
		}
	}
	return nil
}
