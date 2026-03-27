package resources

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	HTTPProxyGVR = schema.GroupVersionResource{
		Group: "networking.datumapis.com", Version: "v1alpha", Resource: "httpproxies",
	}
	GatewayGVR = schema.GroupVersionResource{
		Group: "gateway.networking.k8s.io", Version: "v1", Resource: "gateways",
	}
	DomainGVR = schema.GroupVersionResource{
		Group: "networking.datumapis.com", Version: "v1alpha", Resource: "domains",
	}
	ConnectorGVR = schema.GroupVersionResource{
		Group: "networking.datumapis.com", Version: "v1alpha1", Resource: "connectors",
	}
	ExportPolicyGVR = schema.GroupVersionResource{
		Group: "telemetry.miloapis.com", Version: "v1alpha1", Resource: "exportpolicies",
	}
	DNSRecordSetGVR = schema.GroupVersionResource{
		Group: "dns.networking.miloapis.com", Version: "v1alpha1", Resource: "dnsrecordsets",
	}
)

// ResourceDef describes a switchable resource type.
type ResourceDef struct {
	GVR         schema.GroupVersionResource
	ShortName   string
	DisplayName string
	Columns     []string
}

var All = []ResourceDef{
	{GVR: HTTPProxyGVR, ShortName: "hp", DisplayName: "HTTPProxy", Columns: []string{"Name", "Namespace", "Hostnames", "Accepted", "Programmed", "Cert Ready", "DNS", "Age"}},
	{GVR: GatewayGVR, ShortName: "gw", DisplayName: "Gateway", Columns: []string{"Name", "Namespace", "Listeners", "Age"}},
	{GVR: DomainGVR, ShortName: "do", DisplayName: "Domain", Columns: []string{"Name", "Namespace", "Age"}},
	{GVR: ConnectorGVR, ShortName: "co", DisplayName: "Connector", Columns: []string{"Name", "Namespace", "Class", "Ready", "Age"}},
	{GVR: ExportPolicyGVR, ShortName: "ep", DisplayName: "ExportPolicy", Columns: []string{"Name", "Namespace", "Age"}},
	{GVR: DNSRecordSetGVR, ShortName: "dns", DisplayName: "DNSRecordSet", Columns: []string{"Name", "Namespace", "Hostname", "Type", "Age"}},
}
