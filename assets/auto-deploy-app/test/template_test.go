package main

import (
	"regexp"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	chartName     = "auto-deploy-app-2.3.0"
	helmChartPath = ".."
)

func TestIngressTemplate_ModSecurity(t *testing.T) {
	templates := []string{"templates/ingress.yaml"}
	modSecuritySnippet := "SecRuleEngine DetectionOnly\n"
	modSecuritySnippetWithSecRules := modSecuritySnippet + `SecRule REQUEST_HEADERS:User-Agent \"scanner\" \"log,deny,id:107,status:403,msg:\'Scanner Identified\'\"
SecRule REQUEST_HEADERS:Content-Type \"text/plain\" \"log,deny,id:\'20010\',status:403,msg:\'Text plain not allowed\'\"
`
	defaultAnnotations := map[string]string{
		"kubernetes.io/ingress.class": "nginx",
		"kubernetes.io/tls-acme":      "true",
	}
	defaultModSecurityAnnotations := map[string]string{
		"nginx.ingress.kubernetes.io/modsecurity-transaction-id": "$server_name-$request_id",
	}
	modSecurityAnnotations := make(map[string]string)
	secRulesAnnotations := make(map[string]string)
	mergeStringMap(modSecurityAnnotations, defaultAnnotations)
	mergeStringMap(modSecurityAnnotations, defaultModSecurityAnnotations)
	mergeStringMap(secRulesAnnotations, defaultAnnotations)
	mergeStringMap(secRulesAnnotations, defaultModSecurityAnnotations)
	modSecurityAnnotations["nginx.ingress.kubernetes.io/modsecurity-snippet"] = modSecuritySnippet
	secRulesAnnotations["nginx.ingress.kubernetes.io/modsecurity-snippet"] = modSecuritySnippetWithSecRules

	tcs := []struct {
		name       string
		valueFiles []string
		values     map[string]string
		meta       metav1.ObjectMeta
	}{
		{
			name: "defaults",
			meta: metav1.ObjectMeta{Annotations: defaultAnnotations},
		},
		{
			name:   "with modSecurity enabled without custom secRules",
			values: map[string]string{"ingress.modSecurity.enabled": "true"},
			meta:   metav1.ObjectMeta{Annotations: modSecurityAnnotations},
		},
		{
			name:       "with custom secRules",
			valueFiles: []string{"./testdata/modsecurity-ingress.yaml"},
			meta:       metav1.ObjectMeta{Annotations: secRulesAnnotations},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			opts := &helm.Options{
				ValuesFiles: tc.valueFiles,
				SetValues:   tc.values,
			}
			output := helm.RenderTemplate(t, opts, helmChartPath, "ModSecurity-test-release", templates)

			ingress := new(extensions.Ingress)
			helm.UnmarshalK8SYaml(t, output, ingress)

			require.Equal(t, tc.meta.Annotations, ingress.ObjectMeta.Annotations)
		})
	}
}

func TestIngressTemplate_DifferentTracks(t *testing.T) {
	templates := []string{"templates/ingress.yaml"}
	tcs := []struct {
		name        string
		releaseName string
		values      map[string]string

		expectedName                     string
		expectedLabels                   map[string]string
		expectedSelector                 map[string]string
		expectedAnnotations              map[string]string
		expectedInexistentAnnotationKeys []string
		expectedErrorRegexp              *regexp.Regexp
	}{
		{
			name:                             "defaults",
			releaseName:                      "production",
			expectedName:                     "production-auto-deploy",
			expectedAnnotations:              map[string]string{"kubernetes.io/ingress.class": "nginx"},
			expectedInexistentAnnotationKeys: []string{"nginx.ingress.kubernetes.io/canary"},
		},
		{
			name:                             "with canary track",
			releaseName:                      "production-canary",
			values:                           map[string]string{"application.track": "canary"},
			expectedName:                     "production-canary-auto-deploy",
			expectedAnnotations:              map[string]string{"nginx.ingress.kubernetes.io/canary": "true", "nginx.ingress.kubernetes.io/canary-by-header": "canary", "kubernetes.io/ingress.class": "nginx"},
			expectedInexistentAnnotationKeys: []string{"nginx.ingress.kubernetes.io/canary-weight"},
		},
		{
			name:                "with canary weight",
			releaseName:         "production-canary",
			values:              map[string]string{"application.track": "canary", "ingress.canary.weight": "25"},
			expectedName:        "production-canary-auto-deploy",
			expectedAnnotations: map[string]string{"nginx.ingress.kubernetes.io/canary-weight": "25"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			output, ret := renderTemplate(t, tc.values, tc.releaseName, templates, tc.expectedErrorRegexp)

			if ret == false {
				return
			}

			ingress := new(extensions.Ingress)
			helm.UnmarshalK8SYaml(t, output, ingress)
			require.Equal(t, tc.expectedName, ingress.ObjectMeta.Name)
			for key, value := range tc.expectedAnnotations {
				require.Equal(t, ingress.ObjectMeta.Annotations[key], value)
			}
			for _, key := range tc.expectedInexistentAnnotationKeys {
				require.Empty(t, ingress.ObjectMeta.Annotations[key])
			}
		})
	}
}

