package cloudns

import (
	"fmt"
	"sort"
	"strings"

	cloudns "github.com/ppmathis/cloudns-go"
	"sigs.k8s.io/external-dns/endpoint"
)

// mergeEndpointsByNameType takes a slice of endpoints and returns a new slice of endpoints
// with the endpoints merged based on their DNS name and record type. If no merge occurs,
// the original slice of endpoints is returned.
// From pkg/digitalocean/provider.go
func mergeEndpointsByNameType(endpoints []*endpoint.Endpoint) []*endpoint.Endpoint {
	endpointsByNameType := map[string][]*endpoint.Endpoint{}
	keys := []string{}

	for _, e := range endpoints {
		key := fmt.Sprintf("%s-%s", e.DNSName, e.RecordType)

		if _, ok := endpointsByNameType[key]; !ok {
			keys = append(keys, key)
		}
		endpointsByNameType[key] = append(endpointsByNameType[key], e)
	}

	// If no merge occurred, just return the existing endpoints.
	if len(endpointsByNameType) == len(endpoints) {
		return endpoints
	}

	// Otherwise, construct a new list of endpoints with the endpoints merged.
	var result []*endpoint.Endpoint
	for _, key := range keys {
		endpoints := endpointsByNameType[key]
		dnsName := endpoints[0].DNSName
		recordType := endpoints[0].RecordType
		ttl := endpoints[0].RecordTTL

		targets := make([]string, len(endpoints))
		for i, ep := range endpoints {
			targets[i] = ep.Targets[0]
		}

		e := endpoint.NewEndpoint(dnsName, recordType, targets...)
		e.RecordTTL = ttl
		result = append(result, e)
	}

	return result
}

// isValidTTL checks if the given time-to-live (TTL) value is valid.
// A valid TTL value is a string representation of a positive integer that is one of the following values:
// "60", "300", "900", "1800", "3600", "21600", "43200", "86400", "172800", "259200", "604800", "1209600", "2592000".
// The function returns true if the given TTL value is valid and false otherwise.
func isValidTTL(ttl string) bool {
	validTTLs := []string{"60", "300", "900", "1800", "3600", "21600", "43200", "86400", "172800", "259200", "604800", "1209600", "2592000"}

	for _, validTTL := range validTTLs {
		if ttl == validTTL {
			return true
		}
	}

	return false
}

// rootZone returns the root zone of a domain name.
// A root zone is the last two parts of a domain name, separated by a "." character.
// For example, the root zone of "test.this.program.com" is "program.com" and
// the root zone of "easy.com" is "easy.com".
// If the domain name has less than two parts, the domain name is returned as-is.
//
// Deprecated: Use findZoneForDomain instead, which properly handles subdomain zones.
func rootZone(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return domain
	}
	return strings.Join(parts[len(parts)-2:], ".")
}

// findZoneForDomain finds the longest matching zone suffix from the list of available zones.
// This properly handles subdomain zones (e.g., k8s.example.com) by finding the most specific
// zone that matches the domain.
//
// For example, given zones ["example.com", "k8s.example.com"] and domain "dashboard.k8s.example.com",
// this function returns "k8s.example.com" (the longest matching zone), not "example.com".
//
// It also handles ExternalDNS registry TXT record prefixes (e.g., "reg-a-", "reg-aaaa-", "reg-cname-")
// where the prefix is prepended to the zone apex. For example, "reg-a-example.com" matches zone "example.com".
//
// Returns an empty string if no matching zone is found.
func findZoneForDomain(domain string, zones []cloudns.Zone) string {
	// Build list of zone names sorted by length (longest first)
	zoneNames := make([]string, len(zones))
	for i, z := range zones {
		zoneNames[i] = z.Name
	}
	sort.Slice(zoneNames, func(i, j int) bool {
		return len(zoneNames[i]) > len(zoneNames[j])
	})

	// Find the longest matching zone suffix
	for _, zoneName := range zoneNames {
		if domain == zoneName {
			return zoneName
		}
		if strings.HasSuffix(domain, "."+zoneName) {
			return zoneName
		}
		// Handle ExternalDNS registry TXT record prefixes (e.g., "reg-a-example.com" for zone "example.com")
		// These prefixes are used for heritage tracking and follow the pattern: prefix + zone
		if strings.HasSuffix(domain, "-"+zoneName) {
			return zoneName
		}
	}
	return ""
}

// Returns the domain name with the root zone and any trailing periods removed.
// domain is the domain name to be modified.
// rootZone is the root zone to be removed from the domain name.
func removeRootZone(domain string, rootZone string) string {
	if strings.LastIndex(domain, rootZone) == -1 {
		return domain
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return domain
	}
	rootZoneIndex := len(parts) - len(strings.Split(rootZone, "."))
	return strings.TrimSuffix(strings.Join(parts[:rootZoneIndex], "."), ".")
}

// removeLastOccurrence removes the last occurrence of the given substring from the given string.
// If the substring is not present, the original string is returned.
func removeLastOccurrance(str, subStr string) string {
	i := strings.LastIndex(str, subStr)

	if i == -1 {
		return str
	}

	return strings.Join([]string{str[:i], str[i+len(subStr):]}, "")
}
