/*
 * Connector - functions for reading zones and records from ClouDNS DNS
 *
 * Copyright 2024 Marco Confalonieri.
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
	"context"
	"time"

	"external-dns-cloudns-webhook/internal/metrics"

	cdns "github.com/ppmathis/cloudns-go"
)

const (
	actGetZones     = "get_zones"
	actGetRecords   = "get_records"
	actCreateRecord = "create_record"
	actUpdateRecord = "update_record"
	actDeleteRecord = "delete_record"
)

// fetchRecords fetches all records for a given zoneID.
func fetchRecords(ctx context.Context, zone cdns.Zone, dnsClient *cdns.Client, batchSize int) (cdns.RecordMap, error) {
	metrics := metrics.GetOpenMetricsInstance()
	records := cdns.RecordMap{}

	start := time.Now()
	records, err := dnsClient.Records.List(ctx, zone.Name)
	if err != nil {
		metrics.IncFailedApiCallsTotal(actGetRecords)
		return nil, err
	}
	delay := time.Since(start)
	metrics.IncSuccessfulApiCallsTotal(actGetRecords)
	metrics.AddApiDelayHist(actGetRecords, delay.Milliseconds())

	return records, nil
}

// fetchZones fetches all the zones from the DNS client.
func fetchZones(ctx context.Context, dnsClient *cdns.Client, batchSize int) ([]cdns.Zone, error) {
	metrics := metrics.GetOpenMetricsInstance()
	zones := []cdns.Zone{}

	start := time.Now()
	zones, err := dnsClient.Zones.List(ctx)
	if err != nil {
		metrics.IncFailedApiCallsTotal(actGetZones)
		return nil, err
	}
	delay := time.Since(start)
	metrics.IncSuccessfulApiCallsTotal(actGetZones)
	metrics.AddApiDelayHist(actGetZones, delay.Milliseconds())

	return zones, nil
}
