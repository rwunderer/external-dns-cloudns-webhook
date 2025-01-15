/*
 * Changes Internals - Internal structures for processing changes.
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
	cdns "github.com/ppmathis/cloudns-go"
	log "github.com/sirupsen/logrus"
)

// cloudnsChangeCreate stores the information for a create request.
type cloudnsChangeCreate struct {
	ZoneID string
	Record cdns.Record
}

// GetLogFields returns the log fields for this object.
func (cc cloudnsChangeCreate) GetLogFields() log.Fields {
	return log.Fields{
		"zoneID":     cc.ZoneID,
		"dnsName":    cc.Record.Host,
		"recordType": string(cc.Record.RecordType),
		"value":      cc.Record.Record,
		"ttl":        string(cc.Record.TTL),
	}
}

// cloudnsChangeUpdate stores the information for an update request.
type cloudnsChangeUpdate struct {
	ZoneID string
	Record cdns.Record
}

// GetLogFields returns the log fields for this object. An asterisk indicate
// that the new value is shown.
func (cu cloudnsChangeUpdate) GetLogFields() log.Fields {
	return log.Fields{
		"zoneID":      cu.ZoneID,
		"recordID":    cu.Record.ID,
		"*dnsName":    cu.Record.Host,
		"*recordType": string(cu.Record.RecordType),
		"*value":      cu.Record.Record,
		"ttl":         string(cc.Record.TTL),
	}
}

// cloudnsChangeDelete stores the information for a delete request.
type cloudnsChangeDelete struct {
	ZoneID string
	Record cdns.Record
}

// GetLogFields returns the log fields for this object.
func (cd cloudnsChangeDelete) GetLogFields() log.Fields {
	return log.Fields{
		"zoneID":      cd.ZoneID,
		"*dnsName":    cu.Record.Host,
		"*recordType": string(cu.Record.RecordType),
		"*value":      cu.Record.Record,
	}
}
