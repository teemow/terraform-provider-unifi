package util

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ParseImportID parses import IDs supporting both "id" and "site:id" formats.
func ParseImportID(id string, minParts int, maxParts int) (map[string]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	resp := map[string]string{}

	parts := strings.SplitN(id, ":", maxParts)

	if len(parts) < minParts || len(parts) > maxParts {
		diags.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in the format 'id' or 'site:id', got: %s", id),
		)
		return nil, diags
	}

	if len(parts) == 2 {
		slices.Reverse(parts)
		resp["site"] = parts[1]
	}

	resp["id"] = parts[0]

	return resp, diags
}

// objectIDPattern matches the two ID formats a UniFi controller hands out: the
// 24-hex ObjectID of Network 10.x and older, and the UUID of Network 11, which
// re-keys every existing object on upgrade (the old ID stays readable as
// `legacy_id`).
var objectIDPattern = regexp.MustCompile(
	`^(?:[0-9a-f]{24}|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`,
)

// IsObjectID reports whether s is a controller object ID, as opposed to a name.
func IsObjectID(s string) bool {
	return objectIDPattern.MatchString(s)
}
