/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e_helm_install

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/syngit-org/syngit/test/utils"
)

const (
	testNamespace = "test"
)

// Run e2e tests using the Ginkgo runner.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	fmt.Fprintf(GinkgoWriter, "Starting syngit helm install suite\n")
	RunSpecs(t, "e2e suite")
}

var _ = BeforeSuite(func() {

	By("creating test namespace")
	cmd := exec.Command("kubectl", "create", "ns", testNamespace)
	_, err := utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("installing prometheus operator")
	Expect(utils.InstallPrometheusOperator()).To(Succeed())

	By("installing the cert-manager")
	Expect(utils.InstallCertManager()).To(Succeed())

	By("build the image")
	cmd = exec.Command("make", "docker-build")
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("load the image in the KinD cluster")
	cmd = exec.Command("make", "kind-load-image")
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("installing the syngit chart")
	cmd = exec.Command("make", "chart-install-providers")
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

})

var _ = AfterSuite(func() {

	By("uninstalling the Prometheus manager bundle")
	utils.UninstallPrometheusOperator()

	By("uninstalling the cert-manager bundle")
	utils.UninstallCertManager()

	By("uninstalling the syngit chart")
	cmd := exec.Command("make", "chart-uninstall")
	_, err := utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("removing test namespace")
	cmd = exec.Command("kubectl", "delete", "ns", testNamespace)
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	By("removing syngit namespace")
	cmd = exec.Command("kubectl", "delete", "ns", "syngit")
	_, err = utils.Run(cmd)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

})

// Wait5 sleeps for 5 seconds
func Wait5() {
	time.Sleep(5 * time.Second)
}
