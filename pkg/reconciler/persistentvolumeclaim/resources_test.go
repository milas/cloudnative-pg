/*
Copyright The CloudNativePG Contributors

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

package persistentvolumeclaim

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	utils2 "github.com/cloudnative-pg/cloudnative-pg/pkg/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PVC detection", func() {
	It("will list PVCs with Jobs or Pods or which are Ready", func() {
		clusterName := "myCluster"
		makeClusterPVC := func(serial string, isResizing bool) corev1.PersistentVolumeClaim {
			return makePVC(clusterName, serial, utils2.PVCRolePgData, isResizing)
		}
		pvcs := []corev1.PersistentVolumeClaim{
			makeClusterPVC("1", false), // has a Pod
			makeClusterPVC("2", false), // has a Job
			makeClusterPVC("3", true),  // resizing
			makeClusterPVC("4", false), // dangling
		}
		cluster := &apiv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: clusterName,
			},
		}
		EnrichStatus(
			context.TODO(),
			cluster,
			[]corev1.Pod{
				makePod(clusterName, "1"),
				makePod(clusterName, "3"),
			},
			[]batchv1.Job{makeJob(clusterName, "2")},
			pvcs,
		)

		Expect(cluster.Status.PVCCount).Should(BeEquivalentTo(4))
		Expect(cluster.Status.InstanceNames).Should(Equal([]string{
			clusterName + "-1",
			clusterName + "-2",
			clusterName + "-3",
			clusterName + "-4",
		}))
		Expect(cluster.Status.InitializingPVC).Should(Equal([]string{
			clusterName + "-2",
		}))
		Expect(cluster.Status.ResizingPVC).Should(Equal([]string{
			clusterName + "-3",
		}))
		Expect(cluster.Status.DanglingPVC).Should(Equal([]string{
			clusterName + "-4",
		}))
		Expect(cluster.Status.HealthyPVC).Should(Equal([]string{
			clusterName + "-1",
		}))
		Expect(cluster.Status.UnusablePVC).Should(BeEmpty())
	})
})
