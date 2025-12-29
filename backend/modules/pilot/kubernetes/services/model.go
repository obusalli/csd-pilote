package services

import (
	"time"

	"github.com/google/uuid"
)

// K8sService represents a Kubernetes service
type K8sService struct {
	ClusterID     uuid.UUID         `json:"clusterId"`
	Namespace     string            `json:"namespace"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	ClusterIP     string            `json:"clusterIP"`
	ExternalIP    string            `json:"externalIP,omitempty"`
	Ports         []ServicePort     `json:"ports"`
	Selector      map[string]string `json:"selector,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
}

// ServicePort represents a service port
type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort string `json:"targetPort"`
	NodePort   int32  `json:"nodePort,omitempty"`
}

// ServiceFilter contains filter options
type ServiceFilter struct {
	Search *string `json:"search,omitempty"`
	Type   *string `json:"type,omitempty"`
}

// CreateServiceInput contains input for creating a service
type CreateServiceInput struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	Selector  map[string]string `json:"selector"`
	Ports     []ServicePortInput `json:"ports"`
}

// ServicePortInput contains input for a service port
type ServicePortInput struct {
	Name       string `json:"name,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"targetPort"`
}
