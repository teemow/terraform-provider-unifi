package unifi

import "github.com/hashicorp/terraform-plugin-framework/resource"

// controllerIDIdentity is the behavior of every resource whose identity is the
// controller's object ID. The controller can change that ID: the upgrade to
// Network 11 re-keys every object from a 24-hex ObjectID to a UUID and still
// answers a read by the old ID, returning the new one. Without a mutable
// identity the framework fails that read with "Unexpected Identity Change";
// with it, the refresh moves state and identity to the new ID.
var controllerIDIdentity = resource.ResourceBehavior{MutableIdentity: true}
