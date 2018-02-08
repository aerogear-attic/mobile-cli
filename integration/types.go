package integration

// MobileClientSpec represents a mobile client application
type MobileClientSpec struct {
	ID            string
	Name          string
	APIKey        string
	ClientType    string
	AppIdentifier string
	Namespace     string
}

// MobileClientJSON represents a mobile client application
type MobileClientJSON struct {
	Spec MobileClientSpec
}

type ProvisionServiceParams struct {
	ServiceName string
	Namespace   string
	Params      []string
}
