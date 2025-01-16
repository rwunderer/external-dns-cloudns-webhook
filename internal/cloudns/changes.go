/*
 * Changes - Code for storing changes and sending them to the DNS API.
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
	log "github.com/sirupsen/logrus"
)

// cloudnsChange contains all changes to apply to DNS.
type cloudnsChanges struct {
	dryRun     bool
	defaultTTL int

	creates []*cloudnsChangeCreate
	updates []*cloudnsChangeUpdate
	deletes []*cloudnsChangeDelete
}

// empty returns true if there are no changes left.
func (c *cloudnsChanges) empty() bool {
	return len(c.creates) == 0 && len(c.updates) == 0 && len(c.deletes) == 0
}

// AddChangeCreate adds a new creation entry to the current object.
func (c *cloudnsChanges) AddChangeCreate(zoneID string, record *cdns.Record) {
	changeCreate := &cloudnsChangeCreate{
		ZoneID: zoneID,
		Record: record,
	}
	c.creates = append(c.creates, changeCreate)
}

// AddChangeUpdate adds a new update entry to the current object.
func (c *cloudnsChanges) AddChangeUpdate(zoneID string, record cdns.Record) {
	changeUpdate := &cloudnsChangeUpdate{
		ZoneID: zoneID,
		Record: record,
	}
	c.updates = append(c.updates, changeUpdate)
}

// AddChangeDelete adds a new delete entry to the current object.
func (c *cloudnsChanges) AddChangeDelete(zoneID string, record cdns.Record) {
	changeDelete := &cloudnsChangeDelete{
		ZoneID: zoneID,
		Record: record,
	}
	c.deletes = append(c.deletes, changeDelete)
}

// applyDeletes processes the records to be deleted.
func (c cloudnsChanges) applyDeletes(ctx context.Context, dnsClient *cdns.Client) error {
	metrics := metrics.GetOpenMetricsInstance()
	for _, e := range c.deletes {
		log.WithFields(e.GetLogFields()).Debug("Deleting domain record")
		log.Infof("Deleting record [%s] from zone [%s]", e.Record.Name, e.Record.Zone.Name)
		if c.dryRun {
			continue
		}
		start := time.Now()
		if _, err := dnsClient.DeleteRecord(ctx, e.Record); err != nil {
			metrics.IncFailedApiCallsTotal(actDeleteRecord)
			return err
		}
		delay := time.Since(start)
		metrics.IncSuccessfulApiCallsTotal(actDeleteRecord)
		metrics.AddApiDelayHist(actDeleteRecord, delay.Milliseconds())
	}
	return nil
}

// applyCreates processes the records to be created.
func (c cloudnsChanges) applyCreates(ctx context.Context, dnsClient *cdns.Client) error {
	metrics := metrics.GetOpenMetricsInstance()
	for _, e := range c.creates {
		rec := e.Record
		if rec.TTL == nil {
			ttl := c.defaultTTL
			rec.TTL = ttl
		}
		log.WithFields(e.GetLogFields()).Debug("Creating domain record")
		log.Infof("Creating record [%s] of type [%s] with value [%s] in zone [%s]",
			rec.Name, rec.Type, rec.Value, e.ZoneID)
		if c.dryRun {
			continue
		}
		start := time.Now()
		if _, _, err := dnsClient.Records.Create(ctx, e.ZoneID, rec); err != nil {
			metrics.IncFailedApiCallsTotal(actCreateRecord)
			return err
		}
		delay := time.Since(start)
		metrics.IncSuccessfulApiCallsTotal(actCreateRecord)
		metrics.AddApiDelayHist(actCreateRecord, delay.Milliseconds())
	}
	return nil
}

// applyUpdates processes the records to be updated.
func (c cloudnsChanges) applyUpdates(ctx context.Context, dnsClient *cdns.Client) error {
	metrics := metrics.GetOpenMetricsInstance()
	for _, e := range c.updates {
		rec := e.Record
		if rec.TTL == nil {
			ttl := c.defaultTTL
			rec.TTL = ttl
		}
		log.WithFields(e.GetLogFields()).Debug("Updating domain record")
		log.Infof("Updating record ID [%s] with name [%s], type [%s], value [%s] and TTL [%d] in zone [%s]",
			e.Record.ID, rec.Name, rec.Type, rec.Value, rec.TTL, e.ZoneID)
		if c.dryRun {
			continue
		}
		start := time.Now()
		if _, _, err := dnsClient.Records.Update(ctx, e.ZoneID, rec); err != nil {
			metrics.IncFailedApiCallsTotal(actUpdateRecord)
			return err
		}
		delay := time.Since(start)
		metrics.IncSuccessfulApiCallsTotal(actUpdateRecord)
		metrics.AddApiDelayHist(actUpdateRecord, delay.Milliseconds())
	}
	return nil
}

// ApplyChanges applies the planned changes using dnsClient.
func (c cloudnsChanges) ApplyChanges(ctx context.Context, dnsClient *cdns.Client) error {
	// No changes = nothing to do.
	if c.empty() {
		log.Debug("No changes to be applied found.")
		return nil
	}
	// Process records to be deleted.
	if err := c.applyDeletes(ctx, dnsClient); err != nil {
		return err
	}
	// Process record creations.
	if err := c.applyCreates(ctx, dnsClient); err != nil {
		return err
	}
	// Process record updates.
	if err := c.applyUpdates(ctx, dnsClient); err != nil {
		return err
	}
	return nil
}
