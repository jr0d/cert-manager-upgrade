package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/jetstack/cert-manager/pkg/ctl"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apijson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/versioning"
	"log"

	"github.com/jr0d/cert-manager-upgrade/pkg/config"
)

func Convert(data []byte) (runtime.Object, error) {
	v1alpha2, err := convertV1Alpha2(data)
	if err != nil {
		return nil, fmt.Errorf("error converting to v1alpha2: %v", err)
	}

	obj, err := serializedConversion(v1alpha2)
	if err != nil {
		return nil, fmt.Errorf("error converting object to v1: %v", err)
	}
	return obj, nil
}

// We must first convert to v1alpha2 before attempting full conversion.
func convertV1Alpha2(data []byte) ([]byte, error) {
	var t map[string]interface{}
	err := json.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}

	u := unstructured.Unstructured{
		Object: t,
	}
	log.Printf("converting object: %s/%s", u.GetNamespace(), u.GetName())
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   config.CertManagerGroupV1Alpha2,
		Version: config.CertManagerVersionV1Alpha2,
		Kind:    u.GroupVersionKind().Kind,
	})
	return u.MarshalJSON()
}

var (
	// Use this scheme as it has the internal cert-manager types
	// and their conversion functions registered.
	scheme = ctl.Scheme

	serializer = apijson.NewSerializerWithOptions(
		apijson.DefaultMetaFactory,
		scheme,
		scheme,
		apijson.SerializerOptions{})

	gv = schema.GroupVersion{
		Group:   "cert-manager.io",
		Version: "v1",
	}

	c = versioning.NewCodec(
		serializer,
		serializer,
		runtime.UnsafeObjectConvertor(scheme),
		scheme,
		scheme,
		nil,
		gv,
		runtime.InternalGroupVersioner,
		scheme.Name())
)

func serializedConversion(resource []byte) (runtime.Object, error) {
	raw := runtime.RawExtension{
		Raw: resource,
	}

	dObj, _, err := c.Decode(raw.Raw, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error decoding object: %v\n", err)
	}

	object, err := scheme.ConvertToVersion(dObj, gv)

	if err != nil {
		return nil, fmt.Errorf("error converting object: %v", err)
	}
	return object, nil
}
