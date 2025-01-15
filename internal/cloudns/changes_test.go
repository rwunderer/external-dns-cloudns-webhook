/*
 * Changes - unit tests.
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
	"errors"
	"testing"

	cdns "github.com/ppmathis/cloudns-go"
	"github.com/stretchr/testify/assert"
)

// Test_cloudnsChanges_empty tests cloudnsChanges.empty().
func Test_cloudnsChanges_empty(t *testing.T) {
	type testCase struct {
		name     string
		changes  cloudnsChanges
		expected bool
	}

	run := func(t *testing.T, tc testCase) {
		actual := tc.changes.empty()
		assert.Equal(t, actual, tc.expected)
	}

	testCases := []testCase{
		{
			name:     "Empty",
			changes:  cloudnsChanges{},
			expected: true,
		},
		{
			name: "Creations",
			changes: cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID:  "alphaZoneID",
						Options: &cdns.RecordCreateOpts{},
					},
				},
			},
		},
		{
			name: "Updates",
			changes: cloudnsChanges{
				updates: []*cloudnsChangeUpdate{
					{
						ZoneID:  "alphaZoneID",
						Record:  cdns.Record{},
						Options: &cdns.RecordUpdateOpts{},
					},
				},
			},
		},
		{
			name: "Deletions",
			changes: cloudnsChanges{
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "alphaZoneID",
						Record: cdns.Record{},
					},
				},
			},
		},
		{
			name: "All",
			changes: cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID:  "alphaZoneID",
						Options: &cdns.RecordCreateOpts{},
					},
				},
				updates: []*cloudnsChangeUpdate{
					{
						ZoneID:  "alphaZoneID",
						Record:  cdns.Record{},
						Options: &cdns.RecordUpdateOpts{},
					},
				},
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "alphaZoneID",
						Record: cdns.Record{},
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

// Test_cloudnsChanges_AddChangeCreate tests cloudnsChanges.AddChangeCreate().
func Test_cloudnsChanges_AddChangeCreate(t *testing.T) {
	type testCase struct {
		name     string
		instance cloudnsChanges
		input    struct {
			zoneID  string
			options *cdns.RecordCreateOpts
		}
		expected cloudnsChanges
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := tc.instance
		actual.AddChangeCreate(inp.zoneID, inp.options)
		assert.EqualValues(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name:     "add create",
			instance: cloudnsChanges{},
			input: struct {
				zoneID  string
				options *cdns.RecordCreateOpts
			}{
				zoneID: "zoneIDAlpha",
				options: &cdns.RecordCreateOpts{
					Name:  "www",
					Ttl:   &testTTL,
					Type:  "A",
					Value: "127.0.0.1",
					Zone: &cdns.Zone{
						ID:   "zoneIDAlpha",
						Name: "alpha.com",
					},
				},
			},
			expected: cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "zoneIDAlpha",
						Options: &cdns.RecordCreateOpts{
							Name:  "www",
							Ttl:   &testTTL,
							Type:  "A",
							Value: "127.0.0.1",
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
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

// Test_cloudnsChanges_AddChangeUpdate tests cloudnsChanges.AddChangeUpdate().
func Test_cloudnsChanges_AddChangeUpdate(t *testing.T) {
	type testCase struct {
		name     string
		instance cloudnsChanges
		input    struct {
			zoneID  string
			record  cdns.Record
			options *cdns.RecordUpdateOpts
		}
		expected cloudnsChanges
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := tc.instance
		actual.AddChangeUpdate(inp.zoneID, inp.record, inp.options)
		assert.EqualValues(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name:     "add update",
			instance: cloudnsChanges{},
			input: struct {
				zoneID  string
				record  cdns.Record
				options *cdns.RecordUpdateOpts
			}{
				zoneID: "zoneIDAlpha",
				record: cdns.Record{
					ID:    "id_1",
					Name:  "www",
					Ttl:   -1,
					Type:  "A",
					Value: "127.0.0.1",
					Zone: &cdns.Zone{
						ID:   "zoneIDAlpha",
						Name: "alpha.com",
					},
				},
				options: &cdns.RecordUpdateOpts{
					Name:  "www",
					Ttl:   &testTTL,
					Type:  "A",
					Value: "127.0.0.1",
					Zone: &cdns.Zone{
						ID:   "zoneIDAlpha",
						Name: "alpha.com",
					},
				},
			},
			expected: cloudnsChanges{
				updates: []*cloudnsChangeUpdate{
					{
						ZoneID: "zoneIDAlpha",
						Record: cdns.Record{
							ID:    "id_1",
							Name:  "www",
							Ttl:   -1,
							Type:  "A",
							Value: "127.0.0.1",
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
						},
						Options: &cdns.RecordUpdateOpts{
							Name:  "www",
							Ttl:   &testTTL,
							Type:  "A",
							Value: "127.0.0.1",
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
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

// addChangeDelete adds a new delete entry to the current object.
func Test_cloudnsChanges_AddChangeDelete(t *testing.T) {
	type testCase struct {
		name     string
		instance cloudnsChanges
		input    struct {
			zoneID string
			record cdns.Record
		}
		expected cloudnsChanges
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		actual := tc.instance
		actual.AddChangeDelete(inp.zoneID, inp.record)
		assert.EqualValues(t, tc.expected, actual)
	}

	testCases := []testCase{
		{
			name:     "add update",
			instance: cloudnsChanges{},
			input: struct {
				zoneID string
				record cdns.Record
			}{
				zoneID: "zoneIDAlpha",
				record: cdns.Record{
					ID:    "id_1",
					Name:  "www",
					Ttl:   -1,
					Type:  "A",
					Value: "127.0.0.1",
					Zone: &cdns.Zone{
						ID:   "zoneIDAlpha",
						Name: "alpha.com",
					},
				},
			},
			expected: cloudnsChanges{
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "zoneIDAlpha",
						Record: cdns.Record{
							ID:    "id_1",
							Name:  "www",
							Ttl:   -1,
							Type:  "A",
							Value: "127.0.0.1",
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
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

// applyDeletes processes the records to be deleted.
func Test_cloudnsChanges_applyDeletes(t *testing.T) {
	type testCase struct {
		name     string
		changes  *cloudnsChanges
		input    *mockClient
		expected struct {
			state mockClientState
			err   error
		}
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		exp := tc.expected
		err := tc.changes.applyDeletes(context.Background(), inp)
		assertError(t, exp.err, err)
		assert.Equal(t, exp.state, inp.GetState())
	}

	testCases := []testCase{
		{
			name: "deletion",
			changes: &cloudnsChanges{
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "zoneIDAlpha",
						Record: cdns.Record{
							ID:    "id1",
							Type:  cdns.RecordTypeA,
							Name:  "www",
							Value: "1.1.1.1",
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Ttl: -1,
						},
					},
				},
			},
			input: &mockClient{},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{DeleteRecordCalled: true},
			},
		},
		{
			name: "deletion error",
			changes: &cloudnsChanges{
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "zoneIDAlpha",
						Record: cdns.Record{
							ID:    "id1",
							Type:  cdns.RecordTypeA,
							Name:  "www",
							Value: "1.1.1.1",
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Ttl: -1,
						},
					},
				},
			},
			input: &mockClient{
				deleteRecord: deleteResponse{
					err: errors.New("test delete error"),
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{DeleteRecordCalled: true},
				err:   errors.New("test delete error"),
			},
		},
		{
			name: "deletion dry run",
			changes: &cloudnsChanges{
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "zoneIDAlpha",
						Record: cdns.Record{
							ID:    "id1",
							Type:  cdns.RecordTypeA,
							Name:  "www",
							Value: "1.1.1.1",
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Ttl: -1,
						},
					},
				},
				dryRun: true,
			},
			input: &mockClient{},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// applyCreates processes the records to be created.
func Test_cloudnsChanges_applyCreates(t *testing.T) {
	type testCase struct {
		name     string
		changes  *cloudnsChanges
		input    *mockClient
		expected struct {
			state mockClientState
			err   error
		}
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		exp := tc.expected
		err := tc.changes.applyCreates(context.Background(), inp)
		assertError(t, exp.err, err)
		assert.Equal(t, exp.state, inp.GetState())
	}

	testCases := []testCase{
		{
			name: "creation",
			changes: &cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "zoneIDAlpha",
						Options: &cdns.RecordCreateOpts{
							Name: "www",
							Type: cdns.RecordTypeA,
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Value: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			input: &mockClient{},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{CreateRecordCalled: true},
			},
		},
		{
			name: "creation error",
			changes: &cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "zoneIDAlpha",
						Options: &cdns.RecordCreateOpts{
							Name: "www",
							Type: cdns.RecordTypeA,
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Value: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			input: &mockClient{
				createRecord: recordResponse{
					err: errors.New("test creation error"),
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{CreateRecordCalled: true},
				err:   errors.New("test creation error"),
			},
		},
		{
			name: "creation dry run",
			input: &mockClient{
				createRecord: recordResponse{
					err: errors.New("test creation error"),
				},
			},
			changes: &cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "zoneIDAlpha",
						Options: &cdns.RecordCreateOpts{
							Name: "www",
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
				dryRun: true,
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// applyUpdates processes the records to be updated.
func Test_cloudnsChanges_applyUpdates(t *testing.T) {
	type testCase struct {
		name     string
		changes  *cloudnsChanges
		input    *mockClient
		expected struct {
			state mockClientState
			err   error
		}
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		exp := tc.expected
		err := tc.changes.applyUpdates(context.Background(), inp)
		assertError(t, exp.err, err)
		assert.Equal(t, exp.state, inp.GetState())
	}

	testCases := []testCase{
		{
			name:  "update",
			input: &mockClient{},
			changes: &cloudnsChanges{
				updates: []*cloudnsChangeUpdate{
					{
						ZoneID: "zoneIDAlpha",
						Record: cdns.Record{
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Name:  "www",
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   testTTL,
						},
						Options: &cdns.RecordUpdateOpts{
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Name:  "ftp",
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{UpdateRecordCalled: true},
			},
		},
		{
			name: "update error",
			input: &mockClient{
				updateRecord: recordResponse{
					err: errors.New("test update error"),
				},
			},
			changes: &cloudnsChanges{
				updates: []*cloudnsChangeUpdate{
					{
						ZoneID: "zoneIDAlpha",
						Record: cdns.Record{
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Name:  "www",
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   testTTL,
						},
						Options: &cdns.RecordUpdateOpts{
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Name:  "ftp",
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{UpdateRecordCalled: true},
				err:   errors.New("test update error"),
			},
		},
		{
			name:  "update dry run",
			input: &mockClient{},
			changes: &cloudnsChanges{
				updates: []*cloudnsChangeUpdate{
					{
						ZoneID: "zoneIDAlpha",
						Record: cdns.Record{
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Name:  "www",
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   testTTL,
						},
						Options: &cdns.RecordUpdateOpts{
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Name:  "ftp",
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
				dryRun: true,
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

// Test_cloudnsChanges_ApplyChanges tests cloudnsChanges.ApplyChanges().
func Test_cloudnsChanges_ApplyChanges(t *testing.T) {
	type testCase struct {
		name     string
		changes  *cloudnsChanges
		input    *mockClient
		expected struct {
			state mockClientState
			err   error
		}
	}

	run := func(t *testing.T, tc testCase) {
		inp := tc.input
		exp := tc.expected
		err := tc.changes.ApplyChanges(context.Background(), inp)
		assertError(t, exp.err, err)
		assert.Equal(t, exp.state, inp.GetState())
	}

	testCases := []testCase{
		{
			name:    "no changes",
			changes: &cloudnsChanges{},
			input:   &mockClient{},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{},
			},
		},
		{
			name: "all changes",
			changes: &cloudnsChanges{
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "zoneIDAlpha",
						Record: cdns.Record{
							ID:    "id1",
							Type:  cdns.RecordTypeA,
							Name:  "www",
							Value: "1.1.1.1",
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Ttl: -1,
						},
					},
				},
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "zoneIDAlpha",
						Options: &cdns.RecordCreateOpts{
							Name: "www",
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
				updates: []*cloudnsChangeUpdate{
					{
						ZoneID: "zoneIDAlpha",
						Record: cdns.Record{
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Name:  "www",
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   testTTL,
						},
						Options: &cdns.RecordUpdateOpts{
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Name:  "ftp",
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			input: &mockClient{},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{
					CreateRecordCalled: true,
					DeleteRecordCalled: true,
					UpdateRecordCalled: true,
				},
			},
		},
		{
			name: "deletion error",
			changes: &cloudnsChanges{
				deletes: []*cloudnsChangeDelete{
					{
						ZoneID: "zoneIDAlpha",
						Record: cdns.Record{
							ID:    "id1",
							Type:  cdns.RecordTypeA,
							Name:  "www",
							Value: "1.1.1.1",
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Ttl: -1,
						},
					},
				},
			},
			input: &mockClient{
				deleteRecord: deleteResponse{
					err: errors.New("test delete error"),
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{
					DeleteRecordCalled: true,
				},
				err: errors.New("test delete error"),
			},
		},
		{
			name: "creation error",
			changes: &cloudnsChanges{
				creates: []*cloudnsChangeCreate{
					{
						ZoneID: "zoneIDAlpha",
						Options: &cdns.RecordCreateOpts{
							Name: "www",
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			input: &mockClient{
				createRecord: recordResponse{
					err: errors.New("test creation error"),
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{
					CreateRecordCalled: true,
				},
				err: errors.New("test creation error"),
			},
		},
		{
			name: "update error",
			input: &mockClient{
				updateRecord: recordResponse{
					err: errors.New("test update error"),
				},
			},
			changes: &cloudnsChanges{
				updates: []*cloudnsChangeUpdate{
					{
						ZoneID: "zoneIDAlpha",
						Record: cdns.Record{
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Name:  "www",
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   testTTL,
						},
						Options: &cdns.RecordUpdateOpts{
							Zone: &cdns.Zone{
								ID:   "zoneIDAlpha",
								Name: "alpha.com",
							},
							Name:  "ftp",
							Type:  "A",
							Value: "127.0.0.1",
							Ttl:   &testTTL,
						},
					},
				},
			},
			expected: struct {
				state mockClientState
				err   error
			}{
				state: mockClientState{
					UpdateRecordCalled: true,
				},
				err: errors.New("test update error"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			run(t, tc)
		})
	}
}
