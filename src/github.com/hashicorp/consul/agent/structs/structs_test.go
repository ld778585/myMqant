package structs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/agent/cache"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/hashicorp/consul/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	arg := &RegisterRequest{
		Datacenter: "foo",
		Node:       "bar",
		Address:    "baz",
		Service: &NodeService{
			Service: "test",
			Address: "127.0.0.2",
		},
	}
	buf, err := Encode(RegisterRequestType, arg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	var out RegisterRequest
	err = Decode(buf[1:], &out)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if !reflect.DeepEqual(arg.Service, out.Service) {
		t.Fatalf("bad: %#v %#v", arg.Service, out.Service)
	}
	if !reflect.DeepEqual(arg, &out) {
		t.Fatalf("bad: %#v %#v", arg, out)
	}
}

func TestStructs_Implements(t *testing.T) {
	var (
		_ RPCInfo          = &RegisterRequest{}
		_ RPCInfo          = &DeregisterRequest{}
		_ RPCInfo          = &DCSpecificRequest{}
		_ RPCInfo          = &ServiceSpecificRequest{}
		_ RPCInfo          = &NodeSpecificRequest{}
		_ RPCInfo          = &ChecksInStateRequest{}
		_ RPCInfo          = &KVSRequest{}
		_ RPCInfo          = &KeyRequest{}
		_ RPCInfo          = &KeyListRequest{}
		_ RPCInfo          = &SessionRequest{}
		_ RPCInfo          = &SessionSpecificRequest{}
		_ RPCInfo          = &EventFireRequest{}
		_ RPCInfo          = &ACLPolicyResolveLegacyRequest{}
		_ RPCInfo          = &ACLPolicyBatchGetRequest{}
		_ RPCInfo          = &ACLPolicyGetRequest{}
		_ RPCInfo          = &ACLTokenGetRequest{}
		_ RPCInfo          = &KeyringRequest{}
		_ CompoundResponse = &KeyringResponses{}
	)
}

func TestStructs_RegisterRequest_ChangesNode(t *testing.T) {
	req := &RegisterRequest{
		ID:              types.NodeID("40e4a748-2192-161a-0510-9bf59fe950b5"),
		Node:            "test",
		Address:         "127.0.0.1",
		Datacenter:      "dc1",
		TaggedAddresses: make(map[string]string),
		NodeMeta: map[string]string{
			"role": "server",
		},
	}

	node := &Node{
		ID:              types.NodeID("40e4a748-2192-161a-0510-9bf59fe950b5"),
		Node:            "test",
		Address:         "127.0.0.1",
		Datacenter:      "dc1",
		TaggedAddresses: make(map[string]string),
		Meta: map[string]string{
			"role": "server",
		},
	}

	check := func(twiddle, restore func()) {
		if req.ChangesNode(node) {
			t.Fatalf("should not change")
		}

		twiddle()
		if !req.ChangesNode(node) {
			t.Fatalf("should change")
		}

		req.SkipNodeUpdate = true
		if req.ChangesNode(node) {
			t.Fatalf("should skip")
		}

		req.SkipNodeUpdate = false
		if !req.ChangesNode(node) {
			t.Fatalf("should change")
		}

		restore()
		if req.ChangesNode(node) {
			t.Fatalf("should not change")
		}
	}

	check(func() { req.ID = "nope" }, func() { req.ID = types.NodeID("40e4a748-2192-161a-0510-9bf59fe950b5") })
	check(func() { req.Node = "nope" }, func() { req.Node = "test" })
	check(func() { req.Address = "127.0.0.2" }, func() { req.Address = "127.0.0.1" })
	check(func() { req.Datacenter = "dc2" }, func() { req.Datacenter = "dc1" })
	check(func() { req.TaggedAddresses["wan"] = "nope" }, func() { delete(req.TaggedAddresses, "wan") })
	check(func() { req.NodeMeta["invalid"] = "nope" }, func() { delete(req.NodeMeta, "invalid") })

	if !req.ChangesNode(nil) {
		t.Fatalf("should change")
	}
}

// testServiceNode gives a fully filled out ServiceNode instance.
func testServiceNode(t *testing.T) *ServiceNode {
	return &ServiceNode{
		ID:         types.NodeID("40e4a748-2192-161a-0510-9bf59fe950b5"),
		Node:       "node1",
		Address:    "127.0.0.1",
		Datacenter: "dc1",
		TaggedAddresses: map[string]string{
			"hello": "world",
		},
		NodeMeta: map[string]string{
			"tag": "value",
		},
		ServiceKind:    ServiceKindTypical,
		ServiceID:      "service1",
		ServiceName:    "dogs",
		ServiceTags:    []string{"prod", "v1"},
		ServiceAddress: "127.0.0.2",
		ServiceTaggedAddresses: map[string]ServiceAddress{
			"lan": ServiceAddress{
				Address: "127.0.0.2",
				Port:    8080,
			},
			"wan": ServiceAddress{
				Address: "198.18.0.1",
				Port:    80,
			},
		},
		ServicePort: 8080,
		ServiceMeta: map[string]string{
			"service": "metadata",
		},
		ServiceEnableTagOverride: true,
		RaftIndex: RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 2,
		},
		ServiceProxy: TestConnectProxyConfig(t),
		ServiceConnect: ServiceConnect{
			Native: true,
		},
	}
}

func TestNode_IsSame(t *testing.T) {
	id := types.NodeID("e62f3b31-9284-4e26-ab14-2a59dea85b55")
	node := "mynode1"
	address := ""
	datacenter := "dc1"
	n := &Node{
		ID:              id,
		Node:            node,
		Datacenter:      datacenter,
		Address:         address,
		TaggedAddresses: make(map[string]string),
		Meta:            make(map[string]string),
		RaftIndex: RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 2,
		},
	}
	other := &Node{
		ID:              id,
		Node:            node,
		Datacenter:      datacenter,
		Address:         address,
		TaggedAddresses: make(map[string]string),
		Meta:            make(map[string]string),
		RaftIndex: RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 3,
		},
	}
	check := func(twiddle, restore func()) {
		t.Helper()
		if !n.IsSame(other) || !other.IsSame(n) {
			t.Fatalf("should be the same")
		}

		twiddle()
		if n.IsSame(other) || other.IsSame(n) {
			t.Fatalf("should be different, was %#v VS %#v", n, other)
		}

		restore()
		if !n.IsSame(other) || !other.IsSame(n) {
			t.Fatalf("should be the same")
		}
	}
	check(func() { other.ID = types.NodeID("") }, func() { other.ID = id })
	check(func() { other.Node = "other" }, func() { other.Node = node })
	check(func() { other.Datacenter = "dcX" }, func() { other.Datacenter = datacenter })
	check(func() { other.Address = "127.0.0.1" }, func() { other.Address = address })
	check(func() { other.TaggedAddresses = map[string]string{"my": "address"} }, func() { other.TaggedAddresses = map[string]string{} })
	check(func() { other.Meta = map[string]string{"my": "meta"} }, func() { other.Meta = map[string]string{} })

	if !n.IsSame(other) {
		t.Fatalf("should be equal, was %#v VS %#v", n, other)
	}
}

