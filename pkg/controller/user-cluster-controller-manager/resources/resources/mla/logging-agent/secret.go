/*
Copyright 2021 The Kubermatic Kubernetes Platform contributors.

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

package loggingagent

import (
	"bytes"
	"html/template"

	"k8c.io/kubermatic/v2/pkg/resources"
	"k8c.io/kubermatic/v2/pkg/resources/certificates"
	"k8c.io/reconciler/pkg/reconciling"

	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	MLAGatewayURL string
	TLSCertFile   string
	TLSKeyFile    string
	TLSCACertFile string
  CustomScrapeConfigs string
}

func SecretReconciler(config Config) reconciling.NamedSecretReconcilerFactory {
	return func() (string, reconciling.SecretReconciler) {
		return resources.MLALoggingAgentSecretName, func(secret *corev1.Secret) (*corev1.Secret, error) {
			if secret.Data == nil {
				secret.Data = map[string][]byte{}
			}
			t, err := template.New("agent").Parse(configTemplate)
			if err != nil {
				return nil, err
			}
			configBuf := bytes.Buffer{}
			if err := t.Execute(&configBuf, config); err != nil {
				return nil, err
			}
			secret.Data["agent.yaml"] = configBuf.Bytes()
			secret.Labels = resources.BaseAppLabels(appName, nil)
			return secret, nil
		}
	}
}

const (
	configTemplate = `
logs:
  configs:
  - name: default
    clients:
    - url: {{ .MLAGatewayURL }}
      tls_config:
        cert_file: {{ .TLSCertFile }}
        key_file: {{ .TLSKeyFile }}
        ca_file: {{ .TLSCACertFile }}

    positions:
      filename: /run/grafana-agent/positions.yaml

    scrape_configs:
      # See also https://github.com/grafana/loki/blob/master/production/ksonnet/promtail/scrape_config.libsonnet for reference

      # Pods with a label 'app.kubernetes.io/name'
      - job_name: kubernetes-pods-app-kubernetes-io-name
        pipeline_stages:
          - cri: {}
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - action: replace
            source_labels:
              - __meta_kubernetes_pod_label_app_kubernetes_io_name
            target_label: app
          - action: drop
            regex: ''
            source_labels:
              - app
          - action: replace
            source_labels:
              - __meta_kubernetes_pod_label_app_kubernetes_io_component
            target_label: component
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_node_name
            target_label: node_name
          - action: replace
            source_labels:
            - __meta_kubernetes_namespace
            target_label: namespace
          - action: replace
            replacement: $1
            separator: /
            source_labels:
            - namespace
            - app
            target_label: job
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_name
            target_label: pod
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_container_name
            target_label: container
          - action: replace
            replacement: /var/log/pods/*$1/*.log
            separator: /
            source_labels:
            - __meta_kubernetes_pod_uid
            - __meta_kubernetes_pod_container_name
            target_label: __path__
          - action: replace
            replacement: /var/log/pods/*$1/*.log
            separator: /
            source_labels:
            - __meta_kubernetes_pod_annotation_kubernetes_io_config_hash
            - __meta_kubernetes_pod_container_name
            target_label: __path__

      # Pods with a label 'app'
      - job_name: kubernetes-pods-app
        pipeline_stages:
          - cri: {}
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          # Drop pods with label 'app.kubernetes.io/name'. They are already considered above
          - action: drop
            regex: .+
            source_labels:
              - __meta_kubernetes_pod_label_app_kubernetes_io_name
          - action: replace
            source_labels:
              - __meta_kubernetes_pod_label_app
            target_label: app
          - action: drop
            regex: ''
            source_labels:
              - app
          - action: replace
            source_labels:
              - __meta_kubernetes_pod_label_component
            target_label: component
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_node_name
            target_label: node_name
          - action: replace
            source_labels:
            - __meta_kubernetes_namespace
            target_label: namespace
          - action: replace
            replacement: $1
            separator: /
            source_labels:
            - namespace
            - app
            target_label: job
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_name
            target_label: pod
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_container_name
            target_label: container
          - action: replace
            replacement: /var/log/pods/*$1/*.log
            separator: /
            source_labels:
            - __meta_kubernetes_pod_uid
            - __meta_kubernetes_pod_container_name
            target_label: __path__
          - action: replace
            replacement: /var/log/pods/*$1/*.log
            separator: /
            source_labels:
            - __meta_kubernetes_pod_annotation_kubernetes_io_config_hash
            - __meta_kubernetes_pod_container_name
            target_label: __path__

      # Pods with direct controllers, such as StatefulSet
      - job_name: kubernetes-pods-direct-controllers
        pipeline_stages:
          - cri: {}
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          # Drop pods with label 'app.kubernetes.io/name' or 'app'. They are already considered above
          - action: drop
            regex: .+
            separator: ''
            source_labels:
              - __meta_kubernetes_pod_label_app_kubernetes_io_name
              - __meta_kubernetes_pod_label_app
          - action: drop
            regex: '[0-9a-z-.]+-[0-9a-f]{8,10}'
            source_labels:
              - __meta_kubernetes_pod_controller_name
          - action: replace
            source_labels:
              - __meta_kubernetes_pod_controller_name
            target_label: app
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_node_name
            target_label: node_name
          - action: replace
            source_labels:
            - __meta_kubernetes_namespace
            target_label: namespace
          - action: replace
            replacement: $1
            separator: /
            source_labels:
            - namespace
            - app
            target_label: job
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_name
            target_label: pod
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_container_name
            target_label: container
          - action: replace
            replacement: /var/log/pods/*$1/*.log
            separator: /
            source_labels:
            - __meta_kubernetes_pod_uid
            - __meta_kubernetes_pod_container_name
            target_label: __path__
          - action: replace
            replacement: /var/log/pods/*$1/*.log
            separator: /
            source_labels:
            - __meta_kubernetes_pod_annotation_kubernetes_io_config_hash
            - __meta_kubernetes_pod_container_name
            target_label: __path__

      # Pods with indirect controllers, such as Deployment
      - job_name: kubernetes-pods-indirect-controller
        pipeline_stages:
          - cri: {}
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          # Drop pods with label 'app.kubernetes.io/name' or 'app'. They are already considered above
          - action: drop
            regex: .+
            separator: ''
            source_labels:
              - __meta_kubernetes_pod_label_app_kubernetes_io_name
              - __meta_kubernetes_pod_label_app
          - action: keep
            regex: '[0-9a-z-.]+-[0-9a-f]{8,10}'
            source_labels:
              - __meta_kubernetes_pod_controller_name
          - action: replace
            regex: '([0-9a-z-.]+)-[0-9a-f]{8,10}'
            source_labels:
              - __meta_kubernetes_pod_controller_name
            target_label: app
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_node_name
            target_label: node_name
          - action: replace
            source_labels:
            - __meta_kubernetes_namespace
            target_label: namespace
          - action: replace
            replacement: $1
            separator: /
            source_labels:
            - namespace
            - app
            target_label: job
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_name
            target_label: pod
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_container_name
            target_label: container
          - action: replace
            replacement: /var/log/pods/*$1/*.log
            separator: /
            source_labels:
            - __meta_kubernetes_pod_uid
            - __meta_kubernetes_pod_container_name
            target_label: __path__
          - action: replace
            replacement: /var/log/pods/*$1/*.log
            separator: /
            source_labels:
            - __meta_kubernetes_pod_annotation_kubernetes_io_config_hash
            - __meta_kubernetes_pod_container_name
            target_label: __path__
      # All remaining pods not yet covered
      - job_name: kubernetes-other
        pipeline_stages:
          - cri: {}
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          # Drop what has already been covered
          - action: drop
            regex: .+
            separator: ''
            source_labels:
              - __meta_kubernetes_pod_label_app_kubernetes_io_name
              - __meta_kubernetes_pod_label_app
          - action: drop
            regex: .+
            source_labels:
              - __meta_kubernetes_pod_controller_name
          - action: replace
            source_labels:
              - __meta_kubernetes_pod_name
            target_label: app
          - action: replace
            source_labels:
              - __meta_kubernetes_pod_label_component
            target_label: component
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_node_name
            target_label: node_name
          - action: replace
            source_labels:
            - __meta_kubernetes_namespace
            target_label: namespace
          - action: replace
            replacement: $1
            separator: /
            source_labels:
            - namespace
            - app
            target_label: job
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_name
            target_label: pod
          - action: replace
            source_labels:
            - __meta_kubernetes_pod_container_name
            target_label: container
          - action: replace
            replacement: /var/log/pods/*$1/*.log
            separator: /
            source_labels:
            - __meta_kubernetes_pod_uid
            - __meta_kubernetes_pod_container_name
            target_label: __path__
          - action: replace
            replacement: /var/log/pods/*$1/*.log
            separator: /
            source_labels:
            - __meta_kubernetes_pod_annotation_kubernetes_io_config_hash
            - __meta_kubernetes_pod_container_name
            target_label: __path__
{{- with .CustomScrapeConfigs }}
    #######################################################################
    # custom scraping configurations
{{ . | indent 4 }}
{{- end }}

`
)

func ClientCertificateReconciler(ca *resources.ECDSAKeyPair) reconciling.NamedSecretReconcilerFactory {
	return func() (string, reconciling.SecretReconciler) {
		return resources.MLALoggingAgentCertificatesSecretName,
			certificates.GetECDSAClientCertificateReconciler(
				resources.MLALoggingAgentCertificatesSecretName,
				resources.MLALoggingAgentCertificateCommonName,
				[]string{},
				resources.MLALoggingAgentClientCertSecretKey,
				resources.MLALoggingAgentClientKeySecretKey,
				func() (*resources.ECDSAKeyPair, error) { return ca, nil })
	}
}
