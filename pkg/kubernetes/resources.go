package kubernetes

import (
	"context"
	"fmt"
	"log"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/jr0d/cert-manager-upgrade/pkg/config"
)

const (
	certManagerBackupDataKey = "backupJSON"
	certManagerBackupLabel   = "cert-manager-resource-backup"
)

func HasV1Alpha1(c kubernetes.Interface) (bool, error) {
	return hasGroupVersion(c, schema.GroupVersion{
		Group:   config.CertManagerGroupV1Alpha1,
		Version: config.CertManagerVersionV1Alpha1,
	})
}

func HasV1(c kubernetes.Interface) (bool, error) {
	return hasGroupVersion(c, schema.GroupVersion{
		Group:   config.CertManagerGroupV1,
		Version: config.CertManagerVersionV1,
	})
}

func GetCertManagerResources(dyn dynamic.Interface) ([]unstructured.Unstructured, error) {
	var resources []unstructured.Unstructured
	for _, resource := range config.ResourcesToBackup {
		gvr := schema.GroupVersionResource{
			Group:    config.CertManagerGroupV1Alpha1,
			Version:  config.CertManagerVersionV1Alpha1,
			Resource: resource,
		}
		ul, err := dyn.Resource(gvr).Namespace(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
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

func FetchBackups(c kubernetes.Interface) ([][]byte, error) {
	cms, err := c.CoreV1().ConfigMaps(
		config.AppConfig.CertManagerNamespace).List(
		context.TODO(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=true", certManagerBackupLabel)})
	if err != nil {
		return nil, err
	}

	data := make([][]byte, 0, len(cms.Items))
	for _, cm := range cms.Items {
		j, ok := cm.Data[certManagerBackupDataKey]
		if !ok {
			continue
		}
		data = append(data, []byte(j))
	}
	return data, nil
}

func ApplyObject(dyn dynamic.Interface, object runtime.Object) error {
	d, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return err
	}

	u := unstructured.Unstructured{Object: d}

	cleanObject(&u)

	gvk := u.GroupVersionKind()
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: config.ResourceKindMap[gvk.Kind],
	}

	var dr dynamic.ResourceInterface
	if u.GetNamespace() == "" {
		dr = dyn.Resource(gvr)
	} else {
		dr = dyn.Resource(gvr).Namespace(u.GetNamespace())
	}
	log.Printf("creating %s.%s.%s: %s/%s", gvr.Resource, gvr.Group, gvr.Version, u.GetNamespace(), u.GetName())

	o, err := dr.Get(context.TODO(), u.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = dr.Create(context.TODO(), &u, metav1.CreateOptions{})
		}
		return err
	}
	oGvk := o.GroupVersionKind()
	if oGvk.Group != config.CertManagerGroupV1 || oGvk.Version != config.CertManagerVersionV1 {
		return fmt.Errorf("object exists but is not converted: %s/%s : %v", o.GetNamespace(), o.GetName(), oGvk)
	}
	log.Printf("object is already converted: %s/%s", o.GetNamespace(), o.GetName())
	return nil
}

func DeleteBackups(c kubernetes.Interface) error {
	if err := c.CoreV1().ConfigMaps(
		config.AppConfig.CertManagerNamespace).DeleteCollection(context.TODO(),
		metav1.DeleteOptions{}, metav1.ListOptions{
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

	crdList, err := apiclient.CustomResourceDefinitions().List(context.TODO(), metav1.ListOptions{})
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
		if err := apiclient.CustomResourceDefinitions().Delete(context.TODO(),
			r.Name, metav1.DeleteOptions{}); err != nil {
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

func FixWebhookSecret(c kubernetes.Interface) error {
	secrets, err := getCaInjectorSecretsForUpgrade(c)
	if err != nil {
		return err
	}

	for _, secret := range secrets {
		delete(secret.Annotations, config.InvalidCAInjectorAnnotation)
		secret.Annotations[config.ValidCAInjectorAnnotation] = "true"

		log.Printf("updating secret to allow direct injection: %s/%s", secret.Namespace, secret.Name)
		_, err := c.CoreV1().Secrets(secret.Namespace).Update(
			context.TODO(), &secret, metav1.UpdateOptions{})

		if err != nil {
			return err
		}
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
		Create(context.TODO(), &backupCfm, metav1.CreateOptions{}); err != nil {
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

func hasGroupVersion(c kubernetes.Interface, gv schema.GroupVersion) (bool, error) {
	groups, err := c.Discovery().ServerGroups()
	if err != nil {
		return false, err
	}

	for _, g := range groups.Groups {
		if g.Name == gv.Group {
			for _, v := range g.Versions {
				if v.Version == gv.Version {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func cleanObject(u *unstructured.Unstructured) {
	for _, field := range config.MetaDataFieldsToRemove {
		unstructured.RemoveNestedField(u.Object, "metadata", field)
	}
	unstructured.RemoveNestedField(u.Object, "status")
}

func getCaInjectorSecretsForUpgrade(c kubernetes.Interface) ([]corev1.Secret, error) {
	var secrets []corev1.Secret
	secretList, err := c.CoreV1().Secrets("").List(
		context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, secret := range secretList.Items {
		if _, ok := secret.Annotations[config.InvalidCAInjectorAnnotation]; ok {
			secrets = append(secrets, secret)
		}
	}
	return secrets, nil
}