func TestStructs_ServiceNode_IsSameService(t *testing.T) {
	sn := testServiceNode(t)
	node := "node1"
	serviceID := sn.ServiceID
	serviceAddress := sn.ServiceAddress
	serviceEnableTagOverride := sn.ServiceEnableTagOverride
	serviceMeta := make(map[string]string)
	for k, v := range sn.ServiceMeta {
		serviceMeta[k] = v
	}
	serviceName := sn.ServiceName
	servicePort := sn.ServicePort
	serviceTags := sn.ServiceTags
	serviceWeights := Weights{Passing: 2, Warning: 1}
	sn.ServiceWeights = serviceWeights
	serviceProxy := sn.ServiceProxy
	serviceConnect := sn.ServiceConnect
	serviceTaggedAddresses := sn.ServiceTaggedAddresses

	n := sn.ToNodeService().ToServiceNode(node)
	other := sn.ToNodeService().ToServiceNode(node)

	check := func(twiddle, restore func()) {
		t.Helper()
		if !n.IsSameService(other) || !other.IsSameService(n) {
			t.Fatalf("should be the same")
		}

		twiddle()
		if n.IsSameService(other) || other.IsSameService(n) {
			t.Fatalf("should be different, was %#v VS %#v", n, other)
		}

		restore()
		if !n.IsSameService(other) || !other.IsSameService(n) {
			t.Fatalf("should be the same after restore, was:\n %#v VS\n %#v", n, other)
		}
	}

	check(func() { other.ServiceID = "66fb695a-c782-472f-8d36-4f3edd754b37" }, func() { other.ServiceID = serviceID })
	check(func() { other.Node = "other" }, func() { other.Node = node })
	check(func() { other.ServiceAddress = "1.2.3.4" }, func() { other.ServiceAddress = serviceAddress })
	check(func() { other.ServiceEnableTagOverride = !serviceEnableTagOverride }, func() { other.ServiceEnableTagOverride = serviceEnableTagOverride })
	check(func() { other.ServiceKind = "newKind" }, func() { other.ServiceKind = "" })
	check(func() { other.ServiceMeta = map[string]string{"my": "meta"} }, func() { other.ServiceMeta = serviceMeta })
	check(func() { other.ServiceName = "duck" }, func() { other.ServiceName = serviceName })
	check(func() { other.ServicePort = 65534 }, func() { other.ServicePort = servicePort })
	check(func() { other.ServiceTags = []string{"new", "tags"} }, func() { other.ServiceTags = serviceTags })
	check(func() { other.ServiceWeights = Weights{Passing: 42, Warning: 41} }, func() { other.ServiceWeights = serviceWeights })
	check(func() { other.ServiceProxy = ConnectProxyConfig{} }, func() { other.ServiceProxy = serviceProxy })
	check(func() { other.ServiceConnect = ServiceConnect{} }, func() { other.ServiceConnect = serviceConnect })
	check(func() { other.ServiceTaggedAddresses = nil }, func() { other.ServiceTaggedAddresses = serviceTaggedAddresses })
}

func TestStructs_ServiceNode_PartialClone(t *testing.T) {
	sn := testServiceNode(t)

	clone := sn.PartialClone()

	// Make sure the parts that weren't supposed to be cloned didn't get
	// copied over, then zero-value them out so we can do a DeepEqual() on
	// the rest of the contents.
	if clone.ID != "" ||
		clone.Address != "" ||
		clone.Datacenter != "" ||
		len(clone.TaggedAddresses) != 0 ||
		len(clone.NodeMeta) != 0 {
		t.Fatalf("bad: %v", clone)
	}

	sn.ID = ""
	sn.Address = ""
	sn.Datacenter = ""
	sn.TaggedAddresses = nil
	sn.NodeMeta = nil
	require.Equal(t, sn, clone)

	sn.ServiceTags = append(sn.ServiceTags, "hello")
	if reflect.DeepEqual(sn, clone) {
		t.Fatalf("clone wasn't independent of the original")
	}

	revert := make([]string, len(sn.ServiceTags)-1)
	copy(revert, sn.ServiceTags[0:len(sn.ServiceTags)-1])
	sn.ServiceTags = revert
	if !reflect.DeepEqual(sn, clone) {
		t.Fatalf("bad: %v VS %v", clone, sn)
	}
	oldPassingWeight := clone.ServiceWeights.Passing
	sn.ServiceWeights.Passing = 1000
	if reflect.DeepEqual(sn, clone) {
		t.Fatalf("clone wasn't independent of the original for Meta")
	}
	sn.ServiceWeights.Passing = oldPassingWeight
	sn.ServiceMeta["new_meta"] = "new_value"
	if reflect.DeepEqual(sn, clone) {
		t.Fatalf("clone wasn't independent of the original for Meta")
	}

	// ensure that the tagged addresses were copied and not just a pointer to the map
	sn.ServiceTaggedAddresses["foo"] = ServiceAddress{Address: "consul.is.awesome", Port: 443}
	require.NotEqual(t, sn, clone)
}

func TestStructs_ServiceNode_Conversions(t *testing.T) {
	sn := testServiceNode(t)

	sn2 := sn.ToNodeService().ToServiceNode("node1")

	// These two fields get lost in the conversion, so we have to zero-value
	// them out before we do the compare.
	sn.ID = ""
	sn.Address = ""
	sn.Datacenter = ""
	sn.TaggedAddresses = nil
	sn.NodeMeta = nil
	sn.ServiceWeights = Weights{Passing: 1, Warning: 1}
	require.Equal(t, sn, sn2)
	if !sn.IsSameService(sn2) || !sn2.IsSameService(sn) {
		t.Fatalf("bad: %#v, should be the same %#v", sn2, sn)
	}
	// Those fields are lost in conversion, so IsSameService() should not take them into account
	sn.Address = "y"
	sn.Datacenter = "z"
	sn.TaggedAddresses = map[string]string{"one": "1", "two": "2"}
	sn.NodeMeta = map[string]string{"meta": "data"}
	if !sn.IsSameService(sn2) || !sn2.IsSameService(sn) {
		t.Fatalf("bad: %#v, should be the same %#v", sn2, sn)
	}
}

func TestStructs_NodeService_ValidateMeshGateway(t *testing.T) {
	type testCase struct {
		Modify func(*NodeService)
		Err    string
	}
	cases := map[string]testCase{
		"valid": testCase{
			func(x *NodeService) {},
			"",
		},
		"zero-port": testCase{
			func(x *NodeService) { x.Port = 0 },
			"Port must be non-zero",
		},
		"sidecar-service": testCase{
			func(x *NodeService) { x.Connect.SidecarService = &ServiceDefinition{} },
			"cannot have a sidecar service",
		},
		"proxy-destination-name": testCase{
			func(x *NodeService) { x.Proxy.DestinationServiceName = "foo" },
			"Proxy.DestinationServiceName configuration is invalid",
		},
		"proxy-destination-id": testCase{
			func(x *NodeService) { x.Proxy.DestinationServiceID = "foo" },
			"Proxy.DestinationServiceID configuration is invalid",
		},
		"proxy-local-address": testCase{
			func(x *NodeService) { x.Proxy.LocalServiceAddress = "127.0.0.1" },
			"Proxy.LocalServiceAddress configuration is invalid",
		},
		"proxy-local-port": testCase{
			func(x *NodeService) { x.Proxy.LocalServicePort = 36 },
			"Proxy.LocalServicePort configuration is invalid",
		},
		"proxy-upstreams": testCase{
			func(x *NodeService) { x.Proxy.Upstreams = []Upstream{Upstream{}} },
			"Proxy.Upstreams configuration is invalid",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ns := TestNodeServiceMeshGateway(t)
			tc.Modify(ns)

			err := ns.Validate()
			if tc.Err == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.Err))
			}
		})
	}
}

func TestStructs_NodeService_ValidateTerminatingGateway(t *testing.T) {
	type testCase struct {
		Modify func(*NodeService)
		Err    string
	}

	cases := map[string]testCase{
		"valid": testCase{
			func(x *NodeService) {},
			"",
		},
		"sidecar-service": testCase{
			func(x *NodeService) { x.Connect.SidecarService = &ServiceDefinition{} },
			"cannot have a sidecar service",
		},
		"proxy-destination-name": testCase{
			func(x *NodeService) { x.Proxy.DestinationServiceName = "foo" },
			"Proxy.DestinationServiceName configuration is invalid",
		},
		"proxy-destination-id": testCase{
			func(x *NodeService) { x.Proxy.DestinationServiceID = "foo" },
			"Proxy.DestinationServiceID configuration is invalid",
		},
		"proxy-local-address": testCase{
			func(x *NodeService) { x.Proxy.LocalServiceAddress = "127.0.0.1" },
			"Proxy.LocalServiceAddress configuration is invalid",
		},
		"proxy-local-port": testCase{
			func(x *NodeService) { x.Proxy.LocalServicePort = 36 },
			"Proxy.LocalServicePort configuration is invalid",
		},
		"proxy-upstreams": testCase{
			func(x *NodeService) { x.Proxy.Upstreams = []Upstream{Upstream{}} },
			"Proxy.Upstreams configuration is invalid",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ns := TestNodeServiceTerminatingGateway(t, "10.0.0.5")
			tc.Modify(ns)

			err := ns.Validate()
			if tc.Err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.Err))
			}
		})
	}
}

func TestStructs_NodeService_ValidateIngressGateway(t *testing.T) {
	type testCase struct {
		Modify func(*NodeService)
		Err    string
	}

	cases := map[string]testCase{
		"valid": testCase{
			func(x *NodeService) {},
			"",
		},
		"sidecar-service": testCase{
			func(x *NodeService) { x.Connect.SidecarService = &ServiceDefinition{} },
			"cannot have a sidecar service",
		},
		"proxy-destination-name": testCase{
			func(x *NodeService) { x.Proxy.DestinationServiceName = "foo" },
			"Proxy.DestinationServiceName configuration is invalid",
		},
		"proxy-destination-id": testCase{
			func(x *NodeService) { x.Proxy.DestinationServiceID = "foo" },
			"Proxy.DestinationServiceID configuration is invalid",
		},
		"proxy-local-address": testCase{
			func(x *NodeService) { x.Proxy.LocalServiceAddress = "127.0.0.1" },
			"Proxy.LocalServiceAddress configuration is invalid",
		},
		"proxy-local-port": testCase{
			func(x *NodeService) { x.Proxy.LocalServicePort = 36 },
			"Proxy.LocalServicePort configuration is invalid",
		},
		"proxy-upstreams": testCase{
			func(x *NodeService) { x.Proxy.Upstreams = []Upstream{Upstream{}} },
			"Proxy.Upstreams configuration is invalid",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ns := TestNodeServiceIngressGateway(t, "10.0.0.5")
			tc.Modify(ns)

			err := ns.Validate()
			if tc.Err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.Err))
			}
		})
	}
}

func TestStructs_NodeService_ValidateExposeConfig(t *testing.T) {
	type testCase struct {
		Modify func(*NodeService)
		Err    string
	}
	cases := map[string]testCase{
		"valid": {
			func(x *NodeService) {},
			"",
		},
		"empty path": {
			func(x *NodeService) { x.Proxy.Expose.Paths[0].Path = "" },
			"empty path exposed",
		},
		"invalid port negative": {
			func(x *NodeService) { x.Proxy.Expose.Paths[0].ListenerPort = -1 },
			"invalid listener port",
		},
		"invalid port too large": {
			func(x *NodeService) { x.Proxy.Expose.Paths[0].ListenerPort = 65536 },
			"invalid listener port",
		},
		"duplicate paths": {
			func(x *NodeService) {
				x.Proxy.Expose.Paths[0].Path = "/metrics"
				x.Proxy.Expose.Paths[1].Path = "/metrics"
			},
			"duplicate paths exposed",
		},
		"duplicate ports": {
			func(x *NodeService) {
				x.Proxy.Expose.Paths[0].ListenerPort = 21600
				x.Proxy.Expose.Paths[1].ListenerPort = 21600
			},
			"duplicate listener ports exposed",
		},
		"protocol not supported": {
			func(x *NodeService) { x.Proxy.Expose.Paths[0].Protocol = "foo" },
			"protocol 'foo' not supported for path",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ns := TestNodeServiceExpose(t)
			tc.Modify(ns)

			err := ns.Validate()
			if tc.Err == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tc.Err))
			}
		})
	}
}

func TestStructs_NodeService_ValidateConnectProxy(t *testing.T) {
	cases := []struct {
		Name   string
		Modify func(*NodeService)
		Err    string
	}{
		{
			"valid",
			func(x *NodeService) {},
			"",
		},

		{
			"connect-proxy: no Proxy.DestinationServiceName",
			func(x *NodeService) { x.Proxy.DestinationServiceName = "" },
			"Proxy.DestinationServiceName must be",
		},

		{
			"connect-proxy: whitespace Proxy.DestinationServiceName",
			func(x *NodeService) { x.Proxy.DestinationServiceName = "  " },
			"Proxy.DestinationServiceName must be",
		},

		{
			"connect-proxy: valid Proxy.DestinationServiceName",
			func(x *NodeService) { x.Proxy.DestinationServiceName = "hello" },
			"",
		},

		{
			"connect-proxy: no port set",
			func(x *NodeService) { x.Port = 0 },
			"Port must",
		},

		{
			"connect-proxy: ConnectNative set",
			func(x *NodeService) { x.Connect.Native = true },
			"cannot also be",
		},

		{
			"connect-proxy: upstream missing type (defaulted)",
			func(x *NodeService) {
				x.Proxy.Upstreams = Upstreams{{
					DestinationName: "foo",
					LocalBindPort:   5000,
				}}
			},
			"",
		},
		{
			"connect-proxy: upstream invalid type",
			func(x *NodeService) {
				x.Proxy.Upstreams = Upstreams{{
					DestinationType: "garbage",
					DestinationName: "foo",
					LocalBindPort:   5000,
				}}
			},
			"unknown upstream destination type",
		},
		{
			"connect-proxy: upstream empty name",
			func(x *NodeService) {
				x.Proxy.Upstreams = Upstreams{{
					DestinationType: UpstreamDestTypeService,
					LocalBindPort:   5000,
				}}
			},
			"upstream destination name cannot be empty",
		},
		{
			"connect-proxy: upstream empty bind port",
			func(x *NodeService) {
				x.Proxy.Upstreams = Upstreams{{
					DestinationType: UpstreamDestTypeService,
					DestinationName: "foo",
					LocalBindPort:   0,
				}}
			},
			"upstream local bind port cannot be zero",
		},
		{
			"connect-proxy: Upstreams almost-but-not-quite-duplicated in various ways",
			func(x *NodeService) {
				x.Proxy.Upstreams = Upstreams{
					{ // baseline
						DestinationType: UpstreamDestTypeService,
						DestinationName: "foo",
						LocalBindPort:   5000,
					},
					{ // different bind address
						DestinationType:  UpstreamDestTypeService,
						DestinationName:  "bar",
						LocalBindAddress: "127.0.0.2",
						LocalBindPort:    5000,
					},
					{ // different datacenter
						DestinationType: UpstreamDestTypeService,
						DestinationName: "foo",
						Datacenter:      "dc2",
						LocalBindPort:   5001,
					},
					{ // explicit default namespace
						DestinationType:      UpstreamDestTypeService,
						DestinationName:      "foo",
						DestinationNamespace: "default",
						LocalBindPort:        5003,
					},
					{ // different namespace
						DestinationType:      UpstreamDestTypeService,
						DestinationName:      "foo",
						DestinationNamespace: "alternate",
						LocalBindPort:        5002,
					},
					{ // different type
						DestinationType: UpstreamDestTypePreparedQuery,
						DestinationName: "foo",
						LocalBindPort:   5004,
					},
				}
			},
			"",
		},
		{
			"connect-proxy: Upstreams duplicated by port",
			func(x *NodeService) {
				x.Proxy.Upstreams = Upstreams{
					{
						DestinationType: UpstreamDestTypeService,
						DestinationName: "foo",
						LocalBindPort:   5000,
					},
					{
						DestinationType: UpstreamDestTypeService,
						DestinationName: "foo",
						LocalBindPort:   5000,
					},
				}
			},
			"upstreams cannot contain duplicates",
		},
		{
			"connect-proxy: Upstreams duplicated by ip and port",
			func(x *NodeService) {
				x.Proxy.Upstreams = Upstreams{
					{
						DestinationType:  UpstreamDestTypeService,
						DestinationName:  "foo",
						LocalBindAddress: "127.0.0.2",
						LocalBindPort:    5000,
					},
					{
						DestinationType:  UpstreamDestTypeService,
						DestinationName:  "bar",
						LocalBindAddress: "127.0.0.2",
						LocalBindPort:    5000,
					},
				}
			},
			"upstreams cannot contain duplicates",
		},
		{
			"connect-proxy: Upstreams duplicated by ip and port with ip defaulted in one",
			func(x *NodeService) {
				x.Proxy.Upstreams = Upstreams{
					{
						DestinationType: UpstreamDestTypeService,
						DestinationName: "foo",
						LocalBindPort:   5000,
					},
					{
						DestinationType:  UpstreamDestTypeService,
						DestinationName:  "foo",
						LocalBindAddress: "127.0.0.1",
						LocalBindPort:    5000,
					},
				}
			},
			"upstreams cannot contain duplicates",
		},
		{
			"connect-proxy: Upstreams duplicated by name",
			func(x *NodeService) {
				x.Proxy.Upstreams = Upstreams{
					{
						DestinationType: UpstreamDestTypeService,
						DestinationName: "foo",
						LocalBindPort:   5000,
					},
					{
						DestinationType: UpstreamDestTypeService,
						DestinationName: "foo",
						LocalBindPort:   5001,
					},
				}
			},
			"upstreams cannot contain duplicates",
		},
		{
			"connect-proxy: Upstreams duplicated by name and datacenter",
			func(x *NodeService) {
				x.Proxy.Upstreams = Upstreams{
					{
						DestinationType: UpstreamDestTypeService,
						DestinationName: "foo",
						Datacenter:      "dc2",
						LocalBindPort:   5000,
					},
					{
						DestinationType: UpstreamDestTypeService,
						DestinationName: "foo",
						Datacenter:      "dc2",
						LocalBindPort:   5001,
					},
				}
			},
			"upstreams cannot contain duplicates",
		},
		{
			"connect-proxy: Upstreams duplicated by name and namespace",
			func(x *NodeService) {
				x.Proxy.Upstreams = Upstreams{
					{
						DestinationType:      UpstreamDestTypeService,
						DestinationName:      "foo",
						DestinationNamespace: "alternate",
						LocalBindPort:        5000,
					},
					{
						DestinationType:      UpstreamDestTypeService,
						DestinationName:      "foo",
						DestinationNamespace: "alternate",
						LocalBindPort:        5001,
					},
				}
			},
			"upstreams cannot contain duplicates",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			assert := assert.New(t)
			ns := TestNodeServiceProxy(t)
			tc.Modify(ns)

			err := ns.Validate()
			assert.Equal(err != nil, tc.Err != "", err)
			if err == nil {
				return
			}

			assert.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.Err))
		})
	}
}