func TestIngressTemplate_TLS(t *testing.T) {
	templates := []string{"templates/ingress.yaml"}
	releaseName := "ingress-tls-test"
	tcs := []struct {
		name   string
		values map[string]string

		expectedAnnotations map[string]string
		expectedIngressTLS  []extensions.IngressTLS
		expectedErrorRegexp *regexp.Regexp
	}{
		{
			name:                "defaults",
			expectedAnnotations: map[string]string{"kubernetes.io/ingress.class": "nginx", "kubernetes.io/tls-acme": "true"},
			expectedIngressTLS: []extensions.IngressTLS{
				extensions.IngressTLS{
					Hosts:      []string{"my.host.com"},
					SecretName: releaseName + "-auto-deploy-tls",
				},
			},
		},
		{
			name:                "with tls-acme disabled",
			values:              map[string]string{"ingress.tls.acme": "false"},
			expectedAnnotations: map[string]string{"kubernetes.io/ingress.class": "nginx", "kubernetes.io/tls-acme": "false"},
			expectedIngressTLS: []extensions.IngressTLS{
				extensions.IngressTLS{
					Hosts:      []string{"my.host.com"},
					SecretName: releaseName + "-auto-deploy-tls",
				},
			},
		},
		{
			name:                "with tls disabled",
			values:              map[string]string{"ingress.tls.enabled": "false"},
			expectedAnnotations: map[string]string{"kubernetes.io/ingress.class": "nginx"},
			expectedIngressTLS:  []extensions.IngressTLS(nil),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			output, ret := renderTemplate(t, tc.values, releaseName, templates, tc.expectedErrorRegexp)

			if ret == false {
				return
			}

			ingress := new(extensions.Ingress)
			helm.UnmarshalK8SYaml(t, output, ingress)
			require.Equal(t, tc.expectedAnnotations, ingress.ObjectMeta.Annotations)
			require.Equal(t, tc.expectedIngressTLS, ingress.Spec.TLS)
		})
	}
}

