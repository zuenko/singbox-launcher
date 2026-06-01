// Package xray provides Xray-core sidecar support for XHTTP transport.
package xray

import (
	"encoding/json"
	"fmt"
)

// BuildSidecarConfig converts a sing-box VLESS+XHTTP outbound into a minimal
// Xray JSON config with a SOCKS5 inbound and the given outbound.
// port is the local SOCKS5 listen port (e.g. 15080).
func BuildSidecarConfig(singboxOutbound map[string]interface{}, port int) ([]byte, error) {
	outbound, err := convertOutbound(singboxOutbound)
	if err != nil {
		return nil, fmt.Errorf("convert outbound: %w", err)
	}

	cfg := map[string]interface{}{
		"log": map[string]interface{}{
			"loglevel": "warning",
		},
		"inbounds": []map[string]interface{}{
			{
				"tag":      "socks-in",
				"protocol": "socks",
				"listen":   "127.0.0.1",
				"port":     port,
				"settings": map[string]interface{}{
					"udp": true,
				},
			},
		},
		"outbounds": []map[string]interface{}{
			outbound,
			{
				"tag":      "direct",
				"protocol": "freedom",
			},
		},
	}

	return json.MarshalIndent(cfg, "", "  ")
}

// convertOutbound transforms a sing-box outbound map into an Xray outbound map.
// Currently supports VLESS with xhttp transport.
func convertOutbound(sb map[string]interface{}) (map[string]interface{}, error) {
	typ := mapGetString(sb, "type")
	if typ != "vless" {
		return nil, fmt.Errorf("unsupported outbound type %q (only vless supported)", typ)
	}

	server := mapGetString(sb, "server")
	port := mapGetInt(sb, "server_port")
	uuid := mapGetString(sb, "uuid")
	if server == "" || port == 0 || uuid == "" {
		return nil, fmt.Errorf("missing required vless fields: server=%q port=%d uuid=%q", server, port, uuid)
	}

	vlessSettings := map[string]interface{}{
		"vnext": []map[string]interface{}{
			{
				"address": server,
				"port":    port,
				"users": []map[string]interface{}{
					{
						"id":         uuid,
						"encryption": "none",
						"flow":       mapGetString(sb, "flow"),
					},
				},
			},
		},
	}

	streamSettings := map[string]interface{}{}

	// Transport
	if tr, ok := sb["transport"].(map[string]interface{}); ok {
		transportType := mapGetString(tr, "type")
		switch transportType {
		case "xhttp":
			streamSettings["network"] = "xhttp"
			xh := map[string]interface{}{}
			if p := mapGetString(tr, "path"); p != "" {
				xh["path"] = p
			}
			if h := mapGetString(tr, "host"); h != "" {
				xh["host"] = h
			}
			if m := mapGetString(tr, "mode"); m != "" {
				xh["mode"] = m
			}
			if e := mapGetString(tr, "extra"); e != "" {
				xh["extra"] = e
			}
			if len(xh) > 0 {
				streamSettings["xhttpSettings"] = xh
			}
		default:
			return nil, fmt.Errorf("unsupported transport type %q for xray sidecar", transportType)
		}
	}

	// TLS / Reality
	if tls, ok := sb["tls"].(map[string]interface{}); ok {
		if mapGetBool(tls, "enabled") {
			if reality, ok := tls["reality"].(map[string]interface{}); ok && mapGetBool(reality, "enabled") {
				streamSettings["security"] = "reality"
				rs := map[string]interface{}{}
				if sni := mapGetString(tls, "server_name"); sni != "" {
					rs["serverName"] = sni
				}
				if fp := mapGetString(reality, "fingerprint"); fp != "" {
					rs["fingerprint"] = fp
				}
				if pk := mapGetString(reality, "public_key"); pk != "" {
					rs["publicKey"] = pk
				}
				if sid := mapGetString(reality, "short_id"); sid != "" {
					rs["shortId"] = sid
				}
				if spiderX := mapGetString(reality, "spiderx"); spiderX != "" {
					rs["spiderX"] = spiderX
				}
				if pbk := mapGetString(reality, "pbk"); pbk != "" {
					rs["publicKey"] = pbk
				}
				streamSettings["realitySettings"] = rs
			} else {
				streamSettings["security"] = "tls"
				ts := map[string]interface{}{}
				if sni := mapGetString(tls, "server_name"); sni != "" {
					ts["serverName"] = sni
				}
				if mapGetBool(tls, "insecure") {
					ts["allowInsecure"] = true
				}
				if alpn := mapGetString(tls, "alpn"); alpn != "" {
					ts["alpn"] = []string{alpn}
				}
				streamSettings["tlsSettings"] = ts
			}
		}
	}

	out := map[string]interface{}{
		"tag":      "proxy",
		"protocol": "vless",
		"settings": vlessSettings,
	}
	if len(streamSettings) > 0 {
		out["streamSettings"] = streamSettings
	}
	return out, nil
}

// mapGetString safely extracts a string value from a map.
func mapGetString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		switch s := v.(type) {
		case string:
			return s
		case []byte:
			return string(s)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// mapGetInt safely extracts an int value from a map.
func mapGetInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		case float32:
			return int(n)
		case string:
			var i int
			fmt.Sscanf(n, "%d", &i)
			return i
		}
	}
	return 0
}

// mapGetBool safely extracts a bool value from a map.
func mapGetBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		switch b := v.(type) {
		case bool:
			return b
		case string:
			return b == "true"
		case int, int64, float64:
			return fmt.Sprintf("%v", v) != "0"
		}
	}
	return false
}
