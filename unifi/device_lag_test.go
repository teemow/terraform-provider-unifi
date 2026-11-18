package unifi

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ubiquiti-community/go-unifi/unifi"
)

// A LAG as a USW Aggregation reports it: the lead port carries op_mode
// "aggregate", its members and the group number.
func lagPort(idx, lag int64, name string) unifi.DevicePortOverrides {
	return unifi.DevicePortOverrides{
		PortIDX:          ptrInt64(idx),
		Name:             name,
		OpMode:           "aggregate",
		Forward:          "customize",
		NATiveNetworkID:  "0a1b2c3d-1111-4222-8333-444455556666",
		AggregateMembers: []int64{idx, idx + 1},
		LagIdx:           ptrInt64(lag),
	}
}

func TestFrameworkToPortOverrides_LagIdx(t *testing.T) {
	r := &deviceResource{}

	set := portOverrideSetWith(t, map[string]attr.Value{
		"index":   types.Int64Value(1),
		"op_mode": types.StringValue("aggregate"),
		"lag_idx": types.Int64Value(1),
	})
	pos, diags := r.frameworkToPortOverrides(context.Background(), set)
	if diags.HasError() {
		t.Fatalf("frameworkToPortOverrides errored: %v", diags)
	}
	if pos[0].LagIdx == nil || *pos[0].LagIdx != 1 {
		t.Errorf("LagIdx = %v, want 1", pos[0].LagIdx)
	}

	unset := portOverrideSetWith(t, map[string]attr.Value{"index": types.Int64Value(7)})
	pos, diags = r.frameworkToPortOverrides(context.Background(), unset)
	if diags.HasError() {
		t.Fatalf("frameworkToPortOverrides errored: %v", diags)
	}
	b, err := json.Marshal(pos[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"lag_idx"`) {
		t.Errorf("unset: lag_idx must not be sent: %s", b)
	}
}

// An aggregate port written without its lag_idx falls back to op_mode
// "switch" on the controller. A declared port that leaves lag_idx out therefore
// keeps the controller's, and a declared one wins.
func TestMergePortOverridesByIndex_LagIdx(t *testing.T) {
	current := lagPort(1, 1, "node-a LACP")

	for _, tt := range []struct {
		name     string
		declared *int64
		want     int64
	}{
		{name: "unset carries controller value", declared: nil, want: 1},
		{name: "configured value is sent", declared: ptrInt64(4), want: 4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			declared := lagPort(1, 1, "node-a LACP")
			declared.LagIdx = tt.declared
			got := mergePortOverridesByIndex(
				[]unifi.DevicePortOverrides{current},
				[]unifi.DevicePortOverrides{declared},
			)
			if got[0].LagIdx == nil || *got[0].LagIdx != tt.want {
				t.Errorf("LagIdx = %v, want %d", got[0].LagIdx, tt.want)
			}
		})
	}
}

// Renaming one port of a switch sends the whole port_overrides array. The LAG
// ports in it must go out exactly as the controller holds them.
func TestMergePortOverridesByIndex_changeElsewhereKeepsLAG(t *testing.T) {
	r := &deviceResource{}
	current := []unifi.DevicePortOverrides{
		lagPort(1, 1, "node-a LACP"),
		lagPort(3, 2, "node-b LACP"),
		{
			PortIDX:         ptrInt64(7),
			Name:            "node-c",
			Forward:         "customize",
			NATiveNetworkID: "0a1b2c3d-1111-4222-8333-444455556666",
		},
	}

	block := func(idx, lag int64, name string) attr.Value {
		over := map[string]attr.Value{
			"index":                 types.Int64Value(idx),
			"name":                  types.StringValue(name),
			"forward":               types.StringValue("customize"),
			"native_networkconf_id": types.StringValue("0a1b2c3d-1111-4222-8333-444455556666"),
			"op_mode":               types.StringValue("switch"),
		}
		if lag > 0 {
			over["op_mode"] = types.StringValue("aggregate")
			over["lag_idx"] = types.Int64Value(lag)
			over["aggregate_members"] = types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(idx), types.Int64Value(idx + 1),
			})
		}
		return portOverrideSetWith(t, over).Elements()[0]
	}
	set := types.SetValueMust(
		types.ObjectType{AttrTypes: portOverrideAttrTypes()},
		[]attr.Value{
			block(1, 1, "node-a LACP"),
			block(3, 2, "node-b LACP"),
			block(7, 0, "node-c renamed"),
		},
	)
	declared, diags := r.frameworkToPortOverrides(context.Background(), set)
	if diags.HasError() {
		t.Fatalf("frameworkToPortOverrides errored: %v", diags)
	}

	got := mergePortOverridesByIndex(current, declared)
	if len(got) != 3 {
		t.Fatalf("merged %d ports, want 3", len(got))
	}
	for i := range 2 {
		if !reflect.DeepEqual(got[i], current[i]) {
			t.Errorf("LAG port %d = %+v, want the controller's entry %+v", i, got[i], current[i])
		}
	}
	if got[2].Name != "node-c renamed" {
		t.Errorf("renamed port = %q", got[2].Name)
	}
}

func TestReconcilePortOverrides_LagIdx(t *testing.T) {
	r := &deviceResource{}
	reconcile := func(prior types.Int64, api *int64) types.Int64 {
		t.Helper()
		got, diags := r.reconcilePortOverrides(
			context.Background(),
			portOverrideSetWith(t, map[string]attr.Value{
				"index":   types.Int64Value(1),
				"lag_idx": prior,
			}),
			[]unifi.DevicePortOverrides{{PortIDX: ptrInt64(1), LagIdx: api}},
		)
		if diags.HasError() {
			t.Fatalf("reconcilePortOverrides errored: %v", diags)
		}
		var models []portOverrideModel
		if d := got.ElementsAs(context.Background(), &models, false); d.HasError() {
			t.Fatalf("reading set: %v", d)
		}
		return models[0].LagIdx
	}

	if got := reconcile(types.Int64Value(1), ptrInt64(1)); got.ValueInt64() != 1 {
		t.Errorf("configured value read back as %v, want 1", got)
	}
	// A group the controller dropped shows up as drift.
	if got := reconcile(types.Int64Value(1), nil); !got.IsNull() {
		t.Errorf("drift not surfaced: got %v, want null", got)
	}
	// Unset in config stays null whatever the controller holds.
	if got := reconcile(types.Int64Null(), ptrInt64(2)); !got.IsNull() {
		t.Errorf("unset lag_idx read back as %v, want null", got)
	}
}