func TestStructs_NodeService_ValidateSidecarService(t *testing.T) {
	cases := []struct {
		Name   string
		Modify func(*NodeService)
		Err    string
	}{
		{
			"valid",
			func(x *NodeService) {},
			"",
		},

		{
			"ID can't be set",
			func(x *NodeService) { x.Connect.SidecarService.ID = "foo" },
			"SidecarService cannot specify an ID",
		},

		{
			"Nested sidecar can't be set",
			func(x *NodeService) {
				x.Connect.SidecarService.Connect = &ServiceConnect{
					SidecarService: &ServiceDefinition{},
				}
			},
			"SidecarService cannot have a nested SidecarService",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			assert := assert.New(t)
			ns := TestNodeServiceSidecar(t)
			tc.Modify(ns)

			err := ns.Validate()
			assert.Equal(err != nil, tc.Err != "", err)
			if err == nil {
				return
			}

			assert.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.Err))
		})
	}
}

func TestStructs_NodeService_IsSame(t *testing.T) {
	ns := &NodeService{
		ID:      "node1",
		Service: "theservice",
		Tags:    []string{"foo", "bar"},
		Address: "127.0.0.1",
		TaggedAddresses: map[string]ServiceAddress{
			"lan": ServiceAddress{
				Address: "127.0.0.1",
				Port:    3456,
			},
			"wan": ServiceAddress{
				Address: "198.18.0.1",
				Port:    1234,
			},
		},
		Meta: map[string]string{
			"meta1": "value1",
			"meta2": "value2",
		},
		Port:              1234,
		EnableTagOverride: true,
		Proxy: ConnectProxyConfig{
			DestinationServiceName: "db",
			Config: map[string]interface{}{
				"foo": "bar",
			},
		},
		Weights: &Weights{Passing: 1, Warning: 1},
	}
	if !ns.IsSame(ns) {
		t.Fatalf("should be equal to itself")
	}

	other := &NodeService{
		ID:                "node1",
		Service:           "theservice",
		Tags:              []string{"foo", "bar"},
		Address:           "127.0.0.1",
		Port:              1234,
		EnableTagOverride: true,
		TaggedAddresses: map[string]ServiceAddress{
			"wan": ServiceAddress{
				Address: "198.18.0.1",
				Port:    1234,
			},
			"lan": ServiceAddress{
				Address: "127.0.0.1",
				Port:    3456,
			},
		},
		Meta: map[string]string{
			// We don't care about order
			"meta2": "value2",
			"meta1": "value1",
		},
		Proxy: ConnectProxyConfig{
			DestinationServiceName: "db",
			Config: map[string]interface{}{
				"foo": "bar",
			},
		},
		Weights: &Weights{Passing: 1, Warning: 1},
		RaftIndex: RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 2,
		},
	}
	if !ns.IsSame(other) || !other.IsSame(ns) {
		t.Fatalf("should not care about Raft fields")
	}

	check := func(twiddle, restore func()) {
		t.Helper()
		if !ns.IsSame(other) || !other.IsSame(ns) {
			t.Fatalf("should be the same")
		}

		twiddle()
		if ns.IsSame(other) || other.IsSame(ns) {
			t.Fatalf("should not be the same")
		}

		restore()
		if !ns.IsSame(other) || !other.IsSame(ns) {
			t.Fatalf("should be the same again")
		}
	}

	check(func() { other.ID = "XXX" }, func() { other.ID = "node1" })
	check(func() { other.Service = "XXX" }, func() { other.Service = "theservice" })
	check(func() { other.Tags = nil }, func() { other.Tags = []string{"foo", "bar"} })
	check(func() { other.Tags = []string{"foo"} }, func() { other.Tags = []string{"foo", "bar"} })
	check(func() { other.Address = "XXX" }, func() { other.Address = "127.0.0.1" })
	check(func() { other.Port = 9999 }, func() { other.Port = 1234 })
	check(func() { other.Meta["meta2"] = "wrongValue" }, func() { other.Meta["meta2"] = "value2" })
	check(func() { other.EnableTagOverride = false }, func() { other.EnableTagOverride = true })
	check(func() { other.Kind = ServiceKindConnectProxy }, func() { other.Kind = "" })
	check(func() { other.Proxy.DestinationServiceName = "" }, func() { other.Proxy.DestinationServiceName = "db" })
	check(func() { other.Proxy.DestinationServiceID = "XXX" }, func() { other.Proxy.DestinationServiceID = "" })
	check(func() { other.Proxy.LocalServiceAddress = "XXX" }, func() { other.Proxy.LocalServiceAddress = "" })
	check(func() { other.Proxy.LocalServicePort = 9999 }, func() { other.Proxy.LocalServicePort = 0 })
	check(func() { other.Proxy.Config["baz"] = "XXX" }, func() { delete(other.Proxy.Config, "baz") })
	check(func() { other.Connect.Native = true }, func() { other.Connect.Native = false })
	otherServiceNode := other.ToServiceNode("node1")
	copyNodeService := otherServiceNode.ToNodeService()
	if !copyNodeService.IsSame(other) {
		t.Fatalf("copy should be the same, but was\n %#v\nVS\n %#v", copyNodeService, other)
	}
	otherServiceNodeCopy2 := copyNodeService.ToServiceNode("node1")
	if !otherServiceNode.IsSameService(otherServiceNodeCopy2) {
		t.Fatalf("copy should be the same, but was\n %#v\nVS\n %#v", otherServiceNode, otherServiceNodeCopy2)
	}
	check(func() { other.TaggedAddresses["lan"] = ServiceAddress{Address: "127.0.0.1", Port: 9999} }, func() { other.TaggedAddresses["lan"] = ServiceAddress{Address: "127.0.0.1", Port: 3456} })
}

