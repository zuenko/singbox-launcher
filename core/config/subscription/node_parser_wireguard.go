package subscription

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"singbox-launcher/core/config/configtypes"
	"singbox-launcher/internal/debuglog"
)

// parseWireGuardURI parses wireguard:// URI into ParsedNode with sing-box endpoint in Outbound.
// Format: wireguard://<PRIVATE_KEY>@<SERVER_IP>:<PORT>?publickey=...&address=...&allowedips=...
// Required query: publickey, address, allowedips. Optional: mtu, keepalive, presharedkey, listenport, name, dns.
func parseWireGuardURI(uri string, skipFilters []map[string]string) (*configtypes.ParsedNode, error) {
	debuglog.DebugLog("parseWireGuardURI: start")
	if len(uri) > MaxURILength {
		debuglog.DebugLog("parseWireGuardURI: error URI length exceeded")
		return nil, fmt.Errorf("URI length (%d) exceeds maximum (%d)", len(uri), MaxURILength)
	}
	// Extract fragment from raw URI; url.Parse may not set Fragment for non-standard schemes.
	fragmentFromRaw := ""
	if i := strings.LastIndex(uri, "#"); i >= 0 {
		fragmentFromRaw = strings.TrimSpace(uri[i+1:])
	}
	parsedURL, err := url.Parse(uri)
	if err != nil {
		debuglog.DebugLog("parseWireGuardURI: error parse URL: %v", err)
		return nil, fmt.Errorf("failed to parse wireguard URI: %w", err)
	}
	if parsedURL.Hostname() == "" {
		debuglog.DebugLog("parseWireGuardURI: error missing hostname")
		return nil, fmt.Errorf("invalid wireguard URI: missing hostname")
	}
	if parsedURL.User == nil || parsedURL.User.Username() == "" {
		debuglog.DebugLog("parseWireGuardURI: error missing private key (userinfo)")
		return nil, fmt.Errorf("invalid wireguard URI: missing private key (userinfo)")
	}
	// Use PathUnescape so + in base64 is preserved (QueryUnescape would turn + into space and break the key)
	privateKey, err := url.PathUnescape(parsedURL.User.Username())
	if err != nil {
		privateKey = parsedURL.User.Username()
	}
	privateKey = strings.TrimSpace(privateKey)
	if privateKey == "" {
		return nil, fmt.Errorf("invalid wireguard URI: empty private key")
	}
	// Validate base64 private key (optional but recommended)
	if _, err := base64.StdEncoding.DecodeString(privateKey); err != nil {
		if _, err2 := base64.URLEncoding.DecodeString(privateKey); err2 != nil {
			debuglog.DebugLog("parseWireGuardURI: warning private key may not be valid base64")
		}
	}

	port := 51820
	if p := parsedURL.Port(); p != "" {
		if pi, err := strconv.Atoi(p); err == nil {
			port = pi
		}
	}

	q := parsedURL.Query()
	// Preserve + in base64 (query parser would decode + as space)
	publicKey := queryParamPreservePlus(parsedURL, "publickey")
	if publicKey == "" {
		publicKey = q.Get("publickey")
	}
	addressParam := q.Get("address")
	allowedipsParam := q.Get("allowedips")
	if publicKey == "" {
		debuglog.DebugLog("parseWireGuardURI: error missing publickey")
		return nil, fmt.Errorf("invalid wireguard URI: missing required query parameter publickey")
	}
	if addressParam == "" {
		debuglog.DebugLog("parseWireGuardURI: error missing address")
		return nil, fmt.Errorf("invalid wireguard URI: missing required query parameter address")
	}
	if allowedipsParam == "" {
		debuglog.DebugLog("parseWireGuardURI: error missing allowedips")
		return nil, fmt.Errorf("invalid wireguard URI: missing required query parameter allowedips")
	}

	addressDecoded, _ := url.QueryUnescape(addressParam)
	allowedipsDecoded, _ := url.QueryUnescape(allowedipsParam)
	addressList := splitAndTrim(addressDecoded, ",")
	allowedipsList := splitAndTrim(allowedipsDecoded, ",")
	if len(addressList) == 0 || len(allowedipsList) == 0 {
		return nil, fmt.Errorf("invalid wireguard URI: address or allowedips empty after parse")
	}

	mtu := 1420
	if m := q.Get("mtu"); m != "" {
		if mi, err := strconv.Atoi(m); err == nil {
			mtu = mi
		}
	}
	listenport := 0
	if lp := q.Get("listenport"); lp != "" {
		if lpi, err := strconv.Atoi(lp); err == nil {
			listenport = lpi
		}
	}
	name := q.Get("name")
	if name == "" {
		name = "singbox-wg0"
	}
	if decoded, err := url.QueryUnescape(name); err == nil {
		name = decoded
	}

	peer := map[string]interface{}{
		"address":     parsedURL.Hostname(),
		"port":        port,
		"public_key":  publicKey,
		"allowed_ips": allowedipsList,
	}
	if keepalive := q.Get("keepalive"); keepalive != "" {
		if ki, err := strconv.Atoi(keepalive); err == nil {
			peer["persistent_keepalive_interval"] = ki
		}
	}
	if psk := queryParamPreservePlus(parsedURL, "presharedkey"); psk != "" {
		peer["pre_shared_key"] = psk
	} else if psk := q.Get("presharedkey"); psk != "" {
		peer["pre_shared_key"] = psk
	}

	endpoint := map[string]interface{}{
		"type":        "wireguard",
		"tag":         "", // set below after tag is computed
		"name":        name,
		"system":      false,
		"mtu":         mtu,
		"address":     addressList,
		"private_key": privateKey,
		"peers":       []map[string]interface{}{peer},
	}
	if listenport != 0 {
		endpoint["listen_port"] = listenport
	}

	label := parsedURL.Fragment
	if label == "" && fragmentFromRaw != "" {
		label = fragmentFromRaw
	}
	if label == "" {
		label = name
	}
	if decoded, err := url.QueryUnescape(label); err == nil {
		label = decoded
	}
	label = sanitizeForDisplay(label)
	tag, comment := extractTagAndComment(label)
	if tag == "" {
		tag = generateDefaultTag("wireguard", parsedURL.Hostname(), port)
		comment = tag
	}
	tag = normalizeFlagTag(tag)
	endpoint["tag"] = tag

	node := &configtypes.ParsedNode{
		Scheme:   "wireguard",
		Tag:      tag,
		Server:   parsedURL.Hostname(),
		Port:     port,
		Label:    label,
		Comment:  comment,
		Query:    q,
		Outbound: endpoint,
	}

	if shouldSkipNode(node, skipFilters) {
		return nil, nil
	}
	debuglog.DebugLog("parseWireGuardURI: success tag=%s", node.Tag)
	return node, nil
}

// queryParamPreservePlus returns the first value for key in u.RawQuery, decoded with PathUnescape.
// This preserves '+' in base64 (QueryUnescape decodes '+' as space and would break keys).
func queryParamPreservePlus(u *url.URL, key string) string {
	for _, pair := range strings.Split(u.RawQuery, "&") {
		if i := strings.Index(pair, "="); i >= 0 {
			k := strings.TrimSpace(pair[:i])
			if k != key {
				continue
			}
			val := pair[i+1:]
			if d, err := url.PathUnescape(val); err == nil {
				return d
			}
			return val
		}
	}
	return ""
}

// splitAndTrim splits a string by separator, trims whitespace from each part,
// and returns only non-empty parts.
func splitAndTrim(s string, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
