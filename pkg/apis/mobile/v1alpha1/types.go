package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// MobileClientList is a list of MobileClient objects.
type MobileClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []MobileClient `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type MobileClientSpec struct {
	Name             string   `json:"name"`
	ApiKey           string   `json:"apiKey"`
	ClientType       string   `json:"clientType"`
	AppIdentifier    string   `json:"appIdentifier"`
	ExcludedServices []string `json:"excludedServices"`
}

// +genclient

type MobileClient struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              MobileClientSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}