func TestStructs_HealthCheck_IsSame(t *testing.T) {
	hc := &HealthCheck{
		Node:        "node1",
		CheckID:     "check1",
		Name:        "thecheck",
		Status:      api.HealthPassing,
		Notes:       "it's all good",
		Output:      "lgtm",
		ServiceID:   "service1",
		ServiceName: "theservice",
		ServiceTags: []string{"foo"},
	}
	if !hc.IsSame(hc) {
		t.Fatalf("should be equal to itself")
	}

	other := &HealthCheck{
		Node:        "node1",
		CheckID:     "check1",
		Name:        "thecheck",
		Status:      api.HealthPassing,
		Notes:       "it's all good",
		Output:      "lgtm",
		ServiceID:   "service1",
		ServiceName: "theservice",
		ServiceTags: []string{"foo"},
		RaftIndex: RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 2,
		},
	}
	if !hc.IsSame(other) || !other.IsSame(hc) {
		t.Fatalf("should not care about Raft fields")
	}

	checkCheckIDField := func(field *types.CheckID) {
		if !hc.IsSame(other) || !other.IsSame(hc) {
			t.Fatalf("should be the same")
		}

		old := *field
		*field = "XXX"
		if hc.IsSame(other) || other.IsSame(hc) {
			t.Fatalf("should not be the same")
		}
		*field = old

		if !hc.IsSame(other) || !other.IsSame(hc) {
			t.Fatalf("should be the same")
		}
	}

	checkStringField := func(field *string) {
		if !hc.IsSame(other) || !other.IsSame(hc) {
			t.Fatalf("should be the same")
		}

		old := *field
		*field = "XXX"
		if hc.IsSame(other) || other.IsSame(hc) {
			t.Fatalf("should not be the same")
		}
		*field = old

		if !hc.IsSame(other) || !other.IsSame(hc) {
			t.Fatalf("should be the same")
		}
	}

	checkStringField(&other.Node)
	checkCheckIDField(&other.CheckID)
	checkStringField(&other.Name)
	checkStringField(&other.Status)
	checkStringField(&other.Notes)
	checkStringField(&other.Output)
	checkStringField(&other.ServiceID)
	checkStringField(&other.ServiceName)
}

func TestStructs_HealthCheck_Marshalling(t *testing.T) {
	d := &HealthCheckDefinition{}
	buf, err := d.MarshalJSON()
	require.NoError(t, err)
	require.NotContains(t, string(buf), `"Interval":""`)
	require.NotContains(t, string(buf), `"Timeout":""`)
	require.NotContains(t, string(buf), `"DeregisterCriticalServiceAfter":""`)
}

func TestStructs_HealthCheck_Clone(t *testing.T) {
	hc := &HealthCheck{
		Node:        "node1",
		CheckID:     "check1",
		Name:        "thecheck",
		Status:      api.HealthPassing,
		Notes:       "it's all good",
		Output:      "lgtm",
		ServiceID:   "service1",
		ServiceName: "theservice",
	}
	clone := hc.Clone()
	if !hc.IsSame(clone) {
		t.Fatalf("should be equal to its clone")
	}

	clone.Output = "different"
	if hc.IsSame(clone) {
		t.Fatalf("should not longer be equal to its clone")
	}
}

func TestStructs_CheckServiceNodes_Shuffle(t *testing.T) {
	// Make a huge list of nodes.
	var nodes CheckServiceNodes
	for i := 0; i < 100; i++ {
		nodes = append(nodes, CheckServiceNode{
			Node: &Node{
				Node:    fmt.Sprintf("node%d", i),
				Address: fmt.Sprintf("127.0.0.%d", i+1),
			},
		})
	}

	// Keep track of how many unique shuffles we get.
	uniques := make(map[string]struct{})
	for i := 0; i < 100; i++ {
		nodes.Shuffle()

		var names []string
		for _, node := range nodes {
			names = append(names, node.Node.Node)
		}
		key := strings.Join(names, "|")
		uniques[key] = struct{}{}
	}

	// We have to allow for the fact that there won't always be a unique
	// shuffle each pass, so we just look for smell here without the test
	// being flaky.
	if len(uniques) < 50 {
		t.Fatalf("unique shuffle ratio too low: %d/100", len(uniques))
	}
}

func TestStructs_CheckServiceNodes_Filter(t *testing.T) {
	nodes := CheckServiceNodes{
		CheckServiceNode{
			Node: &Node{
				Node:    "node1",
				Address: "127.0.0.1",
			},
			Checks: HealthChecks{
				&HealthCheck{
					Status: api.HealthWarning,
				},
			},
		},
		CheckServiceNode{
			Node: &Node{
				Node:    "node2",
				Address: "127.0.0.2",
			},
			Checks: HealthChecks{
				&HealthCheck{
					Status: api.HealthPassing,
				},
			},
		},
		CheckServiceNode{
			Node: &Node{
				Node:    "node3",
				Address: "127.0.0.3",
			},
			Checks: HealthChecks{
				&HealthCheck{
					Status: api.HealthCritical,
				},
			},
		},
		CheckServiceNode{
			Node: &Node{
				Node:    "node4",
				Address: "127.0.0.4",
			},
			Checks: HealthChecks{
				// This check has a different ID to the others to ensure it is not
				// ignored by accident
				&HealthCheck{
					CheckID: "failing2",
					Status:  api.HealthCritical,
				},
			},
		},
	}

	// Test the case where warnings are allowed.
	{
		twiddle := make(CheckServiceNodes, len(nodes))
		if n := copy(twiddle, nodes); n != len(nodes) {
			t.Fatalf("bad: %d", n)
		}
		filtered := twiddle.Filter(false)
		expected := CheckServiceNodes{
			nodes[0],
			nodes[1],
		}
		if !reflect.DeepEqual(filtered, expected) {
			t.Fatalf("bad: %v", filtered)
		}
	}

	// Limit to only passing checks.
	{
		twiddle := make(CheckServiceNodes, len(nodes))
		if n := copy(twiddle, nodes); n != len(nodes) {
			t.Fatalf("bad: %d", n)
		}
		filtered := twiddle.Filter(true)
		expected := CheckServiceNodes{
			nodes[1],
		}
		if !reflect.DeepEqual(filtered, expected) {
			t.Fatalf("bad: %v", filtered)
		}
	}

	// Allow failing checks to be ignored (note that the test checks have empty
	// CheckID which is valid).
	{
		twiddle := make(CheckServiceNodes, len(nodes))
		if n := copy(twiddle, nodes); n != len(nodes) {
			t.Fatalf("bad: %d", n)
		}
		filtered := twiddle.FilterIgnore(true, []types.CheckID{""})
		expected := CheckServiceNodes{
			nodes[0],
			nodes[1],
			nodes[2], // Node 3's critical check should be ignored.
			// Node 4 should still be failing since it's got a critical check with a
			// non-ignored ID.
		}
		if !reflect.DeepEqual(filtered, expected) {
			t.Fatalf("bad: %v", filtered)
		}
	}
}

func TestStructs_DirEntry_Clone(t *testing.T) {
	e := &DirEntry{
		LockIndex: 5,
		Key:       "hello",
		Flags:     23,
		Value:     []byte("this is a test"),
		Session:   "session1",
		RaftIndex: RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 2,
		},
	}

	clone := e.Clone()
	if !reflect.DeepEqual(e, clone) {
		t.Fatalf("bad: %v", clone)
	}

	e.Value = []byte("a new value")
	if reflect.DeepEqual(e, clone) {
		t.Fatalf("clone wasn't independent of the original")
	}
}

