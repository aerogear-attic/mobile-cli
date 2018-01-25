package mobile

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// MobileClientList is a list of MobileClient objects.
type MobileClientList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []MobileClient
}

type MobileClientSpec struct {
	Name          string
	ApiKey        string
	ClientType    string
	AppIdentifier string
}

// +genclient

type MobileClient struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec MobileClientSpec
}
