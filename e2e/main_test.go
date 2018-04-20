package main_test

import (
	"log"
	"os"
	"path"
	"testing"

	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kubernetes-sigs/application/e2e/testutil"
	appcs "github.com/kubernetes-sigs/application/pkg/client/clientset/versioned"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"fmt"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("/_workspace/junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Application Type Suite", []Reporter{junitReporter})
}

func getClientConfig() (*rest.Config, error) {
	fmt.Println(os.Getenv("HOME"))
	return clientcmd.BuildConfigFromFlags("", path.Join(os.Getenv("HOME"), ".kube/config"))
}

var _ = Describe("Application CRD should install correctly", func() {
	config, err := getClientConfig()
	if err != nil {
		log.Fatal("Unable to get client configuration", err)
	}

	extClient, err := apiextcs.NewForConfig(config)
	if err != nil {
		log.Fatal("Unable to construct extensions client", err)
	}

	appClient, err := appcs.NewForConfig(config)
	if err != nil {
		log.Fatal("Unable to construct applications client", err)
	}

	It("should create CRD", func() {
		err = testutil.CreateCRD(extClient, "../hack/install.yaml")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should register an application", func() {
		err = testutil.CreateApplication(appClient, "default", "../hack/sample/application.yaml")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should delete application", func() {
		err = testutil.DeleteApplication(appClient, "default", "../hack/sample/application.yaml")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should delete application CRD", func() {
		err = testutil.DeleteCRD(extClient, "../hack/install.yaml")
		Expect(err).NotTo(HaveOccurred())
	})
})
