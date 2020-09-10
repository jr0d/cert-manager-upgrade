package kubernetes

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes"

	"github.com/jr0d/cert-manager-upgrade/pkg/config"
)

// GetCertManagerVersion extracts the cert-manager controller version from the deployment image tag
func GetCertManagerVersion(c kubernetes.Interface, namespace, deploymentName string) (*utilversion.Version, error) {
	deployment, err := c.AppsV1().Deployments(namespace).Get(deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting cert-manager deployment: %v", err)
	}

	var imgTag string
	containers := deployment.Spec.Template.Spec.Containers
	for i := range containers {
		if strings.Contains(containers[i].Image, "cert-manager-controller") {
			s := strings.Split(containers[i].Image, ":")
			if len(s) != 2 {
				continue
			}
			imgTag = s[1]
			break
		}
	}

	if imgTag == "" {
		return nil, fmt.Errorf("could not determine cert-manager version")
	}

	v, err := utilversion.ParseSemantic(imgTag)
	if err != nil {
		return nil, fmt.Errorf("error parsing cert-manager version: %v", containers)
	}

	return v, nil
}

func needsUpdate(v utilversion.Version, target string) bool {
	return v.LessThan(utilversion.MustParseSemantic(target))
}

func GetDefaultStorageClass(c kubernetes.Interface) (*storagev1.StorageClass, bool, error) {
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

func CreatePVC(c kubernetes.Interface, storageClassName string) (*corev1.PersistentVolumeClaim, error) {
	return c.CoreV1().PersistentVolumeClaims(config.AppConfig.CertManagerNamespace).Create(
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
				StorageClassName: &storageClassName,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": resource.MustParse("20Mi"),
					},
				},
			},
		})
}