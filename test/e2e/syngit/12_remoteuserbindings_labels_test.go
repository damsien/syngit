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

package e2e_syngit

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	syngit "github.com/syngit-org/syngit/pkg/api/v1beta2"
	. "github.com/syngit-org/syngit/test/utils"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("12 RemoteUserBinding labels", func() {

	const (
		remoteUserLuffyName        = "remoteuser-luffy"
		remoteUserBindingLuffyName = "remoteuserbinding-luffy"
	)

	It("should instantiate the RemoteUserBinding correctly (without labels)", func() {
		By("adding syngit to scheme")
		err := syngit.AddToScheme(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

		By("creating the RemoteUser for Luffy")
		luffySecretName := string(Luffy) + "-creds"
		remoteUserLuffy := &syngit.RemoteUser{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteUserLuffyName,
				Namespace: namespace,
			},
			Spec: syngit.RemoteUserSpec{
				Email:             "sample@email.com",
				GitBaseDomainFQDN: gitP1Fqdn,
				SecretRef: corev1.SecretReference{
					Name: luffySecretName,
				},
			},
		}
		Eventually(func() bool {
			err := sClient.As(Luffy).CreateOrUpdate(remoteUserLuffy)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		By("creating the RemoteUserBinding")
		remoteuserbinding := &syngit.RemoteUserBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteUserBindingLuffyName,
				Namespace: namespace,
			},
			Spec: syngit.RemoteUserBindingSpec{
				RemoteRefs: []corev1.ObjectReference{{Name: remoteUserLuffyName}},
				Subject: rbacv1.Subject{
					Kind: rbacv1.UserKind,
					Name: "dummyUser",
				},
			},
		}
		Eventually(func() bool {
			err := sClient.As(Luffy).CreateOrUpdate(remoteuserbinding)
			return err == nil
		}, timeout, interval).Should(BeTrue())

	})

	It("should not instantiate the RemoteUserBinding (with label)", func() {
		By("adding syngit to scheme")
		err := syngit.AddToScheme(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

		By("creating the RemoteUser for Luffy")
		luffySecretName := string(Luffy) + "-creds"
		remoteUserLuffy := &syngit.RemoteUser{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteUserLuffyName,
				Namespace: namespace,
			},
			Spec: syngit.RemoteUserSpec{
				Email:             "sample@email.com",
				GitBaseDomainFQDN: gitP1Fqdn,
				SecretRef: corev1.SecretReference{
					Name: luffySecretName,
				},
			},
		}
		Eventually(func() bool {
			err := sClient.As(Luffy).CreateOrUpdate(remoteUserLuffy)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		By("creating the RemoteUserBinding")
		remoteuserbinding := &syngit.RemoteUserBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteUserBindingLuffyName,
				Namespace: namespace,
				Labels: map[string]string{
					"managed-by": "syngit.io",
				},
			},
			Spec: syngit.RemoteUserBindingSpec{
				RemoteRefs: []corev1.ObjectReference{{Name: remoteUserLuffyName}},
				Subject: rbacv1.Subject{
					Kind: rbacv1.UserKind,
					Name: "dummyUser",
				},
			},
		}
		Eventually(func() bool {
			err := sClient.As(Luffy).CreateOrUpdate(remoteuserbinding)
			return err != nil && strings.Contains(err.Error(), rubLabelsDeniedMessage)
		}, timeout, interval).Should(BeTrue())

	})

})
