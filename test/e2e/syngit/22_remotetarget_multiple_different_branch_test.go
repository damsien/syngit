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
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	syngit "github.com/syngit-org/syngit/pkg/api/v1beta3"
	"github.com/syngit-org/syngit/pkg/utils"
	. "github.com/syngit-org/syngit/test/utils"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("22 RemoteTarget multiple different branch", func() {
	ctx := context.TODO()

	const (
		remoteUserLuffyName = "remoteuser-luffy"
		remoteSyncerName1   = "remotesyncer-test22.1"
		remoteSyncerName2   = "remotesyncer-test22.2"
		cmName1             = "test-cm22.1"
		cmName2             = "test-cm22.2"
		upstreamBranch      = "main"
		branch1             = "different-branch1"
		branch2             = "different-branch2"
		branch3             = "different-branch3"
	)

	It("should push the ConfigMap to all the branches and the Luffy's branch", func() {

		repoUrl := fmt.Sprintf("https://%s/%s/%s.git", gitP1Fqdn, giteaBaseNs, repo1)
		targetBranch := string(Luffy)

		By("creating the RemoteUser for Luffy")
		luffySecretName := string(Luffy) + "-creds"
		remoteUserLuffy := &syngit.RemoteUser{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteUserLuffyName,
				Namespace: namespace,
				Annotations: map[string]string{
					syngit.RubAnnotationKeyManaged: "true",
				},
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

		By("creating the RemoteSyncer")
		branches := []string{branch1, branch2, branch3}
		remotesyncer := &syngit.RemoteSyncer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteSyncerName1,
				Namespace: namespace,
				Annotations: map[string]string{
					syngit.RtAnnotationKeyUserSpecific:      string(syngit.RtAnnotationValueOneUserOneBranch),
					syngit.RtAnnotationKeyOneOrManyBranches: strings.Join(branches, ","),
				},
			},
			Spec: syngit.RemoteSyncerSpec{
				InsecureSkipTlsVerify:       true,
				DefaultBranch:               upstreamBranch,
				DefaultUnauthorizedUserMode: syngit.Block,
				Strategy:                    syngit.CommitApply,
				TargetStrategy:              syngit.MultipleTarget,
				RemoteRepository:            repoUrl,
				ScopedResources: syngit.ScopedResources{
					Rules: []admissionv1.RuleWithOperations{{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
						},
						Rule: admissionv1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"configmaps"},
						},
					},
					},
				},
			},
		}
		Eventually(func() bool {
			err := sClient.As(Luffy).CreateOrUpdate(remotesyncer)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		By("creating a test configmap")
		Wait3()
		cm := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: cmName1, Namespace: namespace},
			Data:       map[string]string{"test": "oui"},
		}
		Eventually(func() bool {
			_, err := sClient.KAs(Luffy).CoreV1().ConfigMaps(namespace).Create(ctx,
				cm,
				metav1.CreateOptions{},
			)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		By("checking that the configmap is present in all the branches")
		Wait3()
		repo := &Repo{
			Fqdn:   gitP1Fqdn,
			Owner:  giteaBaseNs,
			Name:   repo1,
			Branch: targetBranch,
		}
		exists, err := IsObjectInRepo(*repo, cm)
		Expect(err).ToNot(HaveOccurred())
		Expect(exists).To(BeTrue())
		for _, branch := range branches {
			repo := &Repo{
				Fqdn:   gitP1Fqdn,
				Owner:  giteaBaseNs,
				Name:   repo1,
				Branch: branch,
			}
			exists, err := IsObjectInRepo(*repo, cm)
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeTrue())
		}

		By("checking that the configmap is present on the cluster")
		nnCm := types.NamespacedName{
			Name:      cmName1,
			Namespace: namespace,
		}
		getCm := &corev1.ConfigMap{}

		Eventually(func() bool {
			err := sClient.As(Luffy).Get(nnCm, getCm)
			return err == nil
		}, timeout, interval).Should(BeTrue())

	})

	It("should deny the request because of the wrong strategy", func() {

		repoUrl := fmt.Sprintf("https://%s/%s/%s.git", gitP1Fqdn, giteaBaseNs, repo1)
		targetBranch := string(Luffy)

		By("creating the RemoteUser for Luffy")
		luffySecretName := string(Luffy) + "-creds"
		remoteUserLuffy := &syngit.RemoteUser{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteUserLuffyName,
				Namespace: namespace,
				Annotations: map[string]string{
					syngit.RubAnnotationKeyManaged: "true",
				},
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

		By("creating the RemoteSyncer")
		branches := []string{branch1, branch2, branch3}
		remotesyncer := &syngit.RemoteSyncer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteSyncerName2,
				Namespace: namespace,
				Annotations: map[string]string{
					syngit.RtAnnotationKeyUserSpecific:      string(syngit.RtAnnotationValueOneUserOneBranch),
					syngit.RtAnnotationKeyOneOrManyBranches: strings.Join(branches, ","),
				},
			},
			Spec: syngit.RemoteSyncerSpec{
				InsecureSkipTlsVerify:       true,
				DefaultBranch:               upstreamBranch,
				DefaultUnauthorizedUserMode: syngit.Block,
				Strategy:                    syngit.CommitApply,
				TargetStrategy:              syngit.OneTarget,
				RemoteRepository:            repoUrl,
				ScopedResources: syngit.ScopedResources{
					Rules: []admissionv1.RuleWithOperations{{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
						},
						Rule: admissionv1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"configmaps"},
						},
					},
					},
				},
			},
		}
		Eventually(func() bool {
			err := sClient.As(Luffy).CreateOrUpdate(remotesyncer)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		By("creating a test configmap")
		Wait3()
		cm := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: cmName2, Namespace: namespace},
			Data:       map[string]string{"test": "oui"},
		}
		Eventually(func() bool {
			_, err := sClient.KAs(Luffy).CoreV1().ConfigMaps(namespace).Create(ctx,
				cm,
				metav1.CreateOptions{},
			)
			return err != nil && utils.ErrorTypeChecker(&utils.MultipleTargetError{}, err.Error())
		}, timeout, interval).Should(BeTrue())

		By("checking that the configmap is not present on the branches")
		Wait3()
		repo := &Repo{
			Fqdn:   gitP1Fqdn,
			Owner:  giteaBaseNs,
			Name:   repo1,
			Branch: targetBranch,
		}
		exists, err := IsObjectInRepo(*repo, cm)
		Expect(err).To(HaveOccurred())
		Expect(exists).To(BeFalse())
		for _, branch := range branches {
			repo := &Repo{
				Fqdn:   gitP1Fqdn,
				Owner:  giteaBaseNs,
				Name:   repo1,
				Branch: branch,
			}
			exists, err := IsObjectInRepo(*repo, cm)
			Expect(err).To(HaveOccurred())
			Expect(exists).To(BeFalse())
		}

		By("checking that the configmap is not present on the cluster")
		nnCm := types.NamespacedName{
			Name:      cmName1,
			Namespace: namespace,
		}
		getCm := &corev1.ConfigMap{}
		Eventually(func() bool {
			err := sClient.As(Luffy).Get(nnCm, getCm)
			return err != nil && strings.Contains(err.Error(), notPresentOnCluser)
		}, timeout, interval).Should(BeTrue())

	})
})