func TestStructs_ValidateServiceAndNodeMetadata(t *testing.T) {
	tooMuchMeta := make(map[string]string)
	for i := 0; i < metaMaxKeyPairs+1; i++ {
		tooMuchMeta[string(i)] = "value"
	}
	type testcase struct {
		Meta              map[string]string
		AllowConsulPrefix bool
		NodeError         string
		ServiceError      string
		GatewayError      string
	}
	cases := map[string]testcase{
		"should succeed": {
			map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			false,
			"",
			"",
			"",
		},
		"invalid key": {
			map[string]string{
				"": "value1",
			},
			false,
			"Couldn't load metadata pair",
			"Couldn't load metadata pair",
			"Couldn't load metadata pair",
		},
		"too many keys": {
			tooMuchMeta,
			false,
			"cannot contain more than",
			"cannot contain more than",
			"cannot contain more than",
		},
		"reserved key prefix denied": {
			map[string]string{
				metaKeyReservedPrefix + "key": "value1",
			},
			false,
			"reserved for internal use",
			"reserved for internal use",
			"reserved for internal use",
		},
		"reserved key prefix allowed": {
			map[string]string{
				metaKeyReservedPrefix + "key": "value1",
			},
			true,
			"",
			"",
			"",
		},
		"reserved key prefix allowed via whitelist just for gateway - " + MetaWANFederationKey: {
			map[string]string{
				MetaWANFederationKey: "value1",
			},
			false,
			"reserved for internal use",
			"reserved for internal use",
			"",
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Run("ValidateNodeMetadata", func(t *testing.T) {
				err := ValidateNodeMetadata(tc.Meta, tc.AllowConsulPrefix)
				if tc.NodeError == "" {
					require.NoError(t, err)
				} else {
					testutil.RequireErrorContains(t, err, tc.NodeError)
				}
			})
			t.Run("ValidateServiceMetadata - typical", func(t *testing.T) {
				err := ValidateServiceMetadata(ServiceKindTypical, tc.Meta, tc.AllowConsulPrefix)
				if tc.ServiceError == "" {
					require.NoError(t, err)
				} else {
					testutil.RequireErrorContains(t, err, tc.ServiceError)
				}
			})
			t.Run("ValidateServiceMetadata - mesh-gateway", func(t *testing.T) {
				err := ValidateServiceMetadata(ServiceKindMeshGateway, tc.Meta, tc.AllowConsulPrefix)
				if tc.GatewayError == "" {
					require.NoError(t, err)
				} else {
					testutil.RequireErrorContains(t, err, tc.GatewayError)
				}
			})
		})
	}
}

func TestStructs_validateMetaPair(t *testing.T) {
	longKey := strings.Repeat("a", metaKeyMaxLength+1)
	longValue := strings.Repeat("b", metaValueMaxLength+1)
	pairs := []struct {
		Key               string
		Value             string
		Error             string
		AllowConsulPrefix bool
		AllowConsulKeys   map[string]struct{}
	}{
		// valid pair
		{"key", "value", "", false, nil},
		// invalid, blank key
		{"", "value", "cannot be blank", false, nil},
		// allowed special chars in key name
		{"k_e-y", "value", "", false, nil},
		// disallowed special chars in key name
		{"(%key&)", "value", "invalid characters", false, nil},
		// key too long
		{longKey, "value", "Key is too long", false, nil},
		// reserved prefix
		{metaKeyReservedPrefix + "key", "value", "reserved for internal use", false, nil},
		// reserved prefix, allowed
		{metaKeyReservedPrefix + "key", "value", "", true, nil},
		// reserved prefix, not allowed via whitelist
		{metaKeyReservedPrefix + "bad", "value", "reserved for internal use", false, map[string]struct{}{metaKeyReservedPrefix + "good": struct{}{}}},
		// reserved prefix, allowed via whitelist
		{metaKeyReservedPrefix + "good", "value", "", true, map[string]struct{}{metaKeyReservedPrefix + "good": struct{}{}}},
		// value too long
		{"key", longValue, "Value is too long", false, nil},
	}

	for _, pair := range pairs {
		err := validateMetaPair(pair.Key, pair.Value, pair.AllowConsulPrefix, pair.AllowConsulKeys)
		if pair.Error == "" && err != nil {
			t.Fatalf("should have succeeded: %v, %v", pair, err)
		} else if pair.Error != "" && !strings.Contains(err.Error(), pair.Error) {
			t.Fatalf("should have failed: %v, %v", pair, err)
		}
	}
}

func TestSpecificServiceRequest_CacheInfo(t *testing.T) {
	tests := []struct {
		name     string
		req      ServiceSpecificRequest
		mutate   func(req *ServiceSpecificRequest)
		want     *cache.RequestInfo
		wantSame bool
	}{
		{
			name: "basic params",
			req: ServiceSpecificRequest{
				QueryOptions: QueryOptions{Token: "foo"},
				Datacenter:   "dc1",
			},
			want: &cache.RequestInfo{
				Token:      "foo",
				Datacenter: "dc1",
			},
			wantSame: true,
		},
		{
			name: "name should be considered",
			req: ServiceSpecificRequest{
				ServiceName: "web",
			},
			mutate: func(req *ServiceSpecificRequest) {
				req.ServiceName = "db"
			},
			wantSame: false,
		},
		{
			name: "node meta should be considered",
			req: ServiceSpecificRequest{
				NodeMetaFilters: map[string]string{
					"foo": "bar",
				},
			},
			mutate: func(req *ServiceSpecificRequest) {
				req.NodeMetaFilters = map[string]string{
					"foo": "qux",
				}
			},
			wantSame: false,
		},
		{
			name: "address should be considered",
			req: ServiceSpecificRequest{
				ServiceAddress: "1.2.3.4",
			},
			mutate: func(req *ServiceSpecificRequest) {
				req.ServiceAddress = "4.3.2.1"
			},
			wantSame: false,
		},
		{
			name: "tag filter should be considered",
			req: ServiceSpecificRequest{
				TagFilter: true,
			},
			mutate: func(req *ServiceSpecificRequest) {
				req.TagFilter = false
			},
			wantSame: false,
		},
		{
			name: "connect should be considered",
			req: ServiceSpecificRequest{
				Connect: true,
			},
			mutate: func(req *ServiceSpecificRequest) {
				req.Connect = false
			},
			wantSame: false,
		},
		{
			name: "tags should be different",
			req: ServiceSpecificRequest{
				ServiceName: "web",
				ServiceTags: []string{"foo"},
			},
			mutate: func(req *ServiceSpecificRequest) {
				req.ServiceTags = []string{"foo", "bar"}
			},
			wantSame: false,
		},
		{
			name: "tags should not depend on order",
			req: ServiceSpecificRequest{
				ServiceName: "web",
				ServiceTags: []string{"bar", "foo"},
			},
			mutate: func(req *ServiceSpecificRequest) {
				req.ServiceTags = []string{"foo", "bar"}
			},
			wantSame: true,
		},
		// DEPRECATED (singular-service-tag) - remove this when upgrade RPC compat
		// with 1.2.x is not required.
		{
			name: "legacy requests with singular tag should be different",
			req: ServiceSpecificRequest{
				ServiceName: "web",
				ServiceTag:  "foo",
			},
			mutate: func(req *ServiceSpecificRequest) {
				req.ServiceTag = "bar"
			},
			wantSame: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info := tc.req.CacheInfo()
			if tc.mutate != nil {
				tc.mutate(&tc.req)
			}
			afterInfo := tc.req.CacheInfo()

			// Check key matches or not
			if tc.wantSame {
				require.Equal(t, info, afterInfo)
			} else {
				require.NotEqual(t, info, afterInfo)
			}

			if tc.want != nil {
				// Reset key since we don't care about the actual hash value as long as
				// it does/doesn't change appropriately (asserted with wantSame above).
				info.Key = ""
				require.Equal(t, *tc.want, info)
			}
		})
	}
}

