# ExternalDNS ClouDNS Webhook - Subdomain Zone Bug

## Repository

- **Project:** https://github.com/rwunderer/external-dns-cloudns-webhook
- **Latest Version:** v0.3.34 (as of Nov 2025)
- **Status:** Actively maintained

---

## Issue Summary

The webhook does not support **subdomain zones** (e.g., `k8s.example.com`). It only works correctly with standard TLD zones (e.g., `example.com`, `myapp.io`).

### Root Cause

The `rootZone()` function in `internal/cloudns/helpers.go` hardcodes zone detection to the **last 2 parts** of a domain name:

```go
// Current buggy implementation (line ~45 in helpers.go)
func rootZone(domain string) string {
    parts := strings.Split(domain, ".")
    if len(parts) < 2 {
        return domain
    }
    return strings.Join(parts[len(parts)-2:], ".")
}
```

### Example Failures

| Domain to create | Calculated `rootZone` | Actual ClouDNS Zone | Result |
|------------------|----------------------|---------------------|--------|
| `dashboard.k8s.glide.sk` | `glide.sk` | `k8s.glide.sk` | ❌ SKIPPED |
| `whoami.i.glide.sk` | `glide.sk` | `i.glide.sk` | ❌ SKIPPED |
| `app.staging.example.com` | `example.com` | `staging.example.com` | ❌ SKIPPED |
| `dashboard.example.com` | `example.com` | `example.com` | ✅ Works |
| `app.vagrantfile.app` | `vagrantfile.app` | `vagrantfile.app` | ✅ Works |

### Log Output

```
level=debug msg="Found: [k8s.glide.sk NS pns54.cloudns.net 3600] [vgf.k8s.glide.sk A 188.245.80.233 3600]"
level=info msg="Creating 2 Record(s), Updating 0 Record(s), Deleting 0 Record(s)"
level=debug msg="Analyzed dashboard.k8s.glide.sk: len=4, rootZone=glide.sk"
level=warning msg="Skipping dashboard.k8s.glide.sk as glide.sk is not one of our zones"
```

The webhook **correctly finds** the zone `k8s.glide.sk` from the ClouDNS API, but then **incorrectly calculates** `rootZone=glide.sk` and fails to match.

---

## Proposed Solution

Replace the naive `rootZone()` function with a proper zone matching algorithm that finds the **longest matching zone suffix** from the list of available zones.

### Option A: Pass zones to rootZone (minimal change)

Modify `rootZone()` to accept available zones and find the best match:

```go
// New implementation - finds longest matching zone suffix
func findZoneForDomain(domain string, zones []cloudns.Zone) string {
    // Build sorted list of zone names (longest first)
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
    }
    return ""
}
```

### Option B: Cache zones at provider level (cleaner)

Store zones in the `ClouDNSProvider` struct and use them for lookups:

```go
type ClouDNSProvider struct {
    // ... existing fields ...
    cachedZones []cloudns.Zone
    zoneNames   []string  // sorted by length (longest first)
}

func (p *ClouDNSProvider) refreshZoneCache(ctx context.Context) error {
    zones, err := p.Zones(ctx)
    if err != nil {
        return err
    }
    p.cachedZones = zones
    
    // Sort zone names by length (longest first) for proper matching
    p.zoneNames = make([]string, len(zones))
    for i, z := range zones {
        p.zoneNames[i] = z.Name
    }
    sort.Slice(p.zoneNames, func(i, j int) bool {
        return len(p.zoneNames[i]) > len(p.zoneNames[j])
    })
    return nil
}

func (p *ClouDNSProvider) findZone(domain string) string {
    for _, zoneName := range p.zoneNames {
        if domain == zoneName || strings.HasSuffix(domain, "."+zoneName) {
            return zoneName
        }
    }
    return ""
}
```

---

## Files to Modify

1. **`internal/cloudns/helpers.go`**
   - Remove or deprecate `rootZone()` function
   - Add new `findZoneForDomain()` function (if using Option A)

2. **`internal/cloudns/cloudns.go`**
   - Modify `createRecords()` to use new zone matching
   - Modify `deleteRecords()` to use new zone matching
   - (Option B) Add zone caching to provider struct

### Current Usage in cloudns.go

The buggy `rootZone()` is called in these places:

```go
// In createRecords() - line ~180
rootZone := rootZone(ep.DNSName)
log.Debugf("Analyzed %s: len=%d, rootZone=%s", ep.DNSName, partLength, rootZone)

zones, err := p.Zones(ctx)
// ... then tries to find rootZone in zones list ...
idx := slices.IndexFunc(zones, func(z cloudns.Zone) bool {
    return z.Name == rootZone  // This fails for subdomain zones!
})

// In deleteRecords() - line ~250
rootZone := rootZone(ep.DNSName)
// ... same pattern ...
```

