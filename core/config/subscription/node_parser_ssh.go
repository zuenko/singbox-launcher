package subscription

import (
	"net/url"
	"strings"

	"singbox-launcher/core/config/configtypes"
	"singbox-launcher/internal/debuglog"
)

// buildSSHOutbound builds outbound configuration for SSH protocol
func buildSSHOutbound(node *configtypes.ParsedNode, outbound map[string]interface{}) {
	// User is required (stored in UUID field from userinfo)
	if node.UUID != "" {
		outbound["user"] = node.UUID
	} else {
		outbound["user"] = "root" // Default user for SSH
		debuglog.WarnLog("Parser: SSH link missing user, using default 'root'")
	}

	// Password is optional (can be in query params from userinfo)
	if password := node.Query.Get("password"); password != "" {
		outbound["password"] = password
	}

	// Private key (inline) - if provided, takes precedence over private_key_path
	if privateKey := node.Query.Get("private_key"); privateKey != "" {
		if decoded, err := url.QueryUnescape(privateKey); err == nil {
			outbound["private_key"] = decoded
		} else {
			outbound["private_key"] = privateKey
		}
	} else if privateKeyPath := node.Query.Get("private_key_path"); privateKeyPath != "" {
		if decoded, err := url.QueryUnescape(privateKeyPath); err == nil {
			outbound["private_key_path"] = decoded
		} else {
			outbound["private_key_path"] = privateKeyPath
		}
	}

	// Private key passphrase
	if passphrase := node.Query.Get("private_key_passphrase"); passphrase != "" {
		if decoded, err := url.QueryUnescape(passphrase); err == nil {
			outbound["private_key_passphrase"] = decoded
		} else {
			outbound["private_key_passphrase"] = passphrase
		}
	}

	// Host key (can be multiple, comma-separated)
	if hostKey := node.Query.Get("host_key"); hostKey != "" {
		hostKeys := strings.Split(hostKey, ",")
		decodedKeys := make([]string, 0, len(hostKeys))
		for _, key := range hostKeys {
			key = strings.TrimSpace(key)
			if key != "" {
				if decoded, err := url.QueryUnescape(key); err == nil {
					decodedKeys = append(decodedKeys, decoded)
				} else {
					decodedKeys = append(decodedKeys, key)
				}
			}
		}
		if len(decodedKeys) > 0 {
			outbound["host_key"] = decodedKeys
		}
	}

	// Host key algorithms (can be multiple, comma-separated)
	if algorithms := node.Query.Get("host_key_algorithms"); algorithms != "" {
		algList := strings.Split(algorithms, ",")
		for i := range algList {
			algList[i] = strings.TrimSpace(algList[i])
		}
		filteredAlgs := make([]string, 0, len(algList))
		for _, alg := range algList {
			if alg != "" {
				filteredAlgs = append(filteredAlgs, alg)
			}
		}
		if len(filteredAlgs) > 0 {
			outbound["host_key_algorithms"] = filteredAlgs
		}
	}

	// Client version
	if clientVersion := node.Query.Get("client_version"); clientVersion != "" {
		if decoded, err := url.QueryUnescape(clientVersion); err == nil {
			outbound["client_version"] = decoded
		} else {
			outbound["client_version"] = clientVersion
		}
	}
}