func TestNodeService_JSON_OmitTaggedAdddresses(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		ns   NodeService
	}{
		{
			"nil",
			NodeService{
				TaggedAddresses: nil,
			},
		},
		{
			"empty",
			NodeService{
				TaggedAddresses: make(map[string]ServiceAddress),
			},
		},
	}

	for _, tc := range cases {
		name := tc.name
		ns := tc.ns
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(ns)
			require.NoError(t, err)
			var raw map[string]interface{}
			err = json.Unmarshal(data, &raw)
			require.NoError(t, err)
			require.NotContains(t, raw, "TaggedAddresses")
			require.NotContains(t, raw, "tagged_addresses")
		})
	}
}

func TestServiceNode_JSON_OmitServiceTaggedAdddresses(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		sn   ServiceNode
	}{
		{
			"nil",
			ServiceNode{
				ServiceTaggedAddresses: nil,
			},
		},
		{
			"empty",
			ServiceNode{
				ServiceTaggedAddresses: make(map[string]ServiceAddress),
			},
		},
	}

	for _, tc := range cases {
		name := tc.name
		sn := tc.sn
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(sn)
			require.NoError(t, err)
			var raw map[string]interface{}
			err = json.Unmarshal(data, &raw)
			require.NoError(t, err)
			require.NotContains(t, raw, "ServiceTaggedAddresses")
			require.NotContains(t, raw, "service_tagged_addresses")
		})
	}
}

func TestNode_BestAddress(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input   Node
		lanAddr string
		wanAddr string
	}

	nodeAddr := "10.1.2.3"
	nodeWANAddr := "198.18.19.20"

	cases := map[string]testCase{
		"address": testCase{
			input: Node{
				Address: nodeAddr,
			},

			lanAddr: nodeAddr,
			wanAddr: nodeAddr,
		},
		"wan-address": testCase{
			input: Node{
				Address: nodeAddr,
				TaggedAddresses: map[string]string{
					"wan": nodeWANAddr,
				},
			},

			lanAddr: nodeAddr,
			wanAddr: nodeWANAddr,
		},
	}

	for name, tc := range cases {
		name := name
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.lanAddr, tc.input.BestAddress(false))
			require.Equal(t, tc.wanAddr, tc.input.BestAddress(true))
		})
	}
}

func TestNodeService_BestAddress(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input   NodeService
		lanAddr string
		lanPort int
		wanAddr string
		wanPort int
	}

	serviceAddr := "10.2.3.4"
	servicePort := 1234
	serviceWANAddr := "198.19.20.21"
	serviceWANPort := 987

	cases := map[string]testCase{
		"no-address": testCase{
			input: NodeService{
				Port: servicePort,
			},

			lanAddr: "",
			lanPort: servicePort,
			wanAddr: "",
			wanPort: servicePort,
		},
		"service-address": testCase{
			input: NodeService{
				Address: serviceAddr,
				Port:    servicePort,
			},

			lanAddr: serviceAddr,
			lanPort: servicePort,
			wanAddr: serviceAddr,
			wanPort: servicePort,
		},
		"service-wan-address": testCase{
			input: NodeService{
				Address: serviceAddr,
				Port:    servicePort,
				TaggedAddresses: map[string]ServiceAddress{
					"wan": ServiceAddress{
						Address: serviceWANAddr,
						Port:    serviceWANPort,
					},
				},
			},

			lanAddr: serviceAddr,
			lanPort: servicePort,
			wanAddr: serviceWANAddr,
			wanPort: serviceWANPort,
		},
		"service-wan-address-default-port": testCase{
			input: NodeService{
				Address: serviceAddr,
				Port:    servicePort,
				TaggedAddresses: map[string]ServiceAddress{
					"wan": ServiceAddress{
						Address: serviceWANAddr,
						Port:    0,
					},
				},
			},

			lanAddr: serviceAddr,
			lanPort: servicePort,
			wanAddr: serviceWANAddr,
			wanPort: servicePort,
		},
		"service-wan-address-node-lan": testCase{
			input: NodeService{
				Port: servicePort,
				TaggedAddresses: map[string]ServiceAddress{
					"wan": ServiceAddress{
						Address: serviceWANAddr,
						Port:    serviceWANPort,
					},
				},
			},

			lanAddr: "",
			lanPort: servicePort,
			wanAddr: serviceWANAddr,
			wanPort: serviceWANPort,
		},
	}

	for name, tc := range cases {
		name := name
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			addr, port := tc.input.BestAddress(false)
			require.Equal(t, tc.lanAddr, addr)
			require.Equal(t, tc.lanPort, port)

			addr, port = tc.input.BestAddress(true)
			require.Equal(t, tc.wanAddr, addr)
			require.Equal(t, tc.wanPort, port)
		})
	}
}

func TestCheckServiceNode_BestAddress(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input   CheckServiceNode
		lanAddr string
		lanPort int
		wanAddr string
		wanPort int
	}

	nodeAddr := "10.1.2.3"
	nodeWANAddr := "198.18.19.20"
	serviceAddr := "10.2.3.4"
	servicePort := 1234
	serviceWANAddr := "198.19.20.21"
	serviceWANPort := 987

	cases := map[string]testCase{
		"node-address": testCase{
			input: CheckServiceNode{
				Node: &Node{
					Address: nodeAddr,
				},
				Service: &NodeService{
					Port: servicePort,
				},
			},

			lanAddr: nodeAddr,
			lanPort: servicePort,
			wanAddr: nodeAddr,
			wanPort: servicePort,
		},
		"node-wan-address": testCase{
			input: CheckServiceNode{
				Node: &Node{
					Address: nodeAddr,
					TaggedAddresses: map[string]string{
						"wan": nodeWANAddr,
					},
				},
				Service: &NodeService{
					Port: servicePort,
				},
			},

			lanAddr: nodeAddr,
			lanPort: servicePort,
			wanAddr: nodeWANAddr,
			wanPort: servicePort,
		},
		"service-address": testCase{
			input: CheckServiceNode{
				Node: &Node{
					Address: nodeAddr,
					// this will be ignored
					TaggedAddresses: map[string]string{
						"wan": nodeWANAddr,
					},
				},
				Service: &NodeService{
					Address: serviceAddr,
					Port:    servicePort,
				},
			},

			lanAddr: serviceAddr,
			lanPort: servicePort,
			wanAddr: serviceAddr,
			wanPort: servicePort,
		},
		"service-wan-address": testCase{
			input: CheckServiceNode{
				Node: &Node{
					Address: nodeAddr,
					// this will be ignored
					TaggedAddresses: map[string]string{
						"wan": nodeWANAddr,
					},
				},
				Service: &NodeService{
					Address: serviceAddr,
					Port:    servicePort,
					TaggedAddresses: map[string]ServiceAddress{
						"wan": ServiceAddress{
							Address: serviceWANAddr,
							Port:    serviceWANPort,
						},
					},
				},
			},

			lanAddr: serviceAddr,
			lanPort: servicePort,
			wanAddr: serviceWANAddr,
			wanPort: serviceWANPort,
		},
		"service-wan-address-default-port": testCase{
			input: CheckServiceNode{
				Node: &Node{
					Address: nodeAddr,
					// this will be ignored
					TaggedAddresses: map[string]string{
						"wan": nodeWANAddr,
					},
				},
				Service: &NodeService{
					Address: serviceAddr,
					Port:    servicePort,
					TaggedAddresses: map[string]ServiceAddress{
						"wan": ServiceAddress{
							Address: serviceWANAddr,
							Port:    0,
						},
					},
				},
			},

			lanAddr: serviceAddr,
			lanPort: servicePort,
			wanAddr: serviceWANAddr,
			wanPort: servicePort,
		},
		"service-wan-address-node-lan": testCase{
			input: CheckServiceNode{
				Node: &Node{
					Address: nodeAddr,
					// this will be ignored
					TaggedAddresses: map[string]string{
						"wan": nodeWANAddr,
					},
				},
				Service: &NodeService{
					Port: servicePort,
					TaggedAddresses: map[string]ServiceAddress{
						"wan": ServiceAddress{
							Address: serviceWANAddr,
							Port:    serviceWANPort,
						},
					},
				},
			},

			lanAddr: nodeAddr,
			lanPort: servicePort,
			wanAddr: serviceWANAddr,
			wanPort: serviceWANPort,
		},
	}

	for name, tc := range cases {
		name := name
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			addr, port := tc.input.BestAddress(false)
			require.Equal(t, tc.lanAddr, addr)
			require.Equal(t, tc.lanPort, port)

			addr, port = tc.input.BestAddress(true)
			require.Equal(t, tc.wanAddr, addr)
			require.Equal(t, tc.wanPort, port)
		})
	}
}

