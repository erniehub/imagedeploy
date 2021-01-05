package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	appsV1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWorkerDeploymentTemplate(t *testing.T) {
	for _, tc := range []struct {
		CaseName string
		Release  string
		Values   map[string]string

		ExpectedErrorRegexp *regexp.Regexp

		ExpectedName        string
		ExpectedRelease     string
		ExpectedDeployments []workerDeploymentTestCase
	}{
		{
			CaseName: "happy",
			Release:  "production",
			Values: map[string]string{
				"releaseOverride":            "productionOverridden",
				"workers.worker1.command[0]": "echo",
				"workers.worker1.command[1]": "worker1",
				"workers.worker2.command[0]": "echo",
				"workers.worker2.command[1]": "worker2",
			},
			ExpectedName:    "productionOverridden",
			ExpectedRelease: "production",
			ExpectedDeployments: []workerDeploymentTestCase{
				{
					ExpectedName:         "productionOverridden-worker1",
					ExpectedCmd:          []string{"echo", "worker1"},
					ExpectedStrategyType: appsV1.DeploymentStrategyType(""),
				},
				{
					ExpectedName:         "productionOverridden-worker2",
					ExpectedCmd:          []string{"echo", "worker2"},
					ExpectedStrategyType: appsV1.DeploymentStrategyType(""),
				},
			},
		}, {
			// See https://github.com/helm/helm/issues/6006
			CaseName: "long release name",
			Release:  strings.Repeat("r", 80),

			ExpectedErrorRegexp: regexp.MustCompile("Error: release name .* exceeds max length of 53"),
		},
		{
			CaseName: "strategyType",
			Release:  "production",
			Values: map[string]string{
				"workers.worker1.command[0]":   "echo",
				"workers.worker1.command[1]":   "worker1",
				"workers.worker1.strategyType": "Recreate",
			},
			ExpectedName:    "production",
			ExpectedRelease: "production",
			ExpectedDeployments: []workerDeploymentTestCase{
				{
					ExpectedName:         "production" + "-worker1",
					ExpectedCmd:          []string{"echo", "worker1"},
					ExpectedStrategyType: appsV1.RecreateDeploymentStrategyType,
				},
			},
		},
	} {
		t.Run(tc.CaseName, func(t *testing.T) {
			namespaceName := "minimal-ruby-app-" + strings.ToLower(random.UniqueId())

			values := map[string]string{
				"gitlab.app": "auto-devops-examples/minimal-ruby-app",
				"gitlab.env": "prod",
			}

			mergeStringMap(values, tc.Values)

			options := &helm.Options{
				SetValues:      values,
				KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
			}

			output, err := helm.RenderTemplateE(t, options, helmChartPath, tc.Release, []string{"templates/worker-deployment.yaml"})

			if tc.ExpectedErrorRegexp != nil {
				require.Regexp(t, tc.ExpectedErrorRegexp, err.Error())
				return
			}
			if err != nil {
				t.Error(err)
				return
			}

			var deployments deploymentList
			helm.UnmarshalK8SYaml(t, output, &deployments)

			require.Len(t, deployments.Items, len(tc.ExpectedDeployments))
			for i, expectedDeployment := range tc.ExpectedDeployments {
				deployment := deployments.Items[i]

				require.Equal(t, expectedDeployment.ExpectedName, deployment.Name)
				require.Equal(t, expectedDeployment.ExpectedStrategyType, deployment.Spec.Strategy.Type)

				require.Equal(t, map[string]string{
					"app.gitlab.com/app": "auto-devops-examples/minimal-ruby-app",
					"app.gitlab.com/env": "prod",
				}, deployment.Annotations)
				require.Equal(t, map[string]string{
					"chart":    chartName,
					"heritage": "Helm",
					"release":  tc.ExpectedRelease,
					"tier":     "worker",
					"track":    "stable",
				}, deployment.Labels)

				require.Equal(t, map[string]string{
					"app.gitlab.com/app":           "auto-devops-examples/minimal-ruby-app",
					"app.gitlab.com/env":           "prod",
					"checksum/application-secrets": "",
				}, deployment.Spec.Template.Annotations)
				require.Equal(t, map[string]string{
					"release": tc.ExpectedRelease,
					"tier":    "worker",
					"track":   "stable",
				}, deployment.Spec.Template.Labels)

				require.Len(t, deployment.Spec.Template.Spec.Containers, 1)
				require.Equal(t, expectedDeployment.ExpectedCmd, deployment.Spec.Template.Spec.Containers[0].Command)
			}
		})
	}

	// Tests worker selector
	for _, tc := range []struct {
		CaseName string
		Release  string
		Values   map[string]string

		ExpectedName        string
		ExpectedRelease     string
		ExpectedDeployments []workerDeploymentSelectorTestCase
	}{
		{
			CaseName: "worker selector",
			Release:  "production",
			Values: map[string]string{
				"workers.worker1.command[0]": "echo",
				"workers.worker1.command[1]": "worker1",
				"workers.worker2.command[0]": "echo",
				"workers.worker2.command[1]": "worker2",
			},
			ExpectedName:    "production",
			ExpectedRelease: "production",
			ExpectedDeployments: []workerDeploymentSelectorTestCase{
				{
					ExpectedName: "production-worker1",
					ExpectedSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"release": "production",
							"tier":    "worker",
							"track":   "stable",
						},
					},
				},
				{
					ExpectedName: "production-worker2",
					ExpectedSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"release": "production",
							"tier":    "worker",
							"track":   "stable",
						},
					},
				},
			},
		},
	} {
		t.Run(tc.CaseName, func(t *testing.T) {
			namespaceName := "minimal-ruby-app-" + strings.ToLower(random.UniqueId())

			values := map[string]string{
				"gitlab.app": "auto-devops-examples/minimal-ruby-app",
				"gitlab.env": "prod",
			}

			mergeStringMap(values, tc.Values)

			options := &helm.Options{
				SetValues:      values,
				KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
			}

			output := helm.RenderTemplate(t, options, helmChartPath, tc.Release, []string{"templates/worker-deployment.yaml"})

			var deployments deploymentAppsV1List
			helm.UnmarshalK8SYaml(t, output, &deployments)

			require.Len(t, deployments.Items, len(tc.ExpectedDeployments))
			for i, expectedDeployment := range tc.ExpectedDeployments {
				deployment := deployments.Items[i]

				require.Equal(t, expectedDeployment.ExpectedName, deployment.Name)

				require.Equal(t, map[string]string{
					"chart":    chartName,
					"heritage": "Helm",
					"release":  tc.ExpectedRelease,
					"tier":     "worker",
					"track":    "stable",
				}, deployment.Labels)

				require.Equal(t, expectedDeployment.ExpectedSelector, deployment.Spec.Selector)

				require.Equal(t, map[string]string{
					"release": tc.ExpectedRelease,
					"tier":    "worker",
					"track":   "stable",
				}, deployment.Spec.Template.Labels)
			}
		})
	}

	// worker livenessProbe, and readinessProbe tests
	for _, tc := range []struct {
		CaseName string
		Values   map[string]string
		Release  string

		ExpectedDeployments []workerDeploymentTestCase
	}{
		{
			CaseName: "default liveness and readiness values",
			Release:  "production",
			Values: map[string]string{
				"workers.worker1.command[0]": "echo",
				"workers.worker1.command[1]": "worker1",
				"workers.worker2.command[0]": "echo",
				"workers.worker2.command[1]": "worker2",
			},
			ExpectedDeployments: []workerDeploymentTestCase{
				{
					ExpectedName:           "production-worker1",
					ExpectedCmd:            []string{"echo", "worker1"},
					ExpectedLivenessProbe:  defaultLivenessProbe(),
					ExpectedReadinessProbe: defaultReadinessProbe(),
				},
				{
					ExpectedName:           "production-worker2",
					ExpectedCmd:            []string{"echo", "worker2"},
					ExpectedLivenessProbe:  defaultLivenessProbe(),
					ExpectedReadinessProbe: defaultReadinessProbe(),
				},
			},
		},
		{
			CaseName: "enableWorkerLivenessProbe",
			Release:  "production",
			Values: map[string]string{
				"workers.worker1.command[0]":              "echo",
				"workers.worker1.command[1]":              "worker1",
				"workers.worker1.livenessProbe.path":      "/worker",
				"workers.worker1.livenessProbe.scheme":    "HTTP",
				"workers.worker1.livenessProbe.probeType": "httpGet",
				"workers.worker2.command[0]":              "echo",
				"workers.worker2.command[1]":              "worker2",
				"workers.worker2.livenessProbe.path":      "/worker",
				"workers.worker2.livenessProbe.scheme":    "HTTP",
				"workers.worker2.livenessProbe.probeType": "httpGet",
			},
			ExpectedDeployments: []workerDeploymentTestCase{
				{
					ExpectedName:           "production-worker1",
					ExpectedCmd:            []string{"echo", "worker1"},
					ExpectedLivenessProbe:  workerLivenessProbe(),
					ExpectedReadinessProbe: defaultReadinessProbe(),
				},
				{
					ExpectedName:           "production-worker2",
					ExpectedCmd:            []string{"echo", "worker2"},
					ExpectedLivenessProbe:  workerLivenessProbe(),
					ExpectedReadinessProbe: defaultReadinessProbe(),
				},
			},
		},
		{
			CaseName: "enableWorkerReadinessProbe",
			Release:  "production",
			Values: map[string]string{
				"workers.worker1.command[0]":               "echo",
				"workers.worker1.command[1]":               "worker1",
				"workers.worker1.readinessProbe.path":      "/worker",
				"workers.worker1.readinessProbe.scheme":    "HTTP",
				"workers.worker1.readinessProbe.probeType": "httpGet",
				"workers.worker2.command[0]":               "echo",
				"workers.worker2.command[1]":               "worker2",
				"workers.worker2.readinessProbe.path":      "/worker",
				"workers.worker2.readinessProbe.scheme":    "HTTP",
				"workers.worker2.readinessProbe.probeType": "httpGet",
			},
			ExpectedDeployments: []workerDeploymentTestCase{
				{
					ExpectedName:           "production-worker1",
					ExpectedCmd:            []string{"echo", "worker1"},
					ExpectedLivenessProbe:  defaultLivenessProbe(),
					ExpectedReadinessProbe: workerReadinessProbe(),
				},
				{
					ExpectedName:           "production-worker2",
					ExpectedCmd:            []string{"echo", "worker2"},
					ExpectedLivenessProbe:  defaultLivenessProbe(),
					ExpectedReadinessProbe: workerReadinessProbe(),
				},
			},
		},
	} {
		t.Run(tc.CaseName, func(t *testing.T) {
			namespaceName := "minimal-ruby-app-" + strings.ToLower(random.UniqueId())

			values := map[string]string{
				"gitlab.app": "auto-devops-examples/minimal-ruby-app",
				"gitlab.env": "prod",
			}

			mergeStringMap(values, tc.Values)

			options := &helm.Options{
				SetValues:      values,
				KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
			}

			output := helm.RenderTemplate(t, options, helmChartPath, tc.Release, []string{"templates/worker-deployment.yaml"})

			var deployments deploymentAppsV1List
			helm.UnmarshalK8SYaml(t, output, &deployments)

			require.Len(t, deployments.Items, len(tc.ExpectedDeployments))

			for i, expectedDeployment := range tc.ExpectedDeployments {
				deployment := deployments.Items[i]
				require.Equal(t, expectedDeployment.ExpectedName, deployment.Name)
				require.Len(t, deployment.Spec.Template.Spec.Containers, 1)
				require.Equal(t, expectedDeployment.ExpectedCmd, deployment.Spec.Template.Spec.Containers[0].Command)
				require.Equal(t, expectedDeployment.ExpectedLivenessProbe, deployment.Spec.Template.Spec.Containers[0].LivenessProbe)
				require.Equal(t, expectedDeployment.ExpectedReadinessProbe, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe)
			}
		})
	}
}
