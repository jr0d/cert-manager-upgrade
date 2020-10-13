package pkg

import (
	"encoding/json"
	"fmt"
	"sigs.k8s.io/yaml"
	"testing"
)

func TestConvertManaged(t *testing.T) {
	s := `apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  annotations:
    helm.sh/hook: post-install,post-upgrade
    helm.sh/hook-delete-policy: before-hook-creation
    helm.sh/hook-weight: "-3"
  creationTimestamp: "2020-10-07T03:33:38Z"
  generation: 3
  name: kubernetes-intermediate-ca
  namespace: cert-manager
  resourceVersion: "2436"
  selfLink: /apis/certmanager.k8s.io/v1alpha1/namespaces/cert-manager/certificates/kubernetes-intermediate-ca
  uid: 8901fc14-e7cf-479e-a1c1-edcae8ff1235
spec:
  commonName: cert-manager
  duration: 87600h0m0s
  isCA: true
  issuerRef:
    kind: Issuer
    name: kubernetes-root-issuer
  keyAlgorithm: rsa
  keyEncoding: pkcs1
  keySize: 2048
  secretName: kubernetes-intermediate-ca
  usages:
  - digital signature
  - key encipherment
`
	// o, err := managedConversion([]byte(s))

	var x map[string]interface{}

	if err := yaml.Unmarshal([]byte(s), &x); err != nil {
		panic(err)
	}

	d, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}

	o, err := serializedConversion(d)
	if err != nil {
		t.Fatalf(fmt.Sprintf("error: %v", err))
	}

	fmt.Printf("VVVV: %v\n", o)

	y, err := yaml.Marshal(o)
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	fmt.Printf("%s\n", y)

}
