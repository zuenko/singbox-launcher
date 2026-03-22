package models

import "encoding/json"

// PersistedDNSState is the dns_options object in state.json (not sing-box config.dns).
// See docs/WIZARD_STATE.md and SPECS/024-F-C-WIZARD_DNS_SECTION/SPEC.md.
// Each element of Servers may include wizard-only keys: "description" (string), "enabled" (bool, default true).
//
// DNS rules: JSON array **`rules`** (same as sing-box dns.rules / wizard_template dns_options.rules).
// The wizard editor uses multiline text; on save it is parsed into **`rules`**.
type PersistedDNSState struct {
	Servers          []json.RawMessage `json:"servers"`
	Rules            []json.RawMessage `json:"rules,omitempty"`
	Final            string            `json:"final"`
	Strategy         string            `json:"strategy,omitempty"`
	IndependentCache *bool             `json:"independent_cache,omitempty"`
	// DefaultDomainResolver mirrors model route default resolver tag when set (round-trip with UI).
	DefaultDomainResolver string `json:"default_domain_resolver,omitempty"`
	ResolverUnset         bool   `json:"default_domain_resolver_unset,omitempty"`
}
