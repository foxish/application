package testutil

import (
	"io"
	"os"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func CreateCRD(kubeClient apiextcs.Interface, relativePath string) error {
	CRD, err := parseCRDYaml(relativePath)
	if err != nil {
		return err
	}

	_, err = kubeClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(CRD.Name, metav1.GetOptions{})

	if err == nil {
		// ClusterRole already exists -> Update
		_, err = kubeClient.ApiextensionsV1beta1().CustomResourceDefinitions().Update(CRD)
		if err != nil {
			return err
		}

	} else {
		// ClusterRole doesn't exists -> Create
		_, err = kubeClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(CRD)
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteCRD(kubeClient apiextcs.Interface, relativePath string) error {
	CRD, err := parseCRDYaml(relativePath)
	if err != nil {
		return err
	}

	if err := kubeClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(CRD.Name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func parseCRDYaml(relativePath string) (*apiextensions.CustomResourceDefinition, error) {
	var manifest *os.File
	var err error

	var crd apiextensions.CustomResourceDefinition
	if manifest, err = PathToOSFile(relativePath); err != nil {
		return nil, err
	}

	decoder := yaml.NewYAMLOrJSONDecoder(manifest, 100)
	for {
		var out unstructured.Unstructured
		err = decoder.Decode(&out)
		if err != nil {
			// this would indicate it's malformed YAML.
			break
		}

		if out.GetKind() == "CustomResourceDefinition" {
			var marshaled []byte
			marshaled, err = out.MarshalJSON()
			json.Unmarshal(marshaled, &crd)
			break
		}
	}

	if err != io.EOF && err != nil {
		return nil, err
	}
	return &crd, nil
}
