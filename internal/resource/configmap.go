package resource

import (
	rabbitmqv1beta1 "github.com/pivotal/rabbitmq-for-kubernetes/api/v1beta1"
	"github.com/pivotal/rabbitmq-for-kubernetes/internal/metadata"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

const (
	serverConfigMapName = "server-conf"
	defaultRabbitmqConf = `cluster_formation.peer_discovery_backend = rabbit_peer_discovery_k8s
cluster_formation.k8s.host = kubernetes.default.svc.cluster.local
cluster_formation.k8s.address_type = hostname
cluster_formation.node_cleanup.interval = 30
cluster_formation.node_cleanup.only_log_warning = true
cluster_partition_handling = pause_minority
queue_master_locator = min-masters`
)

var RequiredPlugins = []string{
	"rabbitmq_peer_discovery_k8s", // required for clustering
	"rabbitmq_prometheus",         // enforce prometheus metrics
	"rabbitmq_management",
}

type ServerConfigMapBuilder struct {
	Instance *rabbitmqv1beta1.RabbitmqCluster
}

func (builder *RabbitmqResourceBuilder) ServerConfigMap() *ServerConfigMapBuilder {
	return &ServerConfigMapBuilder{
		Instance: builder.Instance,
	}
}

func (builder *ServerConfigMapBuilder) Update(object runtime.Object) error {
	configMap := object.(*corev1.ConfigMap)
	configMap.Labels = metadata.GetLabels(builder.Instance.Name, builder.Instance.Labels)
	configMap.Annotations = metadata.ReconcileAnnotations(configMap.GetAnnotations(), builder.Instance.Annotations)
	return nil
}

func (builder *ServerConfigMapBuilder) Build() (runtime.Object, error) {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      builder.Instance.ChildResourceName(serverConfigMapName),
			Namespace: builder.Instance.Namespace,
		},
		Data: map[string]string{
			"rabbitmq.conf":   defaultRabbitmqConf,
			"enabled_plugins": "[" + strings.Join(appendIfUnique(RequiredPlugins, builder.Instance.Spec.Rabbitmq.AdditionalPlugins), ",") + "].",
		},
	}, nil
}

func appendIfUnique(a, b []string) []string {
	check := make(map[string]bool)
	list := append(a, b...)
	set := make([]string, 0)
	for _, s := range list {
		if _, value := check[s]; !value {
			check[s] = true
			set = append(set, s)
		}
	}
	return set
}
