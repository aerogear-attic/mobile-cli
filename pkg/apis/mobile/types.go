package mobile

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ClientList is a list of Client objects.
type ClientList struct {
	metav1.TypeMeta
	metav1.ListMeta

	Items []Client
}

type ClientSpec struct {
	Name          string
	ApiKey        string
	ClientType    string
	AppIdentifier string
	DmzUrl        string
}

// +genclient

type Client struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec ClientSpec
}