### Fix Pattern

Replace:
```go
rootZone := rootZone(ep.DNSName)
zones, err := p.Zones(ctx)
idx := slices.IndexFunc(zones, func(z cloudns.Zone) bool {
    return z.Name == rootZone
})
if idx < 0 {
    log.Warnf("Skipping %s as %s is not one of our zones", ep.DNSName, rootZone)
    continue
}
```

With:
```go
zones, err := p.Zones(ctx)
if err != nil {
    return err
}
matchedZone := findZoneForDomain(ep.DNSName, zones)
if matchedZone == "" {
    log.Warnf("Skipping %s - no matching zone found", ep.DNSName)
    continue
}
log.Debugf("Matched %s to zone %s", ep.DNSName, matchedZone)
```

---

## Testing

### Test Cases to Add

```go
func TestFindZoneForDomain(t *testing.T) {
    zones := []cloudns.Zone{
        {Name: "example.com"},
        {Name: "k8s.example.com"},
        {Name: "staging.k8s.example.com"},
        {Name: "glide.sk"},
        {Name: "k8s.glide.sk"},
        {Name: "i.glide.sk"},
        {Name: "vagrantfile.app"},
    }
    
    tests := []struct {
        domain   string
        expected string
    }{
        // Standard TLD zones
        {"dashboard.example.com", "example.com"},
        {"api.example.com", "example.com"},
        {"example.com", "example.com"},
        
        // Subdomain zones - should match longest suffix
        {"dashboard.k8s.example.com", "k8s.example.com"},
        {"app.k8s.example.com", "k8s.example.com"},
        {"k8s.example.com", "k8s.example.com"},
        
        // Deeper subdomain zones
        {"myapp.staging.k8s.example.com", "staging.k8s.example.com"},
        
        // Real-world cases from issue
        {"dashboard.k8s.glide.sk", "k8s.glide.sk"},
        {"vgf.k8s.glide.sk", "k8s.glide.sk"},
        {"whoami.i.glide.sk", "i.glide.sk"},
        {"traefik.i.glide.sk", "i.glide.sk"},
        {"app.vagrantfile.app", "vagrantfile.app"},
        
        // No match
        {"dashboard.unknown.com", ""},
        {"random.domain.org", ""},
    }
    
    for _, tt := range tests {
        t.Run(tt.domain, func(t *testing.T) {
            result := findZoneForDomain(tt.domain, zones)
            if result != tt.expected {
                t.Errorf("findZoneForDomain(%q) = %q, want %q", 
                    tt.domain, result, tt.expected)
            }
        })
    }
}
```

---

## GitHub Issue Template

### Title
`Bug: Subdomain zones not supported (e.g., k8s.example.com)`

### Body

```markdown
## Description

The webhook fails to create DNS records when using subdomain zones in ClouDNS 
(e.g., `k8s.example.com` instead of `example.com`).

## Environment

- Webhook version: v0.3.34
- ExternalDNS version: v0.15.0
- Kubernetes: k3s v1.31

## ClouDNS Configuration

I have three subdomain zones in ClouDNS:
- `i.glide.sk`
- `k8s.glide.sk`  
- `vagrantfile.app`

Note: The parent zone `glide.sk` is NOT managed by ClouDNS.

## Steps to Reproduce

1. Configure ExternalDNS with ClouDNS webhook for zone `k8s.glide.sk`
2. Create HTTPRoute with hostname `dashboard.k8s.glide.sk`
3. ExternalDNS detects the route but fails to create DNS record

## Expected Behavior

A record `dashboard` should be created in zone `k8s.glide.sk`.

## Actual Behavior

Record is skipped with warning:
```
level=warning msg="Skipping dashboard.k8s.glide.sk as glide.sk is not one of our zones"
```

## Root Cause

The `rootZone()` function in `helpers.go` assumes zones are always 2-part TLDs:

```go
func rootZone(domain string) string {
    parts := strings.Split(domain, ".")
    return strings.Join(parts[len(parts)-2:], ".")
}
```

For `dashboard.k8s.glide.sk`, this returns `glide.sk` instead of matching 
against available zones to find `k8s.glide.sk`.

## Proposed Fix

Replace naive string splitting with proper zone suffix matching that finds 
the longest matching zone from the available zones list.

I'm happy to submit a PR with the fix if you agree with the approach.
```

---

## Notes

- The **cert-manager ClouDNS webhook** (`mschirrmeister/cert-manager-webhook-cloudns`) does NOT have this bug - it correctly handles subdomain zones
- This is why certificate issuance (DNS-01 challenge) works, but ExternalDNS A record creation fails
- The fix is straightforward and backwards-compatible - it will continue to work for standard TLD zones
