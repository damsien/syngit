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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	syngit "github.com/syngit-org/syngit/pkg/api/v1beta3"
	"github.com/syngit-org/syngit/pkg/utils"
	. "github.com/syngit-org/syngit/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("19 RemoteTarget same repo & branch between target and upstream test", func() {

	const (
		remoteTargetName = "remotetarget-test19"
		repo             = "https://my-server.com/my-upstream-repo.git"
		differentRepo    = "https://my-server.com/my-different-repo.git"
		branch           = "main"
		differentBranch  = "different"
	)

	It("should deny the RemoteTarget creation", func() {
		By("creating a RemoteTarget with the same repo & branch for the target & upstream and strategy setup")
		remoteTarget := &syngit.RemoteTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteTargetName,
				Namespace: namespace,
			},
			Spec: syngit.RemoteTargetSpec{
				UpstreamRepository: repo,
				TargetRepository:   repo,
				UpstreamBranch:     branch,
				TargetBranch:       branch,
				MergeStrategy:      syngit.TryFastForwardOrHardReset,
			},
		}
		Eventually(func() bool {
			err := sClient.As(Luffy).CreateOrUpdate(remoteTarget)
			return err != nil && utils.ErrorTypeChecker(&utils.SameUpstreamDifferentMergeStrategyError{}, err.Error())
		}, timeout, interval).Should(BeTrue())

	})

	It("should deny the RemoteTarget creation", func() {
		By("creating a RemoteTarget with different repo & branch for the target & upstream and empty strategy")
		remoteTarget := &syngit.RemoteTarget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      remoteTargetName,
				Namespace: namespace,
			},
			Spec: syngit.RemoteTargetSpec{
				UpstreamRepository: repo,
				TargetRepository:   differentRepo,
				UpstreamBranch:     branch,
				TargetBranch:       differentBranch,
				MergeStrategy:      "",
			},
		}
		Eventually(func() bool {
			err := sClient.As(Luffy).CreateOrUpdate(remoteTarget)
			return err != nil && utils.ErrorTypeChecker(&utils.DifferentUpstreamEmptyMergeStrategyError{}, err.Error())
		}, timeout, interval).Should(BeTrue())

	})
})
