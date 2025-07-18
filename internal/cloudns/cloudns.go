/*
Copyright 2022 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudns

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"external-dns-cloudns-webhook/internal/metrics"

	cloudns "github.com/ppmathis/cloudns-go"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

const (
	actGetZones     = "get_zones"
	actGetRecords   = "get_records"
	actCreateRecord = "create_record"
	actUpdateRecord = "update_record"
	actDeleteRecord = "delete_record"
)

// ClouDNSProvider is a struct representing a CloudDNS provider.
// It embeds the provider.BaseProvider struct and includes fields for the CloudDNS client, context, domain and zone ID filters, owner ID, and flags for dry-run and testing modes.
type ClouDNSProvider struct {
	provider.BaseProvider
	client       *cloudns.Client
	domainFilter *endpoint.DomainFilter
	defaultTTL   int
	ownerID      string
	debug        bool
	dryRun       bool
	testing      bool
}

// ClouDNSConfig is a struct representing the configuration for a CloudDNS provider.
// It includes fields for the context, domain and zone ID filters, owner ID, and flags for dry-run and testing modes.
type ClouDNSConfig struct {
	Auth         cloudns.Option
	DomainFilter *endpoint.DomainFilter
	ZoneIDFilter provider.ZoneIDFilter
	DefaultTTL   int
	OwnerID      string
	Debug        bool
	DryRun       bool
	Testing      bool
}

var listZones = func(client *cloudns.Client, ctx context.Context) ([]cloudns.Zone, error) {
	metrics := metrics.GetOpenMetricsInstance()
	start := time.Now()

	result, err := client.Zones.List(ctx)
	if err != nil {
		metrics.IncFailedApiCallsTotal(actGetZones)
		return nil, err
	}

	delay := time.Since(start)
	metrics.IncSuccessfulApiCallsTotal(actGetZones)
	metrics.AddApiDelayHist(actGetZones, delay.Milliseconds())

	return result, nil
}

var listRecords = func(client *cloudns.Client, ctx context.Context, zoneName string) (cloudns.RecordMap, error) {
	metrics := metrics.GetOpenMetricsInstance()
	start := time.Now()

	result, err := client.Records.List(ctx, zoneName)
	if err != nil {
		metrics.IncFailedApiCallsTotal(actGetRecords)
		return nil, err
	}

	delay := time.Since(start)
	metrics.IncSuccessfulApiCallsTotal(actGetRecords)
	metrics.AddApiDelayHist(actGetRecords, delay.Milliseconds())

	return result, nil
}

var createRecord = func(client *cloudns.Client, ctx context.Context, zoneName string, record cloudns.Record) error {
	metrics := metrics.GetOpenMetricsInstance()
	start := time.Now()

	_, err := client.Records.Create(ctx, zoneName, record)
	if err != nil {
		metrics.IncFailedApiCallsTotal(actCreateRecord)
		return err
	}

	delay := time.Since(start)
	metrics.IncSuccessfulApiCallsTotal(actCreateRecord)
	metrics.AddApiDelayHist(actCreateRecord, delay.Milliseconds())

	return nil
}

var deleteRecord = func(client *cloudns.Client, ctx context.Context, zoneName string, recordID int) error {
	metrics := metrics.GetOpenMetricsInstance()
	start := time.Now()

	_, err := client.Records.Delete(ctx, zoneName, recordID)
	if err != nil {
		metrics.IncFailedApiCallsTotal(actDeleteRecord)
		return err
	}

	delay := time.Since(start)
	metrics.IncSuccessfulApiCallsTotal(actDeleteRecord)
	metrics.AddApiDelayHist(actDeleteRecord, delay.Milliseconds())

	return nil
}

// NewClouDNSProvider creates and returns a new ClouDNSProvider struct based on the given configuration.
// The function authenticates with the CloudDNS service using the login type, user or sub-user ID, and user password specified in the environment variables.
// If an error occurs while authenticating or creating the ClouDNS client, it is returned.
func NewClouDNSProvider(config ClouDNSConfig) (*ClouDNSProvider, error) {
	var logLevel log.Level
	if config.Debug {
		logLevel = log.DebugLevel
	} else {
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	log.Info("Creating ClouDNS Provider")

	client, error := cloudns.New(config.Auth)
	if error != nil {
		return nil, fmt.Errorf("error creating ClouDNS client: %s", error)
	}

	provider := &ClouDNSProvider{
		client:       client,
		domainFilter: config.DomainFilter,
		defaultTTL:   config.DefaultTTL,
		ownerID:      config.OwnerID,
		debug:        config.Debug,
		dryRun:       config.DryRun,
		testing:      config.Testing,
	}

	return provider, nil
}

// Zones retrieves the DNS zone from the ClouDNS provider,
// applies the defined domainFilter and returns the result
func (p *ClouDNSProvider) Zones(ctx context.Context) ([]cloudns.Zone, error) {
	metrics := metrics.GetOpenMetricsInstance()
	result := []cloudns.Zone{}

	zones, err := listZones(p.client, ctx)
	if err != nil {
		return nil, err
	}

	filteredOutZones := 0
	for _, zone := range zones {
		if p.domainFilter.Match(zone.Name) {
			result = append(result, zone)
		} else {
			filteredOutZones++
		}
	}
	metrics.SetFilteredOutZones(filteredOutZones)

	return result, nil
}

// Records retrieves the DNS records from the CloudDNS provider and returns them as a slice of endpoint.Endpoint structs.
// The function retrieves all zones and their corresponding records and filters out unsupported record types.
// If an error occurs while retrieving the zones or records, it is returned.
func (p *ClouDNSProvider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	log.Info("Getting Records from ClouDNS")

	var endpoints []*endpoint.Endpoint

	zones, err := p.Zones(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting zones: %s", err)
	}

	for _, zone := range zones {

		records, err := listRecords(p.client, ctx, zone.Name)
		if err != nil {
			return nil, fmt.Errorf("error getting records: %s", err)
		}

		skippedRecords := 0
		// Add only endpoints from supported types.
		for _, record := range records {
			if provider.SupportedRecordType(string(record.RecordType)) {
				name := ""

				if record.Host == "" || record.Host == "@" {
					name = zone.Name
				} else {
					name = record.Host + "." + zone.Name
				}

				if record.RecordType == cloudns.RecordTypeTXT {
					if record.Host == "adash" {
						name = "a-" + zone.Name
					}
				}

				endpoints = append(endpoints, endpoint.NewEndpointWithTTL(
					name,
					string(record.RecordType),
					endpoint.TTL(record.TTL),
					record.Record,
				))
			} else {
				skippedRecords++
			}
		}
		m := metrics.GetOpenMetricsInstance()
		m.SetSkippedRecords(zone.Name, skippedRecords)
	}

	merged := mergeEndpointsByNameType(endpoints)

	out := "Found:"
	for _, e := range merged {
		if e.RecordType != endpoint.RecordTypeTXT {
			out = out + " [" + e.DNSName + " " + e.RecordType + " " + e.Targets[0] + " " + fmt.Sprint(e.RecordTTL) + "]"
		}
	}
	log.Debugf("%s", out)

	return merged, nil
}

// ApplyChanges applies the given DNS changes to the CloudDNS provider.
// The function retrieves the zones, creates new records, deletes old records, and updates existing records as needed.
// If the provider is in dry-run mode, the changes are not applied but the details of the changes are logged.
// If an error occurs while retrieving the zones or applying the changes, it is returned.
func (p *ClouDNSProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	infoString := "Creating " + fmt.Sprint(len(changes.Create)) + " Record(s), Updating " + fmt.Sprint(len(changes.UpdateNew)) + " Record(s), Deleting " + fmt.Sprint(len(changes.Delete)) + " Record(s)"

	if len(changes.Create) == 0 && len(changes.Delete) == 0 && len(changes.UpdateNew) == 0 && len(changes.UpdateOld) == 0 {
		log.Info("No Changes")
		return nil
	} else if p.dryRun {
		log.Info("DRY RUN: " + infoString)
	} else {
		log.Info(infoString)
	}

	err := p.createRecords(ctx, changes.Create)
	if err != nil {
		return err
	}

	err = p.deleteRecords(ctx, changes.Delete)
	if err != nil {
		return err
	}

	err = p.updateRecords(ctx, changes.UpdateOld, changes.UpdateNew)
	if err != nil {
		return err
	}

	return nil
}

// createRecords creates DNS records in the CloudDNS provider for the given endpoints.
// The function takes in a context and a slice of endpoint.Endpoint structs.
// If an error occurs while creating the records, it is returned.
func (p *ClouDNSProvider) createRecords(ctx context.Context, endpoints []*endpoint.Endpoint) error {
	for _, ep := range endpoints {

		dnsParts := strings.Split(ep.DNSName, ".")
		partLength := len(dnsParts)
		rootZone := rootZone(ep.DNSName)
		log.Debugf("Analyzed %s: len=%d, rootZone=%s", ep.DNSName, partLength, rootZone)

		zones, err := p.Zones(ctx)
		if err != nil {
			return err
		}

		idx := slices.IndexFunc(zones, func(z cloudns.Zone) bool {
			return z.Name == rootZone
		})
		if idx < 0 {
			log.Warnf("Skipping %s as %s is not one of our zones", ep.DNSName, rootZone)
			continue
		}

		if ep.RecordType == "TXT" {
			if !p.dryRun {
				if partLength == 2 && dnsParts[0][0:2] == "a-" {
					err := createRecord(p.client, ctx, rootZone[2:], cloudns.Record{
						Host:       "adash",
						Record:     ep.Targets[0],
						RecordType: cloudns.RecordType("TXT"),
						TTL:        60,
					})
					if err != nil {
						return err
					}
				} else {
					hostName := removeLastOccurrance(ep.DNSName, "."+rootZone)
					if hostName == rootZone {
						hostName = ""
					}

					err := createRecord(p.client, ctx, rootZone, cloudns.Record{
						Host:       hostName,
						Record:     ep.Targets[0],
						RecordType: cloudns.RecordType("TXT"),
						TTL:        60,
					})
					if err != nil {
						return err
					}
				}
				log.Infof("CREATE %s %s %s %s", ep.DNSName, ep.RecordType, ep.Targets[0], fmt.Sprint(ep.RecordTTL))
			} else {
				log.Infof("DRY RUN: CREATE %s %s %s %s", ep.DNSName, ep.RecordType, ep.Targets[0], fmt.Sprint(ep.RecordTTL))
			}
		}

		if ep.RecordTTL == endpoint.TTL(0) {
			ep.RecordTTL = endpoint.TTL(p.defaultTTL)
		}

		if !isValidTTL(strconv.Itoa(int(ep.RecordTTL))) && !(ep.RecordType == "TXT") { //nolint:staticcheck
			return fmt.Errorf("invalid TTL %s (still) for %s - must be one of '60', '300', '900', '1800', '3600', '21600', '43200', '86400', '172800', '259200', '604800', '1209600', '2592000'", fmt.Sprint(ep.RecordTTL), ep.DNSName)
		}

		if partLength == 2 && !(ep.RecordType == "TXT") { //nolint:staticcheck
			for _, target := range ep.Targets {
				if !p.dryRun {
					err := createRecord(p.client, ctx, ep.DNSName, cloudns.Record{
						Host:       "",
						Record:     target,
						RecordType: cloudns.RecordType(ep.RecordType),
						TTL:        int(ep.RecordTTL),
					})
					if err != nil {
						return err
					}

					log.Infof("CREATE %s %s %s %s", ep.DNSName, ep.RecordType, target, fmt.Sprint(ep.RecordTTL))
				} else {
					log.Infof("DRY RUN: CREATE %s %s %s %s", ep.DNSName, ep.RecordType, target, fmt.Sprint(ep.RecordTTL))
				}
			}
		} else if partLength > 2 && !(ep.RecordType == "TXT") { //nolint:staticcheck

			hostName := removeLastOccurrance(ep.DNSName, "."+rootZone)

			for _, target := range ep.Targets {
				if !p.dryRun {
					err := createRecord(p.client, ctx, rootZone, cloudns.Record{
						Host:       hostName,
						Record:     target,
						RecordType: cloudns.RecordType(ep.RecordType),
						TTL:        int(ep.RecordTTL),
					})
					if err != nil {
						return err
					}

					log.Infof("CREATE %s %s %s %s", ep.DNSName, ep.RecordType, target, fmt.Sprint(ep.RecordTTL))
				} else {
					log.Infof("DRY RUN: CREATE %s %s %s %s", ep.DNSName, ep.RecordType, target, fmt.Sprint(ep.RecordTTL))
				}
			}
		}
	}

	return nil
}

// deleteRecords deletes DNS records from the CloudDNS provider for the given endpoints.
// The function takes in a context and a slice of endpoint.Endpoint structs.
// If an error occurs while deleting the records, it is returned.
func (p *ClouDNSProvider) deleteRecords(ctx context.Context, endpoints []*endpoint.Endpoint) error {
	for _, ep := range endpoints {
		rootZone := rootZone(ep.DNSName)
		hostName := ""

		if rootZone[0:2] == "a-" && ep.RecordType == "TXT" {
			rootZone = rootZone[2:]
			hostName = "adash"
		} else {
			hostName = removeRootZone(ep.DNSName, rootZone)
		}

		zones, err := p.Zones(ctx)
		if err != nil {
			return err
		}

		idx := slices.IndexFunc(zones, func(z cloudns.Zone) bool {
			return z.Name == rootZone
		})
		if idx < 0 {
			log.Warnf("Skipping %s as %s is not one of our zones", ep.DNSName, rootZone)
			continue
		}

		for _, target := range ep.Targets {

			id, zone, err := p.recordFromTarget(ctx, ep, target, rootZone, hostName)
			if err != nil {
				return err
			}

			if id == 0 {
				log.Infof("Record not found: %s %s %s", ep.DNSName, ep.RecordType, target)
				continue
			} else if !p.dryRun {
				err := deleteRecord(p.client, ctx, zone, id)
				if err != nil {
					return err
				}
				log.Infof("DELETE %s %s %s %s", ep.DNSName, ep.RecordType, target, fmt.Sprint(ep.RecordTTL))
			} else {
				log.Infof("DRY RUN: DELETE %s %s %s %s", ep.DNSName, ep.RecordType, target, fmt.Sprint(ep.RecordTTL))
			}

		}
	}

	return nil
}

// updateRecords updates the records in the ClouDNS provider by first creating the records in the updateNew slice
// and then deleting the records in the updateOld slice. If an error occurs while creating or deleting the records,
// it is returned.
//
// The updateNew slice should contain the updated records that need to be created, and the updateOld slice should
// contain the old records that need to be deleted.
func (p *ClouDNSProvider) updateRecords(ctx context.Context, updateOld, updateNew []*endpoint.Endpoint) error {
	err := p.createRecords(ctx, updateNew)
	if err != nil {
		return err
	}

	err = p.deleteRecords(ctx, updateOld)
	if err != nil {
		return err
	}

	return nil
}

// recordFromTarget returns the ID and zone name of a record in the ClouDNS provider
// that matches the given endpoint, target, and zone name. If no matching record is found,
// the ID is returned as 0 and the zone name is returned as an empty string.
//
// The function first retrieves a map of zones and their corresponding records from the ClouDNS provider.
// It then iterates over the records, checking if the record type, hostname, zone name, and target
// match the given endpoint and target. If a match is found, the ID and zone name of the record
// are returned. If no match is found, the ID is returned as 0 and the zone name is returned as an empty string.
// If an error occurs while retrieving the map of zones and records, it is returned.
func (p *ClouDNSProvider) recordFromTarget(ctx context.Context, ep *endpoint.Endpoint, target string, epZoneName string, epHostName string) (int, string, error) {
	zoneRecordMap, err := p.zoneRecordMap(ctx)
	if err != nil {
		return 0, "", err
	}

	for zoneName, recordMap := range zoneRecordMap {
		for _, record := range recordMap {
			if string(record.RecordType) == "TXT" {
				if record.Host == "adash" {
					record.Host = "a-"
				}

				if epHostName == "adash" {
					epHostName = "a-"
				}

				if string(record.RecordType) == ep.RecordType && record.Host == epHostName && zoneName == epZoneName && record.Record == strings.Trim(target, "\\\"") {
					return record.ID, zoneName, nil
				}
			} else if string(record.RecordType) == ep.RecordType && record.Host == epHostName && zoneName == epZoneName && record.Record == target {
				return record.ID, zoneName, nil
			}
		}
	}

	return 0, "", nil
}

// zoneRecordMap returns a map of all zones and their corresponding records in the CloudDNS provider.
// The map keys are the zone names and the map values are slices of cloudns.Record structs representing the records in the zone.
// If an error occurs while retrieving the zones or records, it is returned.
func (p *ClouDNSProvider) zoneRecordMap(ctx context.Context) (map[string]cloudns.RecordMap, error) {
	zoneRecordMap := make(map[string]cloudns.RecordMap)

	zones, err := p.Zones(ctx)
	if err != nil {
		return nil, err
	}

	for _, zone := range zones {
		recordMap, err := listRecords(p.client, ctx, zone.Name)
		if err != nil {
			return nil, err
		}

		if len(recordMap) != 0 {
			zoneRecordMap[zone.Name] = make(cloudns.RecordMap)
			zoneRecordMap[zone.Name] = recordMap
		}
	}

	return zoneRecordMap, nil
}
