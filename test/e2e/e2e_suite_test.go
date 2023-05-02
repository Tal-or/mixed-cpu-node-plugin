package e2e_test

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type TestFixture struct {
	ctx    context.Context
	Cli    client.Client
	K8SCli *kubernetes.Clientset
	NS     *corev1.Namespace
}

var fixture TestFixture

func TestE2e(t *testing.T) {
	BeforeSuite(func() {
		fixture.ctx = context.Background()
		Expect(initClient()).ToNot(HaveOccurred())
		Expect(initK8SClient()).ToNot(HaveOccurred())
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")

}

func initClient() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	fixture.Cli, err = client.New(cfg, client.Options{})
	return err
}

func initK8SClient() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}
	fixture.K8SCli, err = kubernetes.NewForConfig(cfg)
	return err
}

func createNamespace(prefix string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix,
			Labels: map[string]string{
				"security.openshift.io/scc.podSecurityLabelSync": "false",
				"pod-security.kubernetes.io/audit":               "privileged",
				"pod-security.kubernetes.io/enforce":             "privileged",
				"pod-security.kubernetes.io/warn":                "privileged",
			},
		},
	}
	if err := fixture.Cli.Create(context.TODO(), ns); err != nil {
		return err
	}
	fixture.NS = ns
	return nil
}

func deleteNamespace(ns *corev1.Namespace) error {
	err := fixture.Cli.Delete(context.TODO(), ns)
	if err != nil {
		return fmt.Errorf("failed deleting namespace %q; %w", ns.Name, err)
	}

	EventuallyWithOffset(1, func() (bool, error) {
		err = fixture.Cli.Get(fixture.ctx, client.ObjectKeyFromObject(ns), ns)
		if err != nil {
			if !errors.IsNotFound(err) {
				return false, err
			}
			return true, nil
		}
		return false, nil
	}).WithPolling(time.Second*5).WithTimeout(time.Minute*5).Should(BeTrue(), "namespace %q has not been terminated", ns.Name)

	return nil
}

// TODO make it possible to read directly from DaemonSet
func GetSharedCPUs() string {
	cpus, ok := os.LookupEnv("E2E_SHARED_CPUS")
	if !ok {
		return ""
	}
	return cpus
}

func Skipf(format string, a ...any) {
	Skip(fmt.Sprintf(format, a...))
}
