package app

import (
	"fmt"
	"github.com/jr0d/cert-manager-upgrade/pkg/config"
	myk8s "github.com/jr0d/cert-manager-upgrade/pkg/kubernetes"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreatePVC(kubeconfig string) error {
	kcs, err := myk8s.GetNativeClientSet(kubeconfig)
	if err != nil {
		return err
	}

	_, err = kcs.
		CoreV1().
		PersistentVolumeClaims(
			config.AppConfig.CertManagerNamespace).Get(
		config.AppConfig.BackupPvName, metav1.GetOptions{})
	if !errors.IsNotFound(err) {
		return fmt.Errorf("backup pvc exists and should not: %v", err)
	}

	sc, ok, err := getDefaultStorageClass(kcs)
	if err != nil {
		return fmt.Errorf("error finding default storageclass: %v", err)
	}
	if !ok {
		return fmt.Errorf("default storageclass is not defined")
	}

	_, err = kcs.CoreV1().PersistentVolumeClaims(config.AppConfig.CertManagerNamespace).Create(
		&corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      config.AppConfig.BackupPvName,
				Namespace: config.AppConfig.CertManagerNamespace,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				VolumeName:       config.AppConfig.BackupPvName,
				StorageClassName: &sc.Name,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": resource.MustParse("20Mi"),
					},
				},
			},
		})
	return nil
}

func getDefaultStorageClass(c kubernetes.Interface) (*storagev1.StorageClass, bool, error) {
	defaultStorageClass := &storagev1.StorageClass{}

	sl, err := c.StorageV1().StorageClasses().List(metav1.ListOptions{})
	if err != nil {
		return nil, false, err
	}

	var found bool
	for _, sc := range sl.Items {
		b, ok := sc.ObjectMeta.Annotations[config.StorageClassAnnotation]
		if ok && b == "true" {
			*defaultStorageClass = sc
			found = true
			break
		}
	}
	return defaultStorageClass, found, nil
}
