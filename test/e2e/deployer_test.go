package e2e_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var _ = Describe("DRA Deployer E2E", func() {
	var (
		k8sClient    client.Client
		ctx          context.Context
		namespace    string
		image        string
		command      string
		deployerBin  string
		projectRoot  string
		deployOutput []byte
		deployError  error
	)

	// BeforeEach: Configuration setup only
	BeforeEach(func() {
		ctx = context.Background()

		// Get configuration from environment or use defaults
		namespace = getEnv("NAMESPACE", "dra-deployer")
		image = getEnv("IMAGE", "quay.io/fromani/dra-driver-memory:v0.0.2025112401")
		command = getEnv("COMMAND", "/bin/dramemory")

		// Get project root and deployer binary path
		var err error
		projectRoot, err = getProjectRoot()
		Expect(err).NotTo(HaveOccurred(), "Failed to get project root")

		deployerBin = filepath.Join(projectRoot, "bin", "dra-deployer")
		Expect(deployerBin).To(BeAnExistingFile(), "Deployer binary not found. Please run 'make build' first")

		// Set up controller-runtime Kubernetes client
		k8sClient, err = getKubeClient()
		Expect(err).NotTo(HaveOccurred(), "Failed to create Kubernetes client")
	})

	// JustBeforeEach: Deployment execution
	JustBeforeEach(func() {
		By("Deploying the DRA memory plugin")
		cmd := exec.Command(deployerBin, "apply", "-i", image, "--command", command)
		deployOutput, deployError = cmd.CombinedOutput()
	})

	Describe("Deploying DRA memory plugin", func() {
		It("should successfully deploy and verify all resources", func() {
			// Verify deployment succeeded
			Expect(deployError).NotTo(HaveOccurred(), fmt.Sprintf("Failed to deploy DRA plugin: %s", string(deployOutput)))

			By("Waiting for namespace to exist")
			ns := &corev1.Namespace{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{Name: namespace}, ns)
			}, 30*time.Second, 1*time.Second).Should(Succeed(), "Namespace should be created")

			By("Verifying ServiceAccount exists")
			Eventually(func() error {
				saList := &corev1.ServiceAccountList{}
				if err := k8sClient.List(ctx, saList, client.InNamespace(namespace)); err != nil {
					return err
				}
				for _, sa := range saList.Items {
					if matchesResource(sa.Name, "dra-driver-memory") {
						return nil
					}
				}
				return fmt.Errorf("ServiceAccount not found")
			}, 30*time.Second, 1*time.Second).Should(Succeed(), "ServiceAccount should exist")

			By("Verifying ClusterRole exists")
			Eventually(func() error {
				crList := &rbacv1.ClusterRoleList{}
				if err := k8sClient.List(ctx, crList); err != nil {
					return err
				}
				for _, cr := range crList.Items {
					if matchesResource(cr.Name, "dra-driver-memory") {
						return nil
					}
				}
				return fmt.Errorf("ClusterRole not found")
			}, 30*time.Second, 1*time.Second).Should(Succeed(), "ClusterRole should exist")

			By("Verifying ClusterRoleBinding exists")
			Eventually(func() error {
				crbList := &rbacv1.ClusterRoleBindingList{}
				if err := k8sClient.List(ctx, crbList); err != nil {
					return err
				}
				for _, crb := range crbList.Items {
					if matchesResource(crb.Name, "dra-driver-memory") {
						return nil
					}
				}
				return fmt.Errorf("ClusterRoleBinding not found")
			}, 30*time.Second, 1*time.Second).Should(Succeed(), "ClusterRoleBinding should exist")

			By("Verifying DaemonSet exists")
			var daemonSetName string
			Eventually(func() error {
				dsList := &appsv1.DaemonSetList{}
				if err := k8sClient.List(ctx, dsList, client.InNamespace(namespace)); err != nil {
					return err
				}
				for _, ds := range dsList.Items {
					if matchesResource(ds.Name, "dra-driver-memory") {
						daemonSetName = ds.Name
						return nil
					}
				}
				return fmt.Errorf("DaemonSet not found")
			}, 30*time.Second, 1*time.Second).Should(Succeed(), "DaemonSet should exist")

			By("Waiting for DaemonSet pods to be ready")
			err := wait.PollUntilContextTimeout(ctx, 5*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
				ds := &appsv1.DaemonSet{}
				if err := k8sClient.Get(ctx, types.NamespacedName{Name: daemonSetName, Namespace: namespace}, ds); err != nil {
					return false, err
				}

				// Check if all desired pods are ready
				if ds.Status.NumberReady > 0 && ds.Status.NumberReady == ds.Status.DesiredNumberScheduled {
					return true, nil
				}

				GinkgoWriter.Printf("DaemonSet status: Desired=%d, Ready=%d, Available=%d\n",
					ds.Status.DesiredNumberScheduled,
					ds.Status.NumberReady,
					ds.Status.NumberAvailable)

				return false, nil
			})
			Expect(err).NotTo(HaveOccurred(), "DaemonSet pods should become ready")

			By("Verifying pods are running")
			podList := &corev1.PodList{}
			err = k8sClient.List(ctx, podList,
				client.InNamespace(namespace),
				client.MatchingLabels{"app.kubernetes.io/name": "dra-driver-memory"})
			Expect(err).NotTo(HaveOccurred(), "Failed to list pods")
			Expect(podList.Items).NotTo(BeEmpty(), "At least one pod should be running")

			for _, pod := range podList.Items {
				Expect(pod.Status.Phase).To(Equal(corev1.PodRunning),
					fmt.Sprintf("Pod %s should be in Running state", pod.Name))
			}

			By("Displaying pod logs for debugging")
			for _, pod := range podList.Items {
				GinkgoWriter.Printf("\n=== Pod %s in namespace %s ===\n", pod.Name, pod.Namespace)
			}
		})
	})
})

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Go up directories until we find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (go.mod)")
		}
		dir = parent
	}
}

func getKubeClient() (client.Client, error) {
	// Get config from default locations or KUBECONFIG env var
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	// Create controller-runtime client
	return client.New(cfg, client.Options{})
}

func matchesResource(name, substring string) bool {
	return len(name) > 0 && (name == substring || strings.Contains(name, substring))
}
