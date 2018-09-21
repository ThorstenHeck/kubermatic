package etcd

import (
	"fmt"

	"github.com/kubermatic/kubermatic/api/pkg/resources"
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Service returns a service for the etcd StatefulSet
func Service(data resources.ServiceDataProvider, existing *corev1.Service) (*corev1.Service, error) {
	se := existing
	if se == nil {
		se = &corev1.Service{}
	}

	se.Name = resources.EtcdServiceName
	se.OwnerReferences = []metav1.OwnerReference{data.GetClusterRef()}
	se.Annotations = map[string]string{
		"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
	}
	se.Spec.ClusterIP = "None"
	se.Spec.Selector = map[string]string{
		resources.AppLabelKey: name,
		"cluster":             data.Cluster().Name,
	}
	se.Spec.Ports = []corev1.ServicePort{
		{
			Name:       "client",
			Port:       2379,
			TargetPort: intstr.FromInt(2379),
			Protocol:   corev1.ProtocolTCP,
		},
		{
			Name:       "peer",
			Port:       2380,
			TargetPort: intstr.FromInt(2380),
			Protocol:   corev1.ProtocolTCP,
		},
	}

	return se, nil
}

// GetClientEndpoints returns the slice with the etcd endpoints for client communication
func GetClientEndpoints(namespace string) []string {
	var endpoints []string
	for i := 0; i < 3; i++ {
		// Pod DNS name
		serviceDNSName := resources.GetAbsoluteServiceDNSName(resources.EtcdServiceName, namespace)
		absolutePodDNSName := fmt.Sprintf("https://etcd-%d.%s:2379", i, serviceDNSName)
		endpoints = append(endpoints, absolutePodDNSName)
	}
	return endpoints
}
