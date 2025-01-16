/*
 * Change processors - this file contains the code for processing changes and
 * queue them.
 *
 * Copyright 2023 Marco Confalonieri.
 * Copyright 2017 The Kubernetes Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package cloudns

import (
	"strings"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"

	cdns "github.com/ppmathis/cloudns-go"
	log "github.com/sirupsen/logrus"
)

// adjustCNAMETarget fixes local CNAME targets. It ensures that targets
// matching the domain are stripped of the domain parts and that "external"
// targets end with a dot.
func adjustCNAMETarget(domain string, target string) string {
	adjustedTarget := target
	if strings.HasSuffix(target, "."+domain) {
		adjustedTarget = strings.TrimSuffix(target, "."+domain)
	} else if strings.HasSuffix(target, "."+domain+".") {
		adjustedTarget = strings.TrimSuffix(target, "."+domain+".")
	} else if !strings.HasSuffix(target, ".") {
		adjustedTarget += "."
	}
	return adjustedTarget
}

// processCreateActionsByZone processes the create actions for one zone.
func processCreateActionsByZone(zoneID, zoneName string, records []cdns.Record, endpoints []*endpoint.Endpoint, changes *cloudnsChanges) {
	for _, ep := range endpoints {
		// Warn if there are existing records since we expect to create only new records.
		matchingRecords := getMatchingDomainRecords(records, zoneName, ep)
		if len(matchingRecords) > 0 {
			log.WithFields(log.Fields{
				"zoneName":   zoneName,
				"dnsName":    ep.DNSName,
				"recordType": ep.RecordType,
			}).Warn("Preexisting records exist which should not exist for creation actions.")
		}

		for _, target := range ep.Targets {
			if ep.RecordType == "CNAME" {
				target = adjustCNAMETarget(zoneName, target)
			}
            rec := cdns.NewRecord(
                cdns.RecordType(ep.RecordType),
				makeEndpointName(zoneName, ep.DNSName),
                target,
                getEndpointTTL(ep),
            )
			changes.AddChangeCreate(zoneID, rec)
		}
	}
}

// processCreateActions processes the create requests.
func processCreateActions(
	zoneIDNameMapper provider.ZoneIDName,
	recordsByZoneID map[string][]cdns.Record,
	createsByZoneID map[string][]*endpoint.Endpoint,
	changes *cloudnsChanges,
) {
	// Process endpoints that need to be created.
	for zoneID, endpoints := range createsByZoneID {
		zoneName := zoneIDNameMapper[zoneID]
		if len(endpoints) == 0 {
			log.WithFields(log.Fields{
				"zoneName": zoneName,
			}).Debug("Skipping domain, no creates found.")
			continue
		}
		records := recordsByZoneID[zoneID]
		processCreateActionsByZone(zoneID, zoneName, records, endpoints, changes)
	}
}

func processUpdateEndpoint(zoneID, zoneName string, matchingRecordsByTarget map[string]cdns.Record, ep *endpoint.Endpoint, changes *cloudnsChanges) {
	// Generate create and delete actions based on existence of a record for each target.
	for _, target := range ep.Targets {
		if ep.RecordType == "CNAME" {
			target = adjustCNAMETarget(zoneName, target)
		}
		if _, ok := matchingRecordsByTarget[target]; ok {
            rec := cdns.NewRecord(
                cdns.RecordType(ep.RecordType),
				makeEndpointName(zoneName, ep.DNSName),
                target,
                getEndpointTTL(ep),
            )
			changes.AddChangeUpdate(zoneID, rec)

			// Updates are removed from this map.
			delete(matchingRecordsByTarget, target)
		} else {
			// Record did not previously exist, create new 'target'
            rec := cdns.NewRecord(
                cdns.RecordType(ep.RecordType),
				makeEndpointName(zoneName, ep.DNSName),
                target,
                getEndpointTTL(ep),
            )
			changes.AddChangeCreate(zoneID, rec)
		}
	}
}

// cleanupRemainingTargets deletes the entries for the updates that are queued for creation.
func cleanupRemainingTargets(zoneID string, matchingRecordsByTarget map[string]cdns.Record, changes *cloudnsChanges) {
	for _, record := range matchingRecordsByTarget {
		changes.AddChangeDelete(zoneID, record)
	}
}

// getMatchingRecordsByTarget organizes a slice of targets in a map with the
// target as key.
func getMatchingRecordsByTarget(records []cdns.Record) map[string]cdns.Record {
	recordsMap := make(map[string]cdns.Record, 0)
	for _, r := range records {
		recordsMap[r.Record] = r
	}
	return recordsMap
}

// processUpdateActionsByZone processes update actions for a single zone.
func processUpdateActionsByZone(zoneID, zoneName string, records []cdns.Record, endpoints []*endpoint.Endpoint, changes *cloudnsChanges) {
	for _, ep := range endpoints {
		matchingRecords := getMatchingDomainRecords(records, zoneName, ep)

		if len(matchingRecords) == 0 {
			log.WithFields(log.Fields{
				"zoneName":   zoneName,
				"dnsName":    ep.DNSName,
				"recordType": ep.RecordType,
			}).Warn("Planning an update but no existing records found.")
		}

		matchingRecordsByTarget := getMatchingRecordsByTarget(matchingRecords)

		processUpdateEndpoint(zoneID, zoneName, matchingRecordsByTarget, ep, changes)

		// Any remaining records have been removed, delete them
		cleanupRemainingTargets(zoneID, matchingRecordsByTarget, changes)
	}
}

// processUpdateActions processes the update requests.
func processUpdateActions(
	zoneIDNameMapper provider.ZoneIDName,
	recordsByZoneID map[string][]cdns.Record,
	updatesByZoneID map[string][]*endpoint.Endpoint,
	changes *cloudnsChanges,
) {
	// Generate creates and updates based on existing
	for zoneID, endpoints := range updatesByZoneID {
		zoneName := zoneIDNameMapper[zoneID]
		if len(endpoints) == 0 {
			log.WithFields(log.Fields{
				"zoneName": zoneName,
			}).Debug("Skipping Zone, no updates found.")
			continue
		}

		records := recordsByZoneID[zoneID]
		processUpdateActionsByZone(zoneID, zoneName, records, endpoints, changes)

	}
}

// targetsMatch determines if a record matches one of the endpoint's targets.
func targetsMatch(record cdns.Record, ep *endpoint.Endpoint) bool {
	for _, t := range ep.Targets {
		endpointTarget := t
		recordTarget := record.Record
		if ep.RecordType == endpoint.RecordTypeCNAME {
			domain := record.Zone.Name
			endpointTarget = adjustCNAMETarget(domain, t)
		}
		if endpointTarget == recordTarget {
			return true
		}
	}
	return false
}

// processDeleteActionsByEndpoint processes delete actions for an endpoint.
func processDeleteActionsByEndpoint(zoneID string, matchingRecords []cdns.Record, ep *endpoint.Endpoint, changes *cloudnsChanges) {
	for _, record := range matchingRecords {
		if targetsMatch(record, ep) {
			changes.AddChangeDelete(zoneID, record)
		}
	}
}

// processDeleteActionsByZone processes delete actions for a single zone.
func processDeleteActionsByZone(zoneID, zoneName string, records []cdns.Record, endpoints []*endpoint.Endpoint, changes *cloudnsChanges) {
	for _, ep := range endpoints {
		matchingRecords := getMatchingDomainRecords(records, zoneName, ep)

		if len(matchingRecords) == 0 {
			log.WithFields(log.Fields{
				"zoneName":   zoneName,
				"dnsName":    ep.DNSName,
				"recordType": ep.RecordType,
			}).Warn("Records to delete not found.")
		}
		processDeleteActionsByEndpoint(zoneID, matchingRecords, ep, changes)
	}
}

// processDeleteActions processes the delete requests.
func processDeleteActions(
	zoneIDNameMapper provider.ZoneIDName,
	recordsByZoneID map[string][]cdns.Record,
	deletesByZoneID map[string][]*endpoint.Endpoint,
	changes *cloudnsChanges,
) {
	for zoneID, endpoints := range deletesByZoneID {
		zoneName := zoneIDNameMapper[zoneID]
		if len(endpoints) == 0 {
			log.WithFields(log.Fields{
				"zoneName": zoneName,
			}).Debug("Skipping Zone, no deletes found.")
			continue
		}

		records := recordsByZoneID[zoneID]
		processDeleteActionsByZone(zoneID, zoneName, records, endpoints, changes)

	}
}