func TestIngressTemplate_Disable(t *testing.T) {
	templates := []string{"templates/ingress.yaml"}
	releaseName := "ingress-disable-test"
	tcs := []struct {
		name   string
		values map[string]string

		expectedName        string
		expectedErrorRegexp *regexp.Regexp
	}{
		{
			name:         "defaults",
			expectedName: releaseName + "-auto-deploy",
		},
		{
			name:         "with ingress.enabled key undefined, but service is enabled",
			values:       map[string]string{"ingress.enabled": "null", "service.enabled": "true"},
			expectedName: releaseName + "-auto-deploy",
		},
		{
			name:                "with service disabled and track stable",
			values:              map[string]string{"service.enabled": "false", "application.track": "stable"},
			expectedErrorRegexp: regexp.MustCompile("Error: could not find template templates/ingress.yaml in chart"),
		},
		{
			name:                "with service disabled and track non-stable",
			values:              map[string]string{"service.enabled": "false", "application.track": "non-stable"},
			expectedErrorRegexp: regexp.MustCompile("Error: could not find template templates/ingress.yaml in chart"),
		},
		{
			name:                "with ingress disabled",
			values:              map[string]string{"ingress.enabled": "false"},
			expectedErrorRegexp: regexp.MustCompile("Error: could not find template templates/ingress.yaml in chart"),
		},
		{
			name:                "with ingress enabled and service disabled",
			values:              map[string]string{"ingress.enabled": "true", "service.enabled": "false"},
			expectedErrorRegexp: regexp.MustCompile("Error: could not find template templates/ingress.yaml in chart"),
		},
		{
			name:                "with ingress disabled and service enabled and track stable",
			values:              map[string]string{"ingress.enabled": "false", "service.enabled": "true", "application.track": "stable"},
			expectedErrorRegexp: regexp.MustCompile("Error: could not find template templates/ingress.yaml in chart"),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			opts := &helm.Options{
				SetValues: tc.values,
			}
			output, err := helm.RenderTemplateE(t, opts, helmChartPath, releaseName, templates)

			if tc.expectedErrorRegexp != nil {
				require.Regexp(t, tc.expectedErrorRegexp, err.Error())
				return
			}
			if err != nil {
				t.Error(err)
				return
			}

			ingress := new(extensions.Ingress)
			helm.UnmarshalK8SYaml(t, output, ingress)
			require.Equal(t, tc.expectedName, ingress.ObjectMeta.Name)
		})
	}
}

func TestTemplateLabels(t *testing.T) {
	templates := []string{
		"templates/deployment.yaml", "templates/hpa.yaml",
		"templates/ingress.yaml", "templates/network-policy.yaml",
		"templates/pdb.yaml", "templates/network-policy.yaml",
		"templates/service.yaml"}
	databaseTemplates := []string{
		"templates/db-initialize-job.yaml", "templates/db-migrate-hook.yaml"}
	tcs := []struct {
		name        string
		releaseName string
		values      map[string]string

		expectedName        string
		expectedLabels      map[string]string
		expectedErrorRegexp *regexp.Regexp
	}{
		{
			name:        "defaults",
			releaseName: "production",
			values: map[string]string{
				"hpa.enabled":                 "true",
				"podDisruptionBudget.enabled": "true",
				"networkPolicy.enabled":       "true",
				"resources.requests.cpu":      "100m",
				"resources.requests.memory":   "128Mi"},
			expectedLabels: map[string]string{
				"app":                          "production",
				"chart":                        chartName,
				"release":                      "production",
				"heritage":                     "Helm",
				"app.kubernetes.io/name":       "production",
				"helm.sh/chart":                chartName,
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/instance":   "production",
			},
		},
		{
			name:        "extraLabels",
			releaseName: "production",
			values: map[string]string{
				"hpa.enabled":                 "true",
				"podDisruptionBudget.enabled": "true",
				"networkPolicy.enabled":       "true",
				"resources.requests.cpu":      "100m",
				"resources.requests.memory":   "128Mi",
				"labels.useRecommended":       "false",
				"extraLabels.team":            "myteam",
				"extraLabels.organization":    "myorg",
			},
			expectedLabels: map[string]string{
				"app":                          "production",
				"chart":                        chartName,
				"release":                      "production",
				"heritage":                     "Helm",
				"app.kubernetes.io/name":       "production",
				"helm.sh/chart":                chartName,
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/instance":   "production",
				"team":                         "myteam",
				"organization":                 "myorg",
			},
		},
	}
	dbtcs := []struct {
		name        string
		releaseName string
		values      map[string]string

		expectedName        string
		expectedLabels      map[string]string
		expectedErrorRegexp *regexp.Regexp
	}{
		{
			name:           "database",
			releaseName:    "production",
			values:         map[string]string{"application.initializeCommand": "true", "application.migrateCommand": "true"},
			expectedLabels: map[string]string{"app": "production", "release": "production", "heritage": "Helm", "chart": chartName},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			output, ret := renderTemplate(t, tc.values, tc.releaseName, templates, tc.expectedErrorRegexp)

			if ret == false {
				return
			}

			metaObj := metav1.PartialObjectMetadata{}
			helm.UnmarshalK8SYaml(t, output, &metaObj)
			for key, value := range tc.expectedLabels {
				require.Equal(t, metaObj.ObjectMeta.Labels[key], value)
			}
		})
	}

	for _, tc := range dbtcs {
		t.Run(tc.name, func(t *testing.T) {
			output, ret := renderTemplate(t, tc.values, tc.releaseName, databaseTemplates, tc.expectedErrorRegexp)

			if ret == false {
				return
			}

			metaObj := metav1.PartialObjectMetadata{}
			helm.UnmarshalK8SYaml(t, output, &metaObj)
			for key, value := range tc.expectedLabels {
				require.Equal(t, metaObj.ObjectMeta.Labels[key], value)
			}
		})
	}
}

