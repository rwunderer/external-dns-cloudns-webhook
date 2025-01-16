/*
 * Change processors - unit tests.
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
	"testing"

	cdns "github.com/ppmathis/cloudns-go"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
)

// testTTL is a test ttl.
var testTTL = 7200

var testZoneIDMapper = provider.ZoneIDName{
	"alpha.com": "alpha.com",
	"beta.com":  "beta.com",
}

func assertEqualChanges(t *testing.T, expected, actual cloudnsChanges) {
	assert.Equal(t, expected.dryRun, actual.dryRun)
	assert.ElementsMatch(t, expected.creates, actual.creates)
	assert.ElementsMatch(t, expected.updates, actual.updates)
	assert.ElementsMatch(t, expected.deletes, actual.deletes)
}

// Test_adjustCNAMETarget tests adjustCNAMETarget()
func Test_adjustCNAMETarget(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			domain string
			target string
		}
		expected string
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := adjustCNAMETarget(inp.domain, inp.target)
		assert.Equal(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name: "target matches domain",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "www.alpha.com",
			},
			expected: "www",
		},
		{
			name: "target matches domain with dot",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "www.alpha.com.",
			},
			expected: "www",
		},
		{
			name: "target without dot does not match domain",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "www.beta.com",
			},
			expected: "www.beta.com.",
		},
		{
			name: "target with dot does not match domain",
			input: struct {
				domain string
				target string
			}{
				domain: "alpha.com",
				target: "www.beta.com.",
			},
			expected: "www.beta.com.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_processCreateActionsByZone tests processCreateActionsByZone().
func Test_processCreateActionsByZone(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			zoneID    string
			zoneName  string
			records   []cdns.Record
			endpoints []*endpoint.Endpoint
		}
		expectedChanges cloudnsChanges
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		changes := cloudnsChanges{}
		processCreateActionsByZone(inp.zoneID, inp.zoneName, inp.records,
			inp.endpoints, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "record already created",
			input: struct {
				zoneID    string
				zoneName  string
				records   []cdns.Record
				endpoints []*endpoint.Endpoint
			}{
				zoneID:   "alpha.com",
				zoneName: "alpha.com",
				records: []cdns.Record{
					{
						RecordType:  "A",
						Host:  "www",
						Record: "127.0.0.1",
						TTL:   7200,
					},
				},
				endpoints: []*endpoint.Endpoint{
					{
						DNSName:    "www.alpha.com",
						Targets:    endpoint.Targets{"127.0.0.1"},
						RecordType: "A",
						RecordTTL:  7200,
					},
				},
			},
			expectedChanges: cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "alpha.com",
						Record: &cdns.Record{
							Host:  "www",
							TTL:   &testTTL,
							RecordType:  "A",
							Record: "127.0.0.1",
						},
					},
				},
			},
		},
		{
			name: "new record created",
			input: struct {
				zoneID    string
				zoneName  string
				records   []cdns.Record
				endpoints []*endpoint.Endpoint
			}{
				zoneID:   "alpha.com",
				zoneName: "alpha.com",
				records: []cdns.Record{
					{
						RecordType:  "A",
						Host:  "ftp",
						Record: "127.0.0.1",
						Ttl:   7200,
					},
				},
				endpoints: []*endpoint.Endpoint{
					{
						DNSName:    "www.alpha.com",
						Targets:    endpoint.Targets{"127.0.0.1"},
						RecordType: "A",
						RecordTTL:  7200,
					},
				},
			},
			expectedChanges: cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "alpha.com",
						Record: &cdns.Record{
							Host:  "www",
							TTL:   &testTTL,
							RecordType:  "A",
							Record: "127.0.0.1",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_processCreateActions tests processCreateActions().
func Test_processCreateActions(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			zoneIDNameMapper provider.ZoneIDName
			recordsByZoneID  map[string][]cdns.Record
			createsByZoneID  map[string][]*endpoint.Endpoint
		}
		expectedChanges cloudnsChanges
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		changes := cloudnsChanges{}
		processCreateActions(inp.zoneIDNameMapper, inp.recordsByZoneID,
			inp.createsByZoneID, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "empty changeset",
			input: struct {
				zoneIDNameMapper provider.ZoneIDName
				recordsByZoneID  map[string][]cdns.Record
				createsByZoneID  map[string][]*endpoint.Endpoint
			}{
				zoneIDNameMapper: testZoneIDMapper,
				recordsByZoneID: map[string][]cdns.Record{
					"alpha.com": {
						cdns.Record{
							RecordType:  "A",
							Host:  "www",
							Record: "127.0.0.1",
						},
					},
				},
				createsByZoneID: map[string][]*endpoint.Endpoint{},
			},
		},
		{
			name: "empty changeset with key present",
			input: struct {
				zoneIDNameMapper provider.ZoneIDName
				recordsByZoneID  map[string][]cdns.Record
				createsByZoneID  map[string][]*endpoint.Endpoint
			}{
				zoneIDNameMapper: testZoneIDMapper,
				recordsByZoneID: map[string][]cdns.Record{
					"alpha.com": {
						cdns.Record{
							RecordType:  "A",
							Host:  "www",
							Record: "127.0.0.1",
						},
					},
				},
				createsByZoneID: map[string][]*endpoint.Endpoint{
					"alpha.com": {},
				},
			},
		},
		{
			name: "record already created",
			input: struct {
				zoneIDNameMapper provider.ZoneIDName
				recordsByZoneID  map[string][]cdns.Record
				createsByZoneID  map[string][]*endpoint.Endpoint
			}{
				zoneIDNameMapper: testZoneIDMapper,
				recordsByZoneID: map[string][]cdns.Record{
					"alpha.com": {
						cdns.Record{
							RecordType:  "A",
							Host:  "www",
							Record: "127.0.0.1",
							TTL:   7200,
						},
					},
				},
				createsByZoneID: map[string][]*endpoint.Endpoint{
					"alpha.com": {
						&endpoint.Endpoint{
							DNSName:    "www.alpha.com",
							Targets:    endpoint.Targets{"127.0.0.1"},
							RecordType: "A",
							RecordTTL:  7200,
						},
					},
				},
			},
			expectedChanges: cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "alpha.com",
						Record: &cdns.Record{
							Host:  "www",
							TTTL:   &testTTL,
							RecordType:  "A",
							Record: "127.0.0.1",
						},
					},
				},
			},
		},
		{
			name: "new record created",
			input: struct {
				zoneIDNameMapper provider.ZoneIDName
				recordsByZoneID  map[string][]cdns.Record
				createsByZoneID  map[string][]*endpoint.Endpoint
			}{
				zoneIDNameMapper: testZoneIDMapper,
				recordsByZoneID: map[string][]cdns.Record{
					"alpha.com": {
						cdns.Record{
							RecordType:  "A",
							Host:  "ftp",
							Record: "127.0.0.1",
							TTL:   7200,
						},
					},
				},
				createsByZoneID: map[string][]*endpoint.Endpoint{
					"alpha.com": {
						&endpoint.Endpoint{
							DNSName:    "www.alpha.com",
							Targets:    endpoint.Targets{"127.0.0.1"},
							RecordType: "A",
							RecordTTL:  7200,
						},
					},
				},
			},
			expectedChanges: cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "alpha.com",
						Record: &cdns.Record{
							Host:  "www",
							TTL:   &testTTL,
							RecordType:  "A",
							Record: "127.0.0.1",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_processUpdateEndpoint tests processUpdateEndpoint().
func Test_processUpdateEndpoint(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			zoneID                  string
			zoneName                string
			matchingRecordsByTarget map[string]cdns.Record
			ep                      *endpoint.Endpoint
		}
		expectedChanges cloudnsChanges
	}

	run := func(t *testing.T, tc testCase) {
		changes := cloudnsChanges{}
		inp := tc.input
		processUpdateEndpoint(inp.zoneID, inp.zoneName, inp.matchingRecordsByTarget,
			inp.ep, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "name changed",
			input: struct {
				zoneID                  string
				zoneName                string
				matchingRecordsByTarget map[string]cdns.Record
				ep                      *endpoint.Endpoint
			}{
				zoneID:   "alpha.com",
				zoneName: "alpha.com",
				matchingRecordsByTarget: map[string]cdns.Record{
					"1.1.1.1": {
						ID:   1,
						RecordType: cdns.RecordTypeA,
						Host: "www",
						Record: "1.1.1.1",
						TTL:   -1,
					},
				},
				ep: &endpoint.Endpoint{
					DNSName:    "ftp.alpha.com",
					RecordType: "A",
					Targets:    []string{"1.1.1.1"},
					RecordTTL:  -1,
				},
			},
			expectedChanges: cloudnsChanges{
				updates: []*cloudnsChangeUpdate{
					{
						ZoneID: "alpha.com",
						Record: cdns.Record{
							ID:   1,
							RecordType: cdns.RecordTypeA,
							Host: "www",
							Record: "1.1.1.1",
							TTL:   -1,
						},
					},
				},
			},
		},
		{
			name: "TTL changed",
			input: struct {
				zoneID                  string
				zoneName                string
				matchingRecordsByTarget map[string]cdns.Record
				ep                      *endpoint.Endpoint
			}{
				zoneID:   "alpha.com",
				zoneName: "alpha.com",
				matchingRecordsByTarget: map[string]cdns.Record{
					"1.1.1.1": {
						ID:   1,
						RecordType: cdns.RecordTypeA,
						Host: "www",
						Record: "1.1.1.1",
						TTL:   -1,
					},
				},
				ep: &endpoint.Endpoint{
					DNSName:    "ftp.alpha.com",
					RecordType: "A",
					Targets:    []string{"1.1.1.1"},
					RecordTTL:  7200,
				},
			},
			expectedChanges: cloudnsChanges{
				updates: []*cloudnsChangeUpdate{
					{
						ZoneID: "alpha.com",
						Record: cdns.Record{
							ID:   1,
							RecordType: cdns.RecordTypeA,
							Host: "www",
							Record: "1.1.1.1",
							Ttl:   -1,
						},
					},
				},
			},
		},
		{
			name: "target changed",
			input: struct {
				zoneID                  string
				zoneName                string
				matchingRecordsByTarget map[string]cdns.Record
				ep                      *endpoint.Endpoint
			}{
				zoneID:   "alpha.com",
				zoneName: "alpha.com",
				matchingRecordsByTarget: map[string]cdns.Record{
					"1.1.1.1": {
						ID:   1,
						Host: "www",
						RecordType: cdns.RecordTypeA,
						Record: "1.1.1.1",
						TTL:   -1,
					},
				},
				ep: &endpoint.Endpoint{
					DNSName:    "www.alpha.com",
					RecordType: "A",
					Targets:    []string{"2.2.2.2"},
					RecordTTL:  -1,
				},
			},
			expectedChanges: cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "alpha.com",
						Record: &cdns.Record{
							Host:  "www",
							TTL:   nil,
							RecordType:  cdns.RecordTypeA,
							Record: "2.2.2.2",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_cleanupRemainingTargets tests cleanupRemainingTargets().
func Test_cleanupRemainingTargets(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			zoneID                  string
			matchingRecordsByTarget map[string]cdns.Record
		}
		expectedChanges cloudnsChanges
	}

	run := func(t *testing.T, tc testCase) {
		changes := cloudnsChanges{}
		inp := tc.input
		cleanupRemainingTargets(inp.zoneID, inp.matchingRecordsByTarget,
			&changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "no deletes",
			input: struct {
				zoneID                  string
				matchingRecordsByTarget map[string]cdns.Record
			}{
				zoneID:                  "alpha.com",
				matchingRecordsByTarget: map[string]cdns.Record{},
			},
		},
		{
			name: "delete",
			input: struct {
				zoneID                  string
				matchingRecordsByTarget map[string]cdns.Record
			}{
				zoneID: "alpha.com",
				matchingRecordsByTarget: map[string]cdns.Record{
					"1.1.1.1": {
						ID:   1,
						Host: "www",
						RecordType: cdns.RecordTypeA,
						Record: "1.1.1.1",
						TTL:   -1,
					},
				},
			},
			expectedChanges: cloudnsChanges{
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "alpha.com",
						Record: cdns.Record{
							ID:   1,
							Host: "www",
							RecordType: cdns.RecordTypeA,
							Record: "1.1.1.1",
							TTL:   -1,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_getMatchingRecordsByTarget tests getMatchingRecordsByTarget().
func Test_getMatchingRecordsByTarget(t *testing.T) {
	type testCase struct {
		name     string
		input    []cdns.Record
		expected map[string]ccdns.Record
	}

	run := func(t *testing.T, tc testCase) {
		actual := getMatchingRecordsByTarget(tc.input)
		assert.EqualValues(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name:     "empty array",
			input:    []cdns.Record{},
			expected: map[string]cdns.Record{},
		},
		{
			name: "some values",
			input: []cdns.Record{
				{
					ID:   1,
					Host: "www",
					RecordType: cdns.RecordTypeA,
					Record: "1.1.1.1",
					TTL:   -1,
				},
				{
					ID:   2,
					Host: "ftp",
					RecordType: cdns.RecordTypeA,
					Record: "2.2.2.2",
					TTL:   -1,
				},
			},
			expected: map[string]cdns.Record{
				"1.1.1.1": {
					ID:   1,
					Host: "www",
					RecordType: cdns.RecordTypeA,
					Record: "1.1.1.1",
					TTL:   -1,
				},
				"2.2.2.2": {
					ID:   2,
					Host: "ftp",
					RecordType: cdns.RecordTypeA,
					Record: "2.2.2.2",
					TTL:   -1,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_processUpdateActionsByZone tests processUpdateActionsByZone().
func Test_processUpdateActionsByZone(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			zoneID    string
			zoneName  string
			records   []cdns.Record
			endpoints []*endpoint.Endpoint
		}
		expectedChanges cloudnsChanges
	}

	run := func(t *testing.T, tc testCase) {
		changes := cloudnsChanges{}
		inp := tc.input
		processUpdateActionsByZone(inp.zoneID, inp.zoneName, inp.records,
			inp.endpoints, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "empty changeset",
			input: struct {
				zoneID    string
				zoneName  string
				records   []cdns.Record
				endpoints []*endpoint.Endpoint
			}{
				zoneID:   "alpha.com",
				zoneName: "alpha.com",
				records: []cdns.Record{
					{
						ID:   1,
						Host: "www",
						Record: "1.1.1.1",
						TTL:   -1,
					},
				},
				endpoints: []*endpoint.Endpoint{},
			},
			expectedChanges: cloudnsChanges{},
		},
		{
			name: "mixed changeset",
			input: struct {
				zoneID    string
				zoneName  string
				records   []cdns.Record
				endpoints []*endpoint.Endpoint
			}{
				zoneID:   "alpha.com",
				zoneName: "alpha.com",
				records: []cdns.Record{
					{
						ID:   1,
						Host: "www",
						RecordType: cdns.RecordTypeA,
						Record: "1.1.1.1",
						TTL:   -1,
					},
					{
						ID:   "id_2",
						Host: "ftp",
						RecordType: cdns.RecordTypeA,
						Record: "2.2.2.2",
						TTL:   -1,
					},
				},
				endpoints: []*endpoint.Endpoint{
					{
						DNSName:    "www.alpha.com",
						RecordType: "A",
						Targets:    []string{"3.3.3.3"},
						RecordTTL:  -1,
					},
					{
						DNSName:    "ftp.alpha.com",
						RecordType: "A",
						Targets:    []string{"2.2.2.2"},
						RecordTTL:  7200,
					},
				},
			},
			expectedChanges: cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "alpha.com",
						Record: &cdns.Record{
							Host:  "www",
							TTL:   nil,
							RecordType:  cdns.RecordTypeA,
							Record: "3.3.3.3",
						},
					},
				},
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "alpha.com",
						Record: cdns.Record{
							ID:   1,
							Host: "www",
							RecordType: cdns.RecordTypeA,
							Record: "1.1.1.1",
							TTL:   -1,
						},
					},
				},
				updates: []*cloudnsChangeUpdate{
					{
						ZoneID: "alpha.com",
						Record: cdns.Record{
							ID:   2,
							Host: "ftp",
							RecordType: cdns.RecordTypeA,
							Record: "2.2.2.2",
							TTL:   -1,
						},
						Options: &cdns.RecordUpdateOpts{
							Host: "ftp",
							RecordType:  cdns.RecordTypeA,
							Record: "2.2.2.2",
							TTL:   &testTTL,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_processUpdateActions tests processUpdateActions().
func Test_processUpdateActions(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			zoneIDNameMapper provider.ZoneIDName
			recordsByZoneID  map[string][]cdns.Record
			updatesByZoneID  map[string][]*endpoint.Endpoint
		}
		expectedChanges cloudnsChanges
	}

	run := func(t *testing.T, tc testCase) {
		changes := cloudnsChanges{}
		inp := tc.input
		processUpdateActions(inp.zoneIDNameMapper, inp.recordsByZoneID,
			inp.updatesByZoneID, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "empty changeset",
			input: struct {
				zoneIDNameMapper provider.ZoneIDName
				recordsByZoneID  map[string][]cdns.Record
				updatesByZoneID  map[string][]*endpoint.Endpoint
			}{
				zoneIDNameMapper: provider.ZoneIDName{
					"alpha.com": "alpha.com",
					"beta.com":  "beta.com",
				},
				recordsByZoneID: map[string][]cdns.Record{
					"alpha.com": {
						cdns.Record{
							ID:   1,
							Host: "www",
							Record: "1.1.1.1",
							TTL:   -1,
						},
					},
					"beta.com": {
						cdns.Record{
							ID:   2,
							Host: "ftp",
							Record: "2.2.2.2",
							TTL:   -1,
						},
					},
				},
				updatesByZoneID: map[string][]*endpoint.Endpoint{},
			},
			expectedChanges: cloudnsChanges{},
		},
		{
			name: "empty changeset with key present",
			input: struct {
				zoneIDNameMapper provider.ZoneIDName
				recordsByZoneID  map[string][]cdns.Record
				updatesByZoneID  map[string][]*endpoint.Endpoint
			}{
				zoneIDNameMapper: provider.ZoneIDName{
					"alpha.com": "alpha.com",
					"beta.com":  "beta.com",
				},
				recordsByZoneID: map[string][]cdns.Record{
					"alpha.com": {
						cdns.Record{
							ID:   1,
							Host: "www",
							Record: "1.1.1.1",
							TTL:   -1,
						},
					},
					"beta.com": {
						cdns.Record{
							ID:   2,
							Host: "ftp",
							Record: "2.2.2.2",
							TTL:   -1,
						},
					},
				},
				updatesByZoneID: map[string][]*endpoint.Endpoint{
					"alpha.com": {},
					"beta.com":  {},
				},
			},
		},
		{
			name: "mixed changeset",
			input: struct {
				zoneIDNameMapper provider.ZoneIDName
				recordsByZoneID  map[string][]cdns.Record
				updatesByZoneID  map[string][]*endpoint.Endpoint
			}{
				zoneIDNameMapper: provider.ZoneIDName{
					"alpha.com": "alpha.com",
					"beta.com":  "beta.com",
				},
				recordsByZoneID: map[string][]cdns.Record{
					"alpha.com": {
						cdns.Record{
							ID:   1,
							Host: "www",
							RecordType: cdns.RecordTypeA,
							Record: "1.1.1.1",
							TTL:   -1,
						},
					},
					"beta.com": {
						cdns.Record{
							ID:   2,
							Name: "ftp",
							RecordType: cdns.RecordTypeA,
							Record: "2.2.2.2",
							TTL:   -1,
						},
					},
				},
				updatesByZoneID: map[string][]*endpoint.Endpoint{
					"alpha.com": {
						&endpoint.Endpoint{
							DNSName:    "www.alpha.com",
							RecordType: "A",
							Targets:    []string{"3.3.3.3"},
							RecordTTL:  -1,
						},
					},
					"beta.com": {
						&endpoint.Endpoint{
							DNSName:    "ftp.beta.com",
							RecordType: "A",
							Targets:    []string{"2.2.2.2"},
							RecordTTL:  7200,
						},
					},
				},
			},
			expectedChanges: cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "alpha.com",
						Record: &cdns.Record{
							Host:  "www",
							RecordType:  cdns.RecordTypeA,
							Record: "3.3.3.3",
							TTL: nil,
						},
					},
				},
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "alpha.com",
						Record: cdns.Record{
							ID:   1,
							Host: "www",
							RecordType: cdns.RecordTypeA,
							Record: "1.1.1.1",
							TTL:   -1,
						},
					},
				},
				updates: []*cloudnsChangeUpdate{
					{
						ZoneID: "beta.com",
						Record: cdns.Record{
							ID:   2,
							Host: "ftp",
							RecordType: cdns.RecordTypeA,
							Record: "2.2.2.2",
							TTL:   -1,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func Test_targetsMatch(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			record cdns.Record
			ep     *endpoint.Endpoint
		}
		expected bool
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := targetsMatch(inp.record, inp.ep)
		assert.EqualValues(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name: "record does not matches",
			input: struct {
				record cdns.Record
				ep     *endpoint.Endpoint
			}{
				record: cdns.Record{
					ID:   1,
					Host: "www",
					RecordType: cdns.RecordTypeA,
					Record: "1.1.1.1",
					TTL:   -1,
				},
				ep: &endpoint.Endpoint{
					DNSName:    "www.alpha.com",
					Targets:    endpoint.Targets{"7.7.7.7"},
					RecordType: "A",
					RecordTTL:  -1,
				},
			},
			expected: false,
		},
		{
			name: "record matches",
			input: struct {
				record cdns.Record
				ep     *endpoint.Endpoint
			}{
				record: cdns.Record{
					ID:   1,
					Host: "www",
					RecordType: cdns.RecordTypeA,
					Record: "1.1.1.1",
					TTL:   -1,
				},
				ep: &endpoint.Endpoint{
					DNSName:    "www.alpha.com",
					Targets:    endpoint.Targets{"1.1.1.1"},
					RecordType: "A",
					RecordTTL:  -1,
				},
			},
			expected: true,
		},
		{
			name: "cname special matching",
			input: struct {
				record cdns.Record
				ep     *endpoint.Endpoint
			}{
				record: cdns.Record{
					ID:   2,
					Host: "ftp",
					RecordType: cdns.RecordTypeCNAME,
					Record: "www.beta.com.",
					TTL:   -1,
				},
				ep: &endpoint.Endpoint{
					DNSName:    "ftp.alpha.com",
					Targets:    endpoint.Targets{"www.beta.com"},
					RecordType: "CNAME",
					RecordTTL:  -1,
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_processDeleteActionsByEndpoint tests processDeleteActionsByEndpoint().
func Test_processDeleteActionsByEndpoint(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			zoneID          string
			matchingRecords []cdns.Record
			ep              *endpoint.Endpoint
		}
		expectedChanges cloudnsChanges
	}

	run := func(t *testing.T, tc testCase) {
		changes := cloudnsChanges{}
		inp := tc.input
		processDeleteActionsByEndpoint(inp.zoneID, inp.matchingRecords,
			inp.ep, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "no matching records",
			input: struct {
				zoneID          string
				matchingRecords []cdns.Record
				ep              *endpoint.Endpoint
			}{
				zoneID:          "alpha.com",
				matchingRecords: []cdns.Record{},
				ep: &endpoint.Endpoint{
					DNSName:    "ccx.alpha.com",
					Targets:    endpoint.Targets{"7.7.7.7"},
					RecordType: "A",
					RecordTTL:  7200,
				},
			},
			expectedChanges: cloudnsChanges{},
		},
		{
			name: "one matching record",
			input: struct {
				zoneID          string
				matchingRecords []cdns.Record
				ep              *endpoint.Endpoint
			}{
				zoneID: "alpha.com",
				matchingRecords: []cdns.Record{
					{
						ID:   1,
						Host: "www",
						RecordType: cdns.RecordTypeA,
						Record: "1.1.1.1",
						TTL:   -1,
					},
					{
						ID:   2,
						Host: "www",
						RecordType: cdns.RecordTypeA,
						Record: "2.2.2.2",
						TTL:   -1,
					},
				},
				ep: &endpoint.Endpoint{
					DNSName:    "www.alpha.com",
					Targets:    endpoint.Targets{"1.1.1.1"},
					RecordType: "A",
					RecordTTL:  -1,
				},
			},
			expectedChanges: cloudnsChanges{
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "alpha.com",
						Record: cdns.Record{
							ID:   1,
							Host: "www",
							RecordType: cdns.RecordTypeA,
							Record: "1.1.1.1",
							TTL:   -1,
						},
					},
				},
			},
		},
		{
			name: "cname special matching",
			input: struct {
				zoneID          string
				matchingRecords []cdns.Record
				ep              *endpoint.Endpoint
			}{
				zoneID: "alpha.com",
				matchingRecords: []cdns.Record{
					{
						ID:   1,
						Host: "www",
						RecordType: cdns.RecordTypeA,
						Record: "1.1.1.1",
						TTL:   -1,
					},
					{
						ID:   2,
						Host: "ftp",
						RecordType: cdns.RecordTypeCNAME,
						Record: "www.beta.com.",
						TTL:   -1,
					},
				},
				ep: &endpoint.Endpoint{
					DNSName:    "ftp.alpha.com",
					Targets:    endpoint.Targets{"www.beta.com"},
					RecordType: "CNAME",
					RecordTTL:  -1,
				},
			},
			expectedChanges: cloudnsChanges{
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "alpha.com",
						Record: cdns.Record{
							ID:   2,
							Host: "ftp",
							RecordType: cdns.RecordTypeCNAME,
							Record: "www.beta.com.",
							TTL:   -1,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_processDeleteActions tests processDeleteActions().
func Test_processDeleteActions(t *testing.T) {
	type testCase struct {
		name  string
		input struct {
			zoneIDNameMapper provider.ZoneIDName
			recordsByZoneID  map[string][]cdns.Record
			deletesByZoneID  map[string][]*endpoint.Endpoint
		}
		expectedChanges cloudnsChanges
	}

	run := func(t *testing.T, tc testCase) {
		changes := cloudnsChanges{}
		inp := tc.input
		processDeleteActions(inp.zoneIDNameMapper, inp.recordsByZoneID,
			inp.deletesByZoneID, &changes)
		assertEqualChanges(t, tc.expectedChanges, changes)
	}

	testCases := []testCase{
		{
			name: "No deletes created",
			input: struct {
				zoneIDNameMapper provider.ZoneIDName
				recordsByZoneID  map[string][]cdns.Record
				deletesByZoneID  map[string][]*endpoint.Endpoint
			}{
				zoneIDNameMapper: provider.ZoneIDName{
					"alpha.com": "alpha.com",
					"beta.com":  "beta.com",
				},
				recordsByZoneID: map[string][]cdns.Record{
					"alpha.com": {
						cdns.Record{
							ID:   1,
							Host: "www",
							Record: "1.1.1.1",
							TTL:   -1,
						},
					},
					"beta.com": {
						cdns.Record{
							ID:   2,
							Host: "ftp",
							Record: "2.2.2.2",
							TTL:   -1,
						},
					},
				},
				deletesByZoneID: map[string][]*endpoint.Endpoint{
					"alpha.com": {
						&endpoint.Endpoint{
							DNSName:    "ccx.alpha.com",
							Targets:    endpoint.Targets{"7.7.7.7"},
							RecordType: "A",
							RecordTTL:  7200,
						},
					},
				},
			},
			expectedChanges: cloudnsChanges{},
		},
		{
			name: "deletes performed",
			input: struct {
				zoneIDNameMapper provider.ZoneIDName
				recordsByZoneID  map[string][]cdns.Record
				deletesByZoneID  map[string][]*endpoint.Endpoint
			}{
				zoneIDNameMapper: provider.ZoneIDName{
					"alpha.com": "alpha.com",
					"beta.com":  "beta.com",
				},
				recordsByZoneID: map[string][]cdns.Record{
					"alpha.com": {
						cdns.Record{
							ID:   1,
							Host: "www",
							RecordType: cdns.RecordTypeA,
							Record: "1.1.1.1",
							TTL:   -1,
						},
					},
					"beta.com": {
						cdns.Record{
							ID:   2,
							Host: "ftp",
							RecordType: cdns.RecordTypeA,
							Record: "2.2.2.2",
							TTL:   -1,
						},
						cdns.Record{
							ID:   3,
							Host: "ftp",
							RecordType: cdns.RecordTypeA,
							Record: "4.4.4.4",
							TTL:   -1,
						},
					},
				},
				deletesByZoneID: map[string][]*endpoint.Endpoint{
					"alpha.com": {
						&endpoint.Endpoint{
							DNSName:    "www.alpha.com",
							Targets:    endpoint.Targets{"1.1.1.1"},
							RecordType: "A",
							RecordTTL:  -1,
						},
					},
					"beta.com": {
						&endpoint.Endpoint{
							DNSName:    "ftp.beta.com",
							Targets:    endpoint.Targets{"2.2.2.2", "4.4.4.4"},
							RecordType: "A",
							RecordTTL:  -1,
						},
					},
				},
			},
			expectedChanges: cloudnsChanges{
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "alpha.com",
						Record: cdns.Record{
							ID:   1,
							Host: "www",
							RecordType: cdns.RecordTypeA,
							Record: "1.1.1.1",
							TTL:   -1,
						},
					},
					{
						ZoneID: "beta.com",
						Record: cdns.Record{
							ID:   2,
							Host: "ftp",
							RecordType: cdns.RecordTypeA,
							Record: "2.2.2.2",
							TTL:   -1,
						},
					},
					{
						ZoneID: "beta.com",
						Record: cdns.Record{
							ID:   3,
							Host: "ftp",
							RecordType: cdns.RecordTypeA,
							Record: "4.4.4.4",
							TTL:   -1,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
