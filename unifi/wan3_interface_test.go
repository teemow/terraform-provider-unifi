package unifi

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// A gateway with three WAN ports (UCG Ultra) names the third one wan3. A port
// forward bound to another interface is stored and then ignored by the
// gateway, so the interface has to be expressible.
func TestWAN3InterfaceIsAccepted(t *testing.T) {
	ctx := context.Background()

	var ddns fwresource.SchemaResponse
	(&dynamicDNSResource{}).Schema(ctx, fwresource.SchemaRequest{}, &ddns)
	var forward fwresource.SchemaResponse
	(&portForwardResource{}).Schema(ctx, fwresource.SchemaRequest{}, &forward)

	wan, ok := forward.Schema.Attributes["wan"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("unifi_port_forward.wan is %T", forward.Schema.Attributes["wan"])
	}
	destinations, ok := forward.Schema.Attributes["destination_ips"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatalf(
			"unifi_port_forward.destination_ips is %T",
			forward.Schema.Attributes["destination_ips"],
		)
	}

	for name, attribute := range map[string]schema.Attribute{
		"unifi_dynamic_dns.interface":                  ddns.Schema.Attributes["interface"],
		"unifi_port_forward.wan.interface":             wan.Attributes["interface"],
		"unifi_port_forward.destination_ips.interface": destinations.NestedObject.Attributes["interface"],
	} {
		t.Run(name, func(t *testing.T) {
			str, ok := attribute.(schema.StringAttribute)
			if !ok {
				t.Fatalf("attribute is %T", attribute)
			}
			for value, valid := range map[string]bool{"wan": true, "wan2": true, "wan3": true, "lan": false} {
				var resp validator.StringResponse
				for _, v := range str.Validators {
					v.ValidateString(ctx, validator.StringRequest{
						Path:        path.Root("interface"),
						ConfigValue: types.StringValue(value),
					}, &resp)
				}
				if got := !resp.Diagnostics.HasError(); got != valid {
					t.Errorf("%q accepted = %v, want %v: %v", value, got, valid, resp.Diagnostics)
				}
			}
		})
	}
}
