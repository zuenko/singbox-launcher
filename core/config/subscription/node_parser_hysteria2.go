package subscription

import (
	"strconv"
	"strings"

	"singbox-launcher/core/config/configtypes"
	"singbox-launcher/internal/debuglog"
)

// isValidHysteria2ObfsType checks if the obfs type is supported by sing-box for Hysteria2
// According to sing-box documentation, only "salamander" is supported
func isValidHysteria2ObfsType(obfsType string) bool {
	return obfsType == "salamander"
}

// buildHysteria2Outbound builds outbound configuration for Hysteria2 protocol
func buildHysteria2Outbound(node *configtypes.ParsedNode, outbound map[string]interface{}) {
	// Password is required (stored in UUID field from userinfo)
	if node.UUID != "" {
		outbound["password"] = node.UUID
	} else {
		debuglog.WarnLog("Parser: Hysteria2 link missing password. URI might be invalid.")
	}

	// Optional: mport / ports query — Hysteria2 multi-port (comma-separated ports and ranges, hyphen in URI).
	// See https://v2.hysteria.network/docs/advanced/Port-Hopping/
	mport := strings.TrimSpace(queryGetFold(node.Query, "mport"))
	if mport == "" {
		mport = strings.TrimSpace(queryGetFold(node.Query, "ports"))
	}
	if sp := hysteria2MportSpecToSingBoxServerPorts(mport); len(sp) > 0 {
		outbound["server_ports"] = sp
	}

	// Optional: obfs (obfuscation)
	if obfs := node.Query.Get("obfs"); obfs != "" {
		if !isValidHysteria2ObfsType(obfs) {
			debuglog.WarnLog("Parser: Invalid or unsupported Hysteria2 obfs type '%s'. Only 'salamander' is supported. Skipping obfs.", obfs)
		} else {
			obfsConfig := map[string]interface{}{
				"type": obfs,
			}
			if obfsPassword := node.Query.Get("obfs-password"); obfsPassword != "" {
				obfsConfig["password"] = obfsPassword
			}
			outbound["obfs"] = obfsConfig
		}
	}

	// Optional: bandwidth (up/down in Mbps)
	if up := node.Query.Get("upmbps"); up != "" {
		if upMBps, err := strconv.Atoi(up); err == nil {
			outbound["up_mbps"] = upMBps
		}
	}
	if down := node.Query.Get("downmbps"); down != "" {
		if downMBps, err := strconv.Atoi(down); err == nil {
			outbound["down_mbps"] = downMBps
		}
	}

	// TLS settings (required for hysteria2)
	buildHysteria2TLS(node, outbound)
}

// buildHysteria2TLS builds TLS configuration for Hysteria2
func buildHysteria2TLS(node *configtypes.ParsedNode, outbound map[string]interface{}) {
	q := node.Query
	sni := queryGetFold(q, "sni")

	// Always enable TLS for hysteria2 (required by protocol)
	tlsData := map[string]interface{}{
		"enabled": true,
	}

	// Set SNI if provided and valid (skip emoji or invalid values)
	// SNI is valid if it contains dot (hostname) or colon (IPv6)
	if sni != "" && sni != "🔒" && (strings.Contains(sni, ".") || strings.Contains(sni, ":")) {
		tlsData["server_name"] = sni
	} else if node.Server != "" {
		tlsData["server_name"] = node.Server
	}

	if tlsInsecureTrue(q) {
		tlsData["insecure"] = true
	} else if queryGetFold(q, "skip-cert-verify") == "true" || queryGetFold(q, "skip-cert-verify") == "1" {
		tlsData["insecure"] = true
	}

	fp := NormalizeUTLSFingerprint(queryGetFold(q, "fp"))
	if fp == "" {
		fp = NormalizeUTLSFingerprint(queryGetFold(q, "fingerprint"))
	}
	if fp != "" {
		tlsData["utls"] = map[string]interface{}{
			"enabled":     true,
			"fingerprint": fp,
		}
	}

	if pin := strings.TrimSpace(queryGetFold(q, "pinSHA256")); pin != "" {
		tlsData["certificate_public_key_sha256"] = []string{pin}
	}

	// Handle ALPN parameter (for hysteria2, typically "h3")
	if alpn := queryGetFold(q, "alpn"); alpn != "" {
		alpnList := strings.Split(alpn, ",")
		for i := range alpnList {
			alpnList[i] = strings.TrimSpace(alpnList[i])
		}
		tlsData["alpn"] = alpnList
	}

	outbound["tls"] = tlsData
}
