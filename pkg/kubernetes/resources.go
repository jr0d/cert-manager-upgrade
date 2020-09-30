package kubernetes

import (
	"fmt"
	"log"
	"strings"

	"github.com/jr0d/cert-manager-upgrade/pkg/config"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	certManagerBackupDataKey = "backupJSON"
	certManagerBackupLabel   = "cert-manager-resource-backup"
)

func HasV1Alpha1(c kubernetes.Interface) (bool, error) {
	groups, err := c.Discovery().ServerGroups()
	if err != nil {
		return false, err
	}

	for _, g := range groups.Groups {
		if g.Name == config.CertManagerGroupV1Alpha1 {
			for _, v := range g.Versions {
				if v.Version == config.CertManagerVersionV1Alpha1 {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func GetCertManagerResources(dyn dynamic.Interface) ([]unstructured.Unstructured, error) {
	var resources []unstructured.Unstructured
	for _, resource := range config.ResourcesToBackup {
		gvr := schema.GroupVersionResource{
			Group:    config.CertManagerGroupV1Alpha1,
			Version:  config.CertManagerVersionV1Alpha1,
			Resource: resource,
		}
		ul, err := dyn.Resource(gvr).Namespace(corev1.NamespaceAll).List(metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("error getting resource list: %v : %v", gvr, err)
		}
		resources = append(resources, ul.Items...)
	}
	return resources, nil
}

func BackupResources(c kubernetes.Interface, resources []unstructured.Unstructured) error {
	hasErrors := false
	for _, resource := range resources {
		log.Printf("backing up %s: %s/%s",
			resource.GetKind(), resource.GetNamespace(), resource.GetName())
		if err := storeResource(c, &resource); err != nil {
			log.Printf("%v", fmt.Errorf("error backing up %s/%s: %v",
				resource.GetNamespace(), resource.GetName(), err))
			hasErrors = true
		}
	}
	if hasErrors {
		return fmt.Errorf("could not backup all resources")
	}
	return nil
}

func DeleteBackups(c kubernetes.Interface) error {
	if err := c.CoreV1().ConfigMaps(
		config.AppConfig.CertManagerNamespace).DeleteCollection(
			&metav1.DeleteOptions{}, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("%s=true", certManagerBackupLabel)}); err != nil {
		return err
	}
	return nil
}

func DeleteCRDs(cfg *rest.Config) error {
	hasErrors := false
	var remove []apiextensionsv1.CustomResourceDefinition
	apiclient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return err
	}

	crdList, err := apiclient.CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, crd := range crdList.Items {
		if strings.Contains(crd.Name, "certmanager.k8s.io") {
			remove = append(remove, crd)
		}
	}

	for _, r := range remove {
		log.Printf("deleting CRD: %s", r.Name)
		if err := apiclient.CustomResourceDefinitions().Delete(
				r.Name, &metav1.DeleteOptions{}); err != nil {
			log.Printf("%v", fmt.Errorf("error deleting CRD %s : %v",
				r.Name, err))
			hasErrors = true
		}
	}
	if hasErrors {
		return fmt.Errorf("could not delete all CRDs")
	}
	return nil
}

func storeResource(c kubernetes.Interface, resource *unstructured.Unstructured) error {
	data, err := resource.MarshalJSON()
	if err != nil {
		return err
	}

	backupCfm := corev1.ConfigMap{
		ObjectMeta: resourceMeta(resource),
		Data: map[string]string{
			certManagerBackupDataKey: string(data),
		},
	}
	if _, err := c.CoreV1().ConfigMaps(config.AppConfig.CertManagerNamespace).
		Create(&backupCfm); err != nil {
		return err
	}
	return nil
}

func resourceMeta(resource *unstructured.Unstructured) metav1.ObjectMeta {
	om := metav1.ObjectMeta{
		GenerateName: fmt.Sprintf("%s-", resource.GetName()),
		Namespace:    config.AppConfig.CertManagerNamespace,
		Labels: map[string]string{
			certManagerBackupLabel: "true",
		},
	}
	return om
}
