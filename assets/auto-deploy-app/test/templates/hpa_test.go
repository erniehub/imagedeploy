package main

import (
	"regexp"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	autoscalingV1 "k8s.io/api/autoscaling/v1"
)

func TestHPA_AutoscalingV1(t *testing.T) {
	templates := []string{"templates/hpa.yaml"}
	releaseName := "hpa-test"

	tcs := []struct {
		name   string
		values map[string]string

		expectedName        string
		expectedMinReplicas int32
		expectedMaxReplicas int32
		expectedTargetCPU   int32

		expectedErrorRegexp *regexp.Regexp
	}{
		{
			name:                "defaults",
			expectedErrorRegexp: regexp.MustCompile("Error: could not find template templates/hpa.yaml in chart"),
		},
		{
			name:                "with hpa enabled, no requests",
			values:              map[string]string{"hpa.enabled": "true"},
			expectedErrorRegexp: regexp.MustCompile("Error: could not find template templates/hpa.yaml in chart"),
		},
		{
			name:                "with hpa enabled and requests defined",
			values:              map[string]string{"hpa.enabled": "true", "resources.requests.cpu": "500"},
			expectedName:        "hpa-test-auto-deploy",
			expectedMinReplicas: 1,
			expectedMaxReplicas: 5,
			expectedTargetCPU:   80,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			output, ret := renderTemplate(t, tc.values, releaseName, templates, tc.expectedErrorRegexp)

			if ret == false {
				return
			}

			hpa := new(autoscalingV1.HorizontalPodAutoscaler)
			helm.UnmarshalK8SYaml(t, output, hpa)
			require.Equal(t, tc.expectedName, hpa.ObjectMeta.Name)
			require.Equal(t, tc.expectedMinReplicas, *hpa.Spec.MinReplicas)
			require.Equal(t, tc.expectedMaxReplicas, hpa.Spec.MaxReplicas)
			require.Equal(t, tc.expectedTargetCPU, *hpa.Spec.TargetCPUUtilizationPercentage)
		})
	}
}