func TestServiceTemplate_DifferentTracks(t *testing.T) {
	templates := []string{"templates/service.yaml"}
	tcs := []struct {
		name        string
		releaseName string
		values      map[string]string

		expectedName        string
		expectedLabels      map[string]string
		expectedSelector    map[string]string
		expectedErrorRegexp *regexp.Regexp
	}{
		{
			name:             "defaults",
			releaseName:      "production",
			expectedName:     "production-auto-deploy",
			expectedLabels:   map[string]string{"app": "production", "release": "production", "track": "stable"},
			expectedSelector: map[string]string{"app": "production", "tier": "web", "track": "stable"},
		},
		{
			name:             "with canary track",
			releaseName:      "production-canary",
			values:           map[string]string{"application.track": "canary"},
			expectedName:     "production-canary-auto-deploy",
			expectedLabels:   map[string]string{"app": "production-canary", "release": "production-canary", "track": "canary"},
			expectedSelector: map[string]string{"app": "production-canary", "tier": "web", "track": "canary"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			output, ret := renderTemplate(t, tc.values, tc.releaseName, templates, tc.expectedErrorRegexp)

			if ret == false {
				return
			}

			service := new(coreV1.Service)
			helm.UnmarshalK8SYaml(t, output, service)
			require.Equal(t, tc.expectedName, service.ObjectMeta.Name)
			for key, value := range tc.expectedLabels {
				require.Equal(t, service.ObjectMeta.Labels[key], value)
			}
			for key, value := range tc.expectedSelector {
				require.Equal(t, service.Spec.Selector[key], value)
			}
		})
	}
}

func TestIngressTemplate_HTTPPath(t *testing.T) {
	templates := []string{"templates/ingress.yaml"}
	releaseName := "ingress-http-path-test"
	tcs := []struct {
		name   string
		values map[string]string

		expectedpath string
	}{
		{
			name:         "defaults",
			expectedpath: "/",
		},
		{
			name:         "with /*",
			values:       map[string]string{"ingress.path": "/*"},
			expectedpath: "/*",
		},
		{
			name:         "with /myapi",
			values:       map[string]string{"ingress.path": "/myapi"},
			expectedpath: "/myapi",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			opts := &helm.Options{
				SetValues: tc.values,
			}
			output := helm.RenderTemplate(t, opts, helmChartPath, releaseName, templates)

			ingress := new(extensions.Ingress)

			helm.UnmarshalK8SYaml(t, output, ingress)
			require.Equal(t, tc.expectedpath, ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path)
		})
	}
}

func TestIngressTemplate_TLSSecret(t *testing.T) {
	templates := []string{"templates/ingress.yaml"}
	releaseName := "ingress-secret-name-test"
	tcs := []struct {
		name   string
		values map[string]string

		expectedsecretname string
	}{
		{
			name:               "default condition from values - use the provided secretName",
			expectedsecretname: releaseName + "-auto-deploy-tls",
		},
		{
			name:               "don't set the secretName, use the default secret/cert",
			values:             map[string]string{"ingress.tls.useDefaultSecret": "true"},
			expectedsecretname: "",
		},
		{
			name:               "use the provided secretName",
			values:             map[string]string{"ingress.useDefaultSecret": "false"},
			expectedsecretname: releaseName + "-auto-deploy-tls",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			opts := &helm.Options{
				SetValues: tc.values,
			}
			output := helm.RenderTemplate(t, opts, helmChartPath, releaseName, templates)

			ingress := new(extensions.Ingress)

			helm.UnmarshalK8SYaml(t, output, ingress)
			require.Equal(t, tc.expectedsecretname, ingress.Spec.TLS[0].SecretName)
		})
	}
}

