package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"singbox-launcher/core/config/subscription"
)

func loadConfigRootMap(configPath string) (map[string]interface{}, error) {
	cleanData, err := getConfigJSON(configPath)
	if err != nil {
		return nil, err
	}
	var root map[string]interface{}
	if err := json.Unmarshal(cleanData, &root); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return root, nil
}

func findTaggedInRoot(root map[string]interface{}, tag, arrayKey, notFoundFmt string) (map[string]interface{}, error) {
	rawList, ok := root[arrayKey].([]interface{})
	if !ok {
		return nil, fmt.Errorf("%s not found or invalid", arrayKey)
	}
	for _, raw := range rawList {
		om, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if t, _ := om["tag"].(string); t == tag {
			return om, nil
		}
	}
	return nil, fmt.Errorf(notFoundFmt, tag)
}

// GetOutboundMapByTag returns the raw outbound object from config.json outbounds[] with the given tag.
func GetOutboundMapByTag(configPath, tag string) (map[string]interface{}, error) {
	if tag == "" {
		return nil, fmt.Errorf("empty outbound tag")
	}
	root, err := loadConfigRootMap(configPath)
	if err != nil {
		return nil, err
	}
	return findTaggedInRoot(root, tag, "outbounds", "outbound with tag %q not found")
}

// GetEndpointMapByTag returns the raw endpoint object from config.json endpoints[] with the given tag (e.g. WireGuard).
func GetEndpointMapByTag(configPath, tag string) (map[string]interface{}, error) {
	if tag == "" {
		return nil, fmt.Errorf("empty endpoint tag")
	}
	root, err := loadConfigRootMap(configPath)
	if err != nil {
		return nil, err
	}
	return findTaggedInRoot(root, tag, "endpoints", "endpoint with tag %q not found")
}

// shareURITryEndpointAfterOutboundError is true when the tag is missing from outbounds (or outbounds absent), so we may resolve WireGuard in endpoints[].
func shareURITryEndpointAfterOutboundError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "not found") || strings.Contains(s, "outbounds not found")
}

// ShareProxyURIForOutboundTag builds a subscription-style share URI from the sing-box outbound with the given tag,
// or from a WireGuard entry in endpoints[] with that tag if no matching outbound exists.
// Parses config.json once per call.
func ShareProxyURIForOutboundTag(configPath, tag string) (string, error) {
	if tag == "" {
		return "", fmt.Errorf("empty outbound tag")
	}
	root, err := loadConfigRootMap(configPath)
	if err != nil {
		return "", err
	}
	out, outErr := findTaggedInRoot(root, tag, "outbounds", "outbound with tag %q not found")
	if outErr == nil {
		return subscription.ShareURIFromOutbound(out)
	}
	if shareURITryEndpointAfterOutboundError(outErr) {
		ep, epErr := findTaggedInRoot(root, tag, "endpoints", "endpoint with tag %q not found")
		if epErr == nil {
			return subscription.ShareURIFromWireGuardEndpoint(ep)
		}
	}
	return "", outErr
}
