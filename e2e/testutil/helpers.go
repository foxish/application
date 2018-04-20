// Copyright 2017 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pkg/errors"
)

func PathToOSFile(relativPath string) (*os.File, error) {
	path, err := filepath.Abs(relativPath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed generate absolut file path of %s", relativPath))
	}

	manifest, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to open file %s", path))
	}

	return manifest, nil
}

// WaitForPodsReady waits for a selection of Pods to be running and each
// container to pass its readiness check.
func WaitForPodsReady(kubeClient kubernetes.Interface, namespace string, timeout time.Duration, expectedReplicas int, opts metav1.ListOptions) error {
	return wait.Poll(time.Second, timeout, func() (bool, error) {
		pl, err := kubeClient.Core().Pods(namespace).List(opts)
		if err != nil {
			return false, err
		}

		runningAndReady := 0
		for _, p := range pl.Items {
			isRunningAndReady, err := PodRunningAndReady(p)
			if err != nil {
				return false, err
			}

			if isRunningAndReady {
				runningAndReady++
			}
		}

		if runningAndReady == expectedReplicas {
			return true, nil
		}
		return false, nil
	})
}

// WaitForCRDReady waits for a custom resource definition to be available for use.
func WaitForCRDReady(listFunc func(opts metav1.ListOptions) (runtime.Object, error)) error {
	err := wait.Poll(3*time.Second, 10*time.Minute, func() (bool, error) {
		_, err := listFunc(metav1.ListOptions{})
		if err != nil {
			if se, ok := err.(*apierrors.StatusError); ok {
				if se.Status().Code == http.StatusNotFound {
					return false, nil
				}
			}
			return false, errors.Wrap(err, "failed to list CRD")
		}
		return true, nil
	})

	return errors.Wrap(err, fmt.Sprintf("timed out waiting for Custom Resource"))
}

// PodRunningAndReady returns whether a pod is running and each container has
// passed it's ready state.
func PodRunningAndReady(pod v1.Pod) (bool, error) {
	switch pod.Status.Phase {
	case v1.PodFailed, v1.PodSucceeded:
		return false, fmt.Errorf("pod completed")
	case v1.PodRunning:
		for _, cond := range pod.Status.Conditions {
			if cond.Type != v1.PodReady {
				continue
			}
			return cond.Status == v1.ConditionTrue, nil
		}
		return false, fmt.Errorf("pod ready condition not found")
	}
	return false, nil
}

func GetLogs(kubeClient kubernetes.Interface, namespace string, podName, containerName string) (string, error) {
	logs, err := kubeClient.Core().RESTClient().Get().
		Resource("pods").
		Namespace(namespace).
		Name(podName).SubResource("log").
		Param("container", containerName).
		Do().
		Raw()
	if err != nil {
		return "", err
	}
	return string(logs), err
}

func Poll(timeout, pollInterval time.Duration, pollFunc func() (bool, error)) error {
	t := time.After(timeout)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t:
			return fmt.Errorf("timed out")
		case <-ticker.C:
			b, err := pollFunc()
			if err != nil {
				return err
			}
			if b {
				return nil
			}
		}
	}
}
