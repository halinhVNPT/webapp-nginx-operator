/*
Copyright 2025.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NginxWebAppSpec defines the desired state of NginxWebApp
type NginxWebAppSpec struct {

	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas,omitempty"`

	// +kubebuilder:default="nginx:stable"
	Image string `json:"image,omitempty"`

	// Port to expose
	// +kubebuilder:default=80
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port,omitempty"`

	// LbConfig (optional) - config cho LoadBalancer/Service
	// +optional
	LbConfig *LoadBalancerConfig `json:"lbConfig,omitempty"`
	// Ingress config (optional)
	// +optional
	Ingress *IngressConfig `json:"ingress,omitempty"`
}
type LoadBalancerConfig struct {
	// Enable to use LB
	// +optional
	Enable bool `json:"enable,omitempty"`

	// Annotations will be applied to the LB/Service/Ingress
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Port exposed by load balancer/ingress
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +optional
	Port int32 `json:"port,omitempty"`
}

// IngressConfig config for creating a networking.k8s.io/v1 Ingress
type IngressConfig struct {
	// Enable creating an Ingress resource
	// +optional
	Enable bool `json:"enable,omitempty"`

	// IngressClassName for the Ingress (e.g., "nginx", "traefik")
	// +optional
	IngressClassName *string `json:"ingressClassName,omitempty"`

	// Annotations applied to the Ingress resource
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Host for the Ingress (e.g., example.com)
	// +optional
	Host string `json:"host,omitempty"`

	// Paths handled by this Ingress; backendPort defaults to spec.port if omitted
	// +optional
	Paths []IngressPath `json:"paths,omitempty"`

	// TLS configuration (optional)
	// +optional
	TLS []IngressTLS `json:"tls,omitempty"`
}
type IngressPath struct {
	// Path (e.g. / or /foo)
	// +optional
	Path string `json:"path,omitempty"`

	// PathType: ImplementationSpecific | Exact | Prefix
	// +kubebuilder:validation:Enum=ImplementationSpecific;Exact;Prefix
	// +optional
	PathType string `json:"pathType,omitempty"`

	// Backend service port (int). If omitted, defaults to spec.port
	// +optional
	BackendPort int32 `json:"backendPort,omitempty"`
}

type IngressTLS struct {
	// Hosts covered by this TLS entry
	// +optional
	Hosts []string `json:"hosts,omitempty"`

	// Secret name that contains TLS cert for the hosts
	// +optional
	SecretName string `json:"secretName,omitempty"`
}

// NginxWebAppStatus defines the observed state of NginxWebApp.
type NginxWebAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the NginxWebApp resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// AvailableReplicas shows how many replicas are up and running
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// Endpoint (ingress or lb)
	Endpoint string `json:"endpoint,omitempty"`

	// Phase ( Pending, Running, Error...)
	// +kubebuilder:default=Pending
	Phase string `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="ENDPOINT",type=string,JSONPath=`.status.endpoint`
// +kubebuilder:printcolumn:name="AVAILABLE",type=integer,JSONPath=`.status.availableReplicas`
// NginxWebApp is the Schema for the nginxwebapps API
type NginxWebApp struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of NginxWebApp
	// +required
	Spec NginxWebAppSpec `json:"spec"`

	// status defines the observed state of NginxWebApp
	// +optional
	Status NginxWebAppStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// NginxWebAppList contains a list of NginxWebApp
type NginxWebAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NginxWebApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NginxWebApp{}, &NginxWebAppList{})
}