func TestNodeService_JSON_Marshal(t *testing.T) {
	ns := &NodeService{
		Service: "foo",
		Proxy: ConnectProxyConfig{
			Config: map[string]interface{}{
				"bind_addresses": map[string]interface{}{
					"default": map[string]interface{}{
						"Address": "0.0.0.0",
						"Port":    "443",
					},
				},
			},
		},
	}
	buf, err := json.Marshal(ns)
	require.NoError(t, err)

	var out NodeService
	require.NoError(t, json.Unmarshal(buf, &out))
	require.Equal(t, *ns, out)
}

func TestServiceNode_JSON_Marshal(t *testing.T) {
	sn := &ServiceNode{
		Node:        "foo",
		ServiceName: "foo",
		ServiceProxy: ConnectProxyConfig{
			Config: map[string]interface{}{
				"bind_addresses": map[string]interface{}{
					"default": map[string]interface{}{
						"Address": "0.0.0.0",
						"Port":    "443",
					},
				},
			},
		},
	}
	buf, err := json.Marshal(sn)
	require.NoError(t, err)

	var out ServiceNode
	require.NoError(t, json.Unmarshal(buf, &out))
	require.Equal(t, *sn, out)
}

// frankensteinStruct is an amalgamation of all of the different kinds of
// fields you could have on struct defined in the agent/structs package that we
// send through msgpack
type frankensteinStruct struct {
	Child      *monsterStruct
	ChildSlice []*monsterStruct
	ChildMap   map[string]*monsterStruct
}
type monsterStruct struct {
	Bool    bool
	Int     int
	Uint8   uint8
	Uint64  uint64
	Float32 float32
	Float64 float64
	String  string

	Hash         []byte
	Uint32Slice  []uint32
	Float64Slice []float64
	StringSlice  []string

	MapInt         map[string]int
	MapString      map[string]string
	MapStringSlice map[string][]string

	// We explicitly DO NOT try to test the following types that involve
	// interface{} as the TestMsgpackEncodeDecode test WILL fail.
	//
	// These are tested elsewhere for the very specific scenario in question,
	// which usually takes a secondary trip through mapstructure during decode
	// which papers over some of the additional conversions necessary to finish
	// decoding.
	// MapIface    map[string]interface{}
	// MapMapIface map[string]map[string]interface{}

	Dur     time.Duration
	DurPtr  *time.Duration
	Time    time.Time
	TimePtr *time.Time

	RaftIndex
}

func makeFrank() *frankensteinStruct {
	return &frankensteinStruct{
		Child: makeMonster(),
		ChildSlice: []*monsterStruct{
			makeMonster(),
			makeMonster(),
		},
		ChildMap: map[string]*monsterStruct{
			"one": makeMonster(), // only put one key in here so the map order is fixed
		},
	}
}

func makeMonster() *monsterStruct {
	var d time.Duration = 9 * time.Hour
	var t time.Time = time.Date(2008, 1, 2, 3, 4, 5, 0, time.UTC)

	return &monsterStruct{
		Bool:    true,
		Int:     -8,
		Uint8:   5,
		Uint64:  9,
		Float32: 5.25,
		Float64: 99.5,
		String:  "strval",

		Hash:         []byte("hello"),
		Uint32Slice:  []uint32{1, 2, 3, 4},
		Float64Slice: []float64{9.2, 6.25},
		StringSlice:  []string{"foo", "bar"},

		// // MapIface will hold an amalgam of what AuthMethods and
		// // CAConfigurations use in 'Config'
		// MapIface: map[string]interface{}{
		// 	"Name":  "inner",
		// 	"Dur":   "5s",
		// 	"Bool":  true,
		// 	"Float": 15.25,
		// 	"Int":   int64(94),
		// 	"Nested": map[string]string{ // this doesn't survive
		// 		"foo": "bar",
		// 	},
		// },
		// // MapMapIface    map[string]map[string]interface{}

		MapInt: map[string]int{
			"int": 5,
		},
		MapString: map[string]string{
			"aaa": "bbb",
		},
		MapStringSlice: map[string][]string{
			"aaa": []string{"bbb"},
		},

		Dur:     5 * time.Second,
		DurPtr:  &d,
		Time:    t.Add(-5 * time.Hour),
		TimePtr: &t,

		RaftIndex: RaftIndex{
			CreateIndex: 1,
			ModifyIndex: 3,
		},
	}
}

func TestStructs_MsgpackEncodeDecode_Monolith(t *testing.T) {
	t.Run("monster", func(t *testing.T) {
		in := makeMonster()
		TestMsgpackEncodeDecode(t, in, false)
	})
	t.Run("frankenstein", func(t *testing.T) {
		in := makeFrank()
		TestMsgpackEncodeDecode(t, in, false)
	})
}

func TestSnapshotRequestResponse_MsgpackEncodeDecode(t *testing.T) {
	t.Run("request", func(t *testing.T) {
		in := &SnapshotRequest{
			Datacenter: "foo",
			Token:      "blah",
			AllowStale: true,
			Op:         SnapshotRestore,
		}
		TestMsgpackEncodeDecode(t, in, true)
	})
	t.Run("response", func(t *testing.T) {
		in := &SnapshotResponse{
			Error: "blah",
			QueryMeta: QueryMeta{
				Index:            3,
				LastContact:      5 * time.Second,
				KnownLeader:      true,
				ConsistencyLevel: "default",
			},
		}
		TestMsgpackEncodeDecode(t, in, true)
	})

}

func TestGatewayService_IsSame(t *testing.T) {
	gateway := NewServiceID("gateway", nil)
	svc := NewServiceID("web", nil)
	kind := ServiceKindTerminatingGateway
	ca := "ca.pem"
	cert := "client.pem"
	key := "tls.key"
	sni := "mydomain"
	wildcard := false

	g := &GatewayService{
		Gateway:      gateway,
		Service:      svc,
		GatewayKind:  kind,
		CAFile:       ca,
		CertFile:     cert,
		KeyFile:      key,
		SNI:          sni,
		FromWildcard: wildcard,
	}
	other := &GatewayService{
		Gateway:      gateway,
		Service:      svc,
		GatewayKind:  kind,
		CAFile:       ca,
		CertFile:     cert,
		KeyFile:      key,
		SNI:          sni,
		FromWildcard: wildcard,
	}
	check := func(twiddle, restore func()) {
		t.Helper()
		if !g.IsSame(other) || !other.IsSame(g) {
			t.Fatalf("should be the same")
		}

		twiddle()
		if g.IsSame(other) || other.IsSame(g) {
			t.Fatalf("should be different, was %#v VS %#v", g, other)
		}

		restore()
		if !g.IsSame(other) || !other.IsSame(g) {
			t.Fatalf("should be the same")
		}
	}
	check(func() { other.Gateway = NewServiceID("other", nil) }, func() { other.Gateway = gateway })
	check(func() { other.Service = NewServiceID("other", nil) }, func() { other.Service = svc })
	check(func() { other.GatewayKind = ServiceKindIngressGateway }, func() { other.GatewayKind = kind })
	check(func() { other.CAFile = "/certs/cert.pem" }, func() { other.CAFile = ca })
	check(func() { other.CertFile = "/certs/cert.pem" }, func() { other.CertFile = cert })
	check(func() { other.KeyFile = "/certs/cert.pem" }, func() { other.KeyFile = key })
	check(func() { other.SNI = "alt-domain" }, func() { other.SNI = sni })
	check(func() { other.FromWildcard = true }, func() { other.FromWildcard = wildcard })

	if !g.IsSame(other) {
		t.Fatalf("should be equal, was %#v VS %#v", g, other)
	}
}
