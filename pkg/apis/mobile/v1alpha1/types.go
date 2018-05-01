package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// MobileClientList is a list of MobileClient objects.
type MobileClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []MobileClient `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type MobileClientSpec struct {
	ApiKey        string `json:"apiKey"`
	AppIdentifier string `json:"appIdentifier"`
	ClientType    string `json:"clientType"`
	DmzUrl        string `json:"dmzUrl"`
	Name          string `json:"name"`
}

// +genclient

type MobileClient struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              MobileClientSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}