func TestServiceTemplate_Disable(t *testing.T) {
	templates := []string{"templates/service.yaml"}
	releaseName := "service-disable-test"
	tcs := []struct {
		name   string
		values map[string]string

		expectedName        string
		expectedErrorRegexp *regexp.Regexp
	}{
		{
			name:         "defaults",
			expectedName: releaseName + "-auto-deploy",
		},
		{
			name:                "with service disabled and track stable",
			values:              map[string]string{"service.enabled": "false", "application.track": "stable"},
			expectedErrorRegexp: regexp.MustCompile("Error: could not find template templates/service.yaml in chart"),
		},
		{
			name:                "with service disabled and track non-stable",
			values:              map[string]string{"service.enabled": "false", "application.track": "non-stable"},
			expectedErrorRegexp: regexp.MustCompile("Error: could not find template templates/service.yaml in chart"),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			output, ret := renderTemplate(t, tc.values, releaseName, templates, tc.expectedErrorRegexp)

			if ret == false {
				return
			}

			service := new(coreV1.Service)
			helm.UnmarshalK8SYaml(t, output, service)
			require.Equal(t, tc.expectedName, service.ObjectMeta.Name)
		})
	}
}

func renderTemplate(t *testing.T, values map[string]string, releaseName string, templates []string, expectedErrorRegexp *regexp.Regexp) (string, bool) {
	opts := &helm.Options{
		SetValues: values,
	}

	output, err := helm.RenderTemplateE(t, opts, helmChartPath, releaseName, templates)
	if expectedErrorRegexp != nil {
		if err == nil {
			t.Error("Expected error but didn't happen")
		} else {
			require.Regexp(t, expectedErrorRegexp, err.Error())
		}
		return "", false
	}
	if err != nil {
		t.Error(err)
		return "", false
	}

	return output, true
}

type workerDeploymentTestCase struct {
	ExpectedName           string
	ExpectedCmd            []string
	ExpectedStrategyType   appsV1.DeploymentStrategyType
	ExpectedSelector       *metav1.LabelSelector
	ExpectedLivenessProbe  *coreV1.Probe
	ExpectedReadinessProbe *coreV1.Probe
}

type workerDeploymentSelectorTestCase struct {
	ExpectedName     string
	ExpectedSelector *metav1.LabelSelector
}

type deploymentList struct {
	metav1.TypeMeta `json:",inline"`

	Items []appsV1.Deployment `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type deploymentAppsV1List struct {
	metav1.TypeMeta `json:",inline"`

	Items []appsV1.Deployment `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func mergeStringMap(dst, src map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}

func defaultLivenessProbe() *coreV1.Probe {
	return &coreV1.Probe{
		Handler: coreV1.Handler{
			HTTPGet: &coreV1.HTTPGetAction{
				Path:   "/",
				Port:   intstr.FromInt(5000),
				Scheme: coreV1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      15,
	}
}

func defaultReadinessProbe() *coreV1.Probe {
	return &coreV1.Probe{
		Handler: coreV1.Handler{
			HTTPGet: &coreV1.HTTPGetAction{
				Path:   "/",
				Port:   intstr.FromInt(5000),
				Scheme: coreV1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 5,
		TimeoutSeconds:      3,
	}
}

func workerLivenessProbe() *coreV1.Probe {
	return &coreV1.Probe{
		Handler: coreV1.Handler{
			HTTPGet: &coreV1.HTTPGetAction{
				Path:   "/worker",
				Port:   intstr.FromInt(5000),
				Scheme: coreV1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 0,
		TimeoutSeconds:      0,
	}
}

func workerReadinessProbe() *coreV1.Probe {
	return &coreV1.Probe{
		Handler: coreV1.Handler{
			HTTPGet: &coreV1.HTTPGetAction{
				Path:   "/worker",
				Port:   intstr.FromInt(5000),
				Scheme: coreV1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 0,
		TimeoutSeconds:      0,
	}
}
