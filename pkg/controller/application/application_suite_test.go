


package application_test

import (
    "testing"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/kubernetes-sigs/kubebuilder/pkg/controller"
    "github.com/kubernetes-sigs/kubebuilder/pkg/inject/run"
    "github.com/kubernetes-sigs/kubebuilder/pkg/test"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"

    "github.com/kubernetes-sigs/application/pkg/client/clientset/versioned"
    "github.com/kubernetes-sigs/application/pkg/inject"
    "github.com/kubernetes-sigs/application/pkg/inject/args"
)

var (
    testenv *test.TestEnvironment
    config *rest.Config
    cs *versioned.Clientset
    ks *kubernetes.Clientset
    shutdown chan struct{}
    ctrl *controller.GenericController
)

func TestBee(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecsWithDefaultAndCustomReporters(t, "Application Suite", []Reporter{test.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
    testenv = &test.TestEnvironment{CRDs: inject.Injector.CRDs}
    var err error
    config, err = testenv.Start()
    Expect(err).NotTo(HaveOccurred())
    cs = versioned.NewForConfigOrDie(config)
    ks = kubernetes.NewForConfigOrDie(config)

    shutdown = make(chan struct{})
    arguments := args.CreateInjectArgs(config)
    go func() {
        defer GinkgoRecover()
        Expect(inject.RunAll(run.RunArguments{Stop: shutdown}, arguments)).
            To(BeNil())
    }()

    // Wait for RunAll to create the controllers and then set the reference
    defer GinkgoRecover()
    Eventually(func() interface{} { return arguments.ControllerManager.GetController("ApplicationController") }).
        Should(Not(BeNil()))
    ctrl = arguments.ControllerManager.GetController("ApplicationController")
})

var _ = AfterSuite(func() {
    close(shutdown)
    testenv.Stop()
})
