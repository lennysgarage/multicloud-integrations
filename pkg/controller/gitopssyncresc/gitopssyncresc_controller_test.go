// Copyright 2021 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitopssyncresc

import (
	"testing"

	"github.com/onsi/gomega"
	appsetreport "open-cluster-management.io/multicloud-integrations/pkg/apis/appsetreport/v1alpha1"
)

func TestCreateOrUpdateAppSetReport(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	synResc := &GitOpsSyncResource{
		Client:      nil,
		Interval:    10,
		ResourceDir: "/tmp",
	}

	appReportsMap := make(map[string]*appsetreport.MulticlusterApplicationSetReport)

	appset1 := make(map[string]interface{})
	appset1["namespace"] = "test-NS1"
	appset1["name"] = "app1"
	appset1["_hostingResource"] = "ApplicationSet/gitops/appset1"
	appset1["apigroup"] = "argoproj.io"
	appset1["apiversion"] = "v1alpha1"
	appset1["_uid"] = "cluster1/abc"
	appset1["_condition.SyncError"] = "something's not right"
	appset1["_condition.SharedResourceWarning"] = "I think it crashed"

	appset1Resources1 := make(map[string]interface{})
	appset1Resources1["kind"] = "Service"
	appset1Resources1["apiversion"] = "v1"
	appset1Resources1["name"] = "welcome-php"
	appset1Resources1["namespace"] = "welcome-waves-and-hooks"
	appset1Resources1["cluster"] = "cluster1"

	appset1Resources2 := make(map[string]interface{})
	appset1Resources2["apigroup"] = "batch"
	appset1Resources2["apiversion"] = "v1"
	appset1Resources2["kind"] = "Job"
	appset1Resources2["name"] = "welcome-presyncjob"
	appset1Resources2["namespace"] = "welcome-waves-and-hooks"
	appset1Resources2["cluster"] = "cluster1"

	appset1Resources3 := make(map[string]interface{})
	appset1Resources3["kind"] = "Deployment"
	appset1Resources3["apiversion"] = "v1"
	appset1Resources3["name"] = "welcome-presyncjob-kcbqk"
	appset1Resources3["namespace"] = "welcome-waves-and-hooks"
	appset1Resources3["cluster"] = "cluster2"

	related1 := make(map[string]interface{})
	related1["kind"] = "Service"
	related1["items"] = []interface{}{appset1Resources1}
	related2 := make(map[string]interface{})
	related2["kind"] = "Job"
	related2["items"] = []interface{}{appset1Resources2}
	related3 := make(map[string]interface{})
	related3["kind"] = "Deployment"
	related3["items"] = []interface{}{appset1Resources3}
	appset1["related"] = []interface{}{related1, related2, related3}

	managedClustersAppNameMap := make(map[string]map[string]string)
	c1ResourceListMap := getResourceMapList(appset1["related"].([]interface{}), "cluster1")
	err := synResc.createOrUpdateAppSetReportConditions(appReportsMap, appset1, "cluster1", managedClustersAppNameMap)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(appReportsMap["gitops_appset1"]).NotTo(gomega.BeNil())
	g.Expect(appReportsMap["gitops_appset1"].GetName()).To(gomega.Equal("gitops_appset1"))
	g.Expect(len(c1ResourceListMap)).To(gomega.Equal(2))
	g.Expect(len(appReportsMap["gitops_appset1"].Statuses.ClusterConditions)).To(gomega.Equal(1))
	g.Expect(appReportsMap["gitops_appset1"].Statuses.ClusterConditions[0].Cluster).To(gomega.Equal("cluster1"))
	g.Expect(len(appReportsMap["gitops_appset1"].Statuses.ClusterConditions[0].Conditions)).To(gomega.Equal(2))
	g.Expect(managedClustersAppNameMap["appset1"]["cluster1"], "test-NS1_app1")

	// Add to same appset from cluster2
	c2ResourceListMap := getResourceMapList(appset1["related"].([]interface{}), "cluster2")
	err = synResc.createOrUpdateAppSetReportConditions(appReportsMap, appset1, "cluster2", managedClustersAppNameMap)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(len(appReportsMap["gitops_appset1"].Statuses.ClusterConditions)).To(gomega.Equal(2))
	g.Expect(len(c2ResourceListMap)).To(gomega.Equal(1))
	g.Expect(appReportsMap["gitops_appset1"].Statuses.ClusterConditions[0].Cluster).To(gomega.Equal("cluster1"))
	g.Expect(len(appReportsMap["gitops_appset1"].Statuses.ClusterConditions[0].Conditions)).To(gomega.Equal(2))
	g.Expect(appReportsMap["gitops_appset1"].Statuses.ClusterConditions[1].Cluster).To(gomega.Equal("cluster2"))
	g.Expect(len(appReportsMap["gitops_appset1"].Statuses.ClusterConditions[1].Conditions)).To(gomega.Equal(2))
	g.Expect(managedClustersAppNameMap["appset1"]["cluster2"], "test-NS1_app1")
}
