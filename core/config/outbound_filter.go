// Package config: outbound_filter.go — filtering logic for selector outbounds.
//
// Functions here determine which nodes match a selector's filters (tag, host, scheme, label, etc.).
// Supports literal match, negation !literal, regex /pattern/i, negation regex !/pattern/i.
// Used by outbound_generator.go (GenerateSelectorWithFilteredAddOutbounds, buildOutboundsInfo)
// and by PreviewSelectorNodes for UI preview.
package config

import (
	"regexp"
	"strings"

	"singbox-launcher/internal/debuglog"
)

// filterNodesForSelector returns nodes that match the filter. filter may be nil (all nodes),
// a single map (AND of key/pattern), or a slice of maps (OR of maps). Empty map = no filter.
func filterNodesForSelector(allNodes []*ParsedNode, filter interface{}) []*ParsedNode {
	if filter == nil {
		return allNodes // No filter, return all nodes
	}

	// Check if filter is an empty map - treat as no filter
	if filterMap, ok := filter.(map[string]interface{}); ok {
		if len(filterMap) == 0 {
			return allNodes // Empty filter object means no filter, return all nodes
		}
	}

	filtered := make([]*ParsedNode, 0)

	// Check if filter is an array
	if filterArray, ok := filter.([]interface{}); ok {
		// OR between filter objects
		for _, node := range allNodes {
			for _, filterObj := range filterArray {
				if filterMap, ok := filterObj.(map[string]interface{}); ok {
					filterStrMap := convertFilterToStringMap(filterMap)
					if matchesFilter(node, filterStrMap) {
						filtered = append(filtered, node)
						break // Node matched at least one filter, add it
					}
				}
			}
		}
	} else if filterMap, ok := filter.(map[string]interface{}); ok {
		// Single filter object (AND between keys)
		filterStrMap := convertFilterToStringMap(filterMap)
		for _, node := range allNodes {
			if matchesFilter(node, filterStrMap) {
				filtered = append(filtered, node)
			}
		}
	}

	return filtered
}

// convertFilterToStringMap flattens filter map to string values for matching (non-string values are skipped).
func convertFilterToStringMap(filter map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range filter {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}

// matchesFilter returns true if the node has matching values for every key in filter (AND); each value is checked with matchesPattern.
func matchesFilter(node *ParsedNode, filter map[string]string) bool {
	for key, pattern := range filter {
		value := getNodeValue(node, key)
		if !matchesPattern(value, pattern) {
			return false // At least one key doesn't match
		}
	}
	return true // All keys match
}

// getNodeValue returns the node field used in filters: tag, host, label, scheme, fragment (alias for label), comment.
func getNodeValue(node *ParsedNode, key string) string {
	switch key {
	case "tag":
		return node.Tag
	case "host":
		return node.Server
	case "label":
		return node.Label
	case "scheme":
		return node.Scheme
	case "fragment":
		return node.Label // fragment == label
	case "comment":
		return node.Comment
	default:
		return ""
	}
}

// matchesPattern matches value against pattern: literal, !literal, /regex/i, !/regex/i. Case-insensitive for regex.
func matchesPattern(value, pattern string) bool {
	// Negation literal: !literal
	if strings.HasPrefix(pattern, "!") && !strings.HasPrefix(pattern, "!/") {
		literal := strings.TrimPrefix(pattern, "!")
		return value != literal
	}

	// Negation regex: !/regex/i
	if strings.HasPrefix(pattern, "!/") && strings.HasSuffix(pattern, "/i") {
		regexStr := strings.TrimPrefix(pattern, "!/")
		regexStr = strings.TrimSuffix(regexStr, "/i")
		re, err := regexp.Compile("(?i)" + regexStr)
		if err != nil {
			debuglog.WarnLog("Parser: Invalid regex pattern %s: %v", pattern, err)
			return false
		}
		return !re.MatchString(value)
	}

	// Regex: /regex/i
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/i") {
		regexStr := strings.TrimPrefix(pattern, "/")
		regexStr = strings.TrimSuffix(regexStr, "/i")
		re, err := regexp.Compile("(?i)" + regexStr)
		if err != nil {
			debuglog.WarnLog("Parser: Invalid regex pattern %s: %v", pattern, err)
			return false
		}
		return re.MatchString(value)
	}

	// Literal match
	return value == pattern
}

// PreviewSelectorNodes returns nodes that match outboundConfig.Filters and the default tag
// based on outboundConfig.PreferredDefault. It is used by UI layers to build a selector
// preview that is consistent with the real selector generation logic.
//
// allNodes must be the same set of nodes that will be used for selector generation
// (i.e. result of the same LoadNodesFromSource pipeline that GenerateOutboundsFromParserConfig uses).
func PreviewSelectorNodes(allNodes []*ParsedNode, outboundConfig OutboundConfig) ([]*ParsedNode, string) {
	filtered := filterNodesForSelector(allNodes, outboundConfig.Filters)

	defaultTag := ""
	if len(outboundConfig.PreferredDefault) > 0 {
		preferredFilter := convertFilterToStringMap(outboundConfig.PreferredDefault)
		for _, node := range filtered {
			if matchesFilter(node, preferredFilter) {
				defaultTag = node.Tag
				break
			}
		}
	}

	return filtered, defaultTag
}
