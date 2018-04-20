package testutil

import (
	"io"
	"os"

	applicationsv1alpha1 "github.com/kubernetes-sigs/application/pkg/apis/app/v1alpha1"
	appcs "github.com/kubernetes-sigs/application/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func CreateApplication(kubeClient appcs.Interface, ns string, relativePath string) error {
	app, err := parseApplicationYaml(relativePath)
	if err != nil {
		return err
	}

	_, err = kubeClient.AppV1alpha1().Applications(ns).Get(app.Name, metav1.GetOptions{})

	if err == nil {
		// Application already exists -> Update
		_, err = kubeClient.AppV1alpha1().Applications(ns).Update(app)
		if err != nil {
			return err
		}

	} else {
		// Application doesn't exists -> Create
		_, err = kubeClient.AppV1alpha1().Applications(ns).Create(app)
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteApplication(kubeClient appcs.Interface, ns string, relativePath string) error {
	app, err := parseApplicationYaml(relativePath)
	if err != nil {
		return err
	}

	if err := kubeClient.AppV1alpha1().Applications(ns).Delete(app.Name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func parseApplicationYaml(relativePath string) (*applicationsv1alpha1.Application, error) {
	var manifest *os.File
	var err error

	var app applicationsv1alpha1.Application
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

		if out.GetKind() == "Application" {
			var marshaled []byte
			marshaled, err = out.MarshalJSON()
			json.Unmarshal(marshaled, &app)
			break
		}
	}

	if err != io.EOF && err != nil {
		return nil, err
	}
	return &app, nil
}
