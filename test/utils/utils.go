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

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2" //nolint:golint,revive
	syngit_utils "github.com/syngit-org/syngit/pkg/utils"
)

const (
	prometheusOperatorVersion = "v0.68.0"
	prometheusOperatorURL     = "https://github.com/prometheus-operator/prometheus-operator/" +
		"releases/download/%s/bundle.yaml"

	certmanagerVersion = "v1.17.2"
	certmanagerCRDsURL = "https://github.com/cert-manager/cert-manager/releases/download/%s/cert-manager.crds.yaml"

	certmanagerURLTmpl = "https://github.com/jetstack/cert-manager/releases/download/%s/cert-manager.yaml"
)

func warnError(err error) {
	fmt.Fprintf(GinkgoWriter, "warning: %v\n", err) //nolint
}

// InstallPrometheusOperator installs the prometheus Operator to be used to export the enabled metrics.
func InstallPrometheusOperator() error {
	url := fmt.Sprintf(prometheusOperatorURL, prometheusOperatorVersion)
	cmd := exec.Command("kubectl", "create", "-f", url)
	_, err := Run(cmd)
	return err
}

// Run executes the provided command within this context
func Run(cmd *exec.Cmd) ([]byte, error) {
	dir, _ := GetProjectDir()
	cmd.Dir = dir

	if err := os.Chdir(cmd.Dir); err != nil {
		fmt.Fprintf(GinkgoWriter, "chdir dir: %s\n", err) //nolint
	}

	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	command := strings.Join(cmd.Args, " ")
	fmt.Fprintf(GinkgoWriter, "running: %s\n", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("%s failed with error: (%v) %s", command, err, string(output))
	}

	return output, nil
}

// UninstallPrometheusOperator uninstalls the prometheus
func UninstallPrometheusOperator() {
	url := fmt.Sprintf(prometheusOperatorURL, prometheusOperatorVersion)
	cmd := exec.Command("kubectl", "delete", "-f", url)
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
}

// UninstallCertManager uninstalls the cert manager

func UninstallCertManager() {
	url := fmt.Sprintf(certmanagerURLTmpl, certmanagerVersion)
	cmd := exec.Command("kubectl", "delete", "-f", url)
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
	cmd = exec.Command("helm", "uninstall", "-n", "cert-manager", "cert-manager")
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
}

func UninstallCertManagerCRDs() {
	url := fmt.Sprintf(certmanagerCRDsURL, certmanagerVersion)
	cmd := exec.Command("kubectl", "delete", "-f", url)
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
}

// InstallCertManager installs the cert manager bundle.
// func InstallCertManager() error {
// 	url := fmt.Sprintf(certmanagerURLTmpl, certmanagerVersion)
// 	cmd := exec.Command("kubectl", "apply", "--force", "-f", url)
// 	if _, err := Run(cmd); err != nil {
// 		return err
// 	}
// 	// Wait for cert-manager-webhook to be ready, which can take time if cert-manager
// 	// was re-installed after uninstalling on a cluster.
// 	cmd = exec.Command("kubectl", "wait", "deployment.apps/cert-manager-webhook",
// 		"--for", "condition=Available",
// 		"--namespace", "cert-manager",
// 		"--timeout", "5m",
// 	)

// 	_, err := Run(cmd)
// 	if err != nil {
// 		return err
// 	}

// 	return err
// }

// TO DO: delete following function when last stable version of syngit Helm chat is >= 0.4.8
func InstallCertManager() error {
	cmd := exec.Command("helm", "repo", "add", "jetstack", "https://charts.jetstack.io")
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}

	cmd = exec.Command("helm", "install", "cert-manager", "-n", "cert-manager", "--version", "v1.16.2", "--create-namespace", "jetstack/cert-manager", "--set", "installCRDs=true")
	if _, err := Run(cmd); err != nil {
		return err
	}
	// Wait for cert-manager-webhook to be ready, which can take time if cert-manager
	// was re-installed after uninstalling on a cluster.
	cmd = exec.Command("kubectl", "wait", "deployment.apps/cert-manager-webhook",
		"--for", "condition=Available",
		"--namespace", "cert-manager",
		"--timeout", "5m",
	)

	_, err := Run(cmd)
	if err != nil {
		return err
	}

	return err
}

func InstallCertManagerCRDs() error {
	url := fmt.Sprintf(certmanagerCRDsURL, certmanagerVersion)
	cmd := exec.Command("kubectl", "apply", "-f", url)
	_, err := Run(cmd)
	if err != nil {
		warnError(err)
	}

	return err
}

// LoadImageToKindCluster loads a local docker image to the kind cluster
func LoadImageToKindClusterWithName(name string) error {
	cluster := "syngit-dev-cluster"
	if v, ok := os.LookupEnv("KIND_CLUSTER"); ok {
		cluster = v
	}
	kindOptions := []string{"load", "docker-image", name, "--name", cluster}
	cmd := exec.Command("kind", kindOptions...)
	_, err := Run(cmd)
	return err
}

// GetNonEmptyLines converts given command output string into individual objects
// according to line breakers, and ignores the empty elements in it.
func GetNonEmptyLines(output string) []string {
	var res []string
	elements := strings.Split(output, "\n")
	for _, element := range elements {
		if element != "" {
			res = append(res, element)
		}
	}

	return res
}

// GetProjectDir will return the directory where the project is err="secrets \"cert-manager-webhook-ca\" already exists"
func GetProjectDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return wd, err
	}
	wd = strings.Split(wd, "/test/e2e")[0]
	return wd, nil
}

func SanitizeUsername(username string) string {
	return syngit_utils.Sanitize(username)
}
