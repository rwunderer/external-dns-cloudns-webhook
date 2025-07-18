package cloudns

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/codingconcepts/env"
	cloudns "github.com/ppmathis/cloudns-go"
	log "github.com/sirupsen/logrus"

	"sigs.k8s.io/external-dns/endpoint"
)

// var mockProvider = &ClouDNSProvider{}
var mockZones = []cloudns.Zone{
	{
		Name:     "test1.com",
		Type:     1,
		Kind:     1,
		IsActive: true,
	},
	{
		Name:     "test2.com",
		Type:     1,
		Kind:     1,
		IsActive: true,
	},
}

var mockRecords = [][]cloudns.Record{
	{
		{
			ID:         1,
			Host:       "",
			Record:     "1.1.1.1",
			RecordType: "A",
			TTL:        60,
			IsActive:   true,
		},
		{
			ID:         2,
			Host:       "sub2",
			Record:     "2.2.2.2",
			RecordType: "A",
			TTL:        60,
			IsActive:   true,
		},
		{
			ID:         3,
			Host:       "sub3",
			Record:     "3.3.3.3",
			RecordType: "A",
			TTL:        60,
			IsActive:   true,
		},
		{
			ID:         4,
			Host:       "",
			Record:     "TextRecord",
			RecordType: "TXT",
			TTL:        60,
			IsActive:   true,
		},
		{
			ID:         5,
			Host:       "sub5",
			Record:     "SubTextRecord",
			RecordType: "TXT",
			TTL:        60,
			IsActive:   true,
		},
	},
	{
		{
			ID:         6,
			Host:       "",
			Record:     "6.6.6.6",
			RecordType: "A",
			TTL:        60,
			IsActive:   true,
		},
		{
			ID:         7,
			Host:       "sub7",
			Record:     "7.7.7.7",
			RecordType: "A",
			TTL:        60,
			IsActive:   true,
		},
		{
			ID:         8,
			Host:       "sub8",
			Record:     "8.8.8.8",
			RecordType: "A",
			TTL:        60,
			IsActive:   true,
		},
		{
			ID:         9,
			Host:       "",
			Record:     "TextRecord",
			RecordType: "TXT",
			TTL:        60,
			IsActive:   true,
		},
		{
			ID:         10,
			Host:       "sub5",
			Record:     "SubTextRecord",
			RecordType: "TXT",
			TTL:        60,
			IsActive:   true,
		},
	},
}

/*
var expectedEndpointsOne = []*endpoint.Endpoint{
    // endpoint 1
    endpoint.NewEndpointWithTTL(
        "test1.com",
        "A",
        endpoint.TTL(60),
        "1.1.1.1",
    ),
    // endpoint 2
    endpoint.NewEndpointWithTTL(
        "sub2.test1.com",
        "A",
        endpoint.TTL(60),
        "2.2.2.2",
    ),
    // endpoint 3
    endpoint.NewEndpointWithTTL(
        "sub3.test1.com",
        "A",
        endpoint.TTL(60),
        "3.3.3.3",
    ),
    // endpoint 4
    endpoint.NewEndpointWithTTL(
        "test1.com",
        "TXT",
        endpoint.TTL(60),
        "TextRecord",
    ),
    // endpoint 5
    endpoint.NewEndpointWithTTL(
        "sub5.test1.com",
        "TXT",
        endpoint.TTL(60),
        "SubTextRecord",
    ),
    // endpoint 6
    endpoint.NewEndpointWithTTL(
        "test2.com",
        "A",
        endpoint.TTL(60),
        "6.6.6.6",
    ),
    // endpoint 7
    endpoint.NewEndpointWithTTL(
        "sub7.test2.com",
        "A",
        endpoint.TTL(60),
        "7.7.7.7",
    ),
    // endpoint 8
    endpoint.NewEndpointWithTTL(
        "sub8.test2.com",
        "A",
        endpoint.TTL(60),
        "8.8.8.8",
    ),
    // endpoint 9
    endpoint.NewEndpointWithTTL(
        "test2.com",
        "TXT",
        endpoint.TTL(60),
        "TextRecord",
    ),
    // endpoint 10
    endpoint.NewEndpointWithTTL(
        "sub5.test2.com",
        "TXT",
        endpoint.TTL(60),
        "SubTextRecord",
    ),
}
*/

// NewClouDNSProvider creates a new ClouDNSProvider using the specified ClouDNSConfig.
// It authenticates with ClouDNS using the login type specified in the CLOUDNS_LOGIN_TYPE environment variable,
// which can be "user-id", "sub-user", or "sub-user-name". If the CLOUDNS_USER_PASSWORD environment variable is not set,
// an error will be returned. If the CLOUDNS_USER_ID or CLOUDNS_SUB_USER_ID environment variables are not set or are not valid integers,
// an error will be returned. If the CLOUDNS_SUB_USER_NAME environment variable is not set, an error will be returned.
// config is the ClouDNSConfig to be used for creating the ClouDNSProvider.
// It returns the created ClouDNSProvider and a possible error.code
// NewClouDNSProvider creates a new ClouDNSProvider using the specified ClouDNSConfig.
// It authenticates with ClouDNS using the login type specified in the CLOUDNS_LOGIN_TYPE environment variable,
// which can be "user-id", "sub-user", or "sub-user-name". If the CLOUDNS_USER_PASSWORD environment variable is not set,
// an error will be returned. If the CLOUDNS_USER_ID or CLOUDNS_SUB_USER_ID environment variables are not set or are not valid integers,
// an error will be returned. If the CLOUDNS_SUB_USER_NAME environment variable is not set, an error will be returned.
// config is the ClouDNSConfig to be used for creating the ClouDNSProvider.
// It returns the created ClouDNSProvider and a possible error.
func TestNewClouDNSProvider(t *testing.T) {
	tests := []struct {
		name             string
		userIDType       string
		userID           string
		subUserName      string
		userPassword     string
		expectedError    string
		expectedErrorNil bool
	}{
		{
			name:          "valid user-id login",
			userID:        "12345",
			userPassword:  "password",
			expectedError: "",
		},
		{
			name:             "invalid user-id login",
			userIDType:       "auth-id",
			userID:           "invalid",
			userPassword:     "password",
			expectedError:    "error setting \"AuthID\": strconv.ParseInt: parsing \"invalid\": invalid syntax",
			expectedErrorNil: false,
		},
		{
			name:          "valid sub-user login type",
			userIDType:    "sub-auth-id",
			userID:        "12345",
			userPassword:  "password",
			expectedError: "",
		},
		{
			name:             "invalid login type",
			userIDType:       "invalid",
			userID:           "12345",
			userPassword:     "password",
			expectedError:    "CLOUDNS_AUTH_ID_TYPE is not valid. Expected one of 'auth-id' or 'sub-auth-id' but was: 'invalid'",
			expectedErrorNil: false,
		},
		{
			name:          "missing user password",
			userID:        "12345",
			expectedError: "CLOUDNS_AUTH_PASSWORD environment configuration was missing",
		},
		{
			name:          "missing user id sub-user",
			userPassword:  "password",
			expectedError: "CLOUDNS_AUTH_ID environment configuration was missing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.userID != "" {
				if err := os.Setenv("CLOUDNS_AUTH_ID", test.userID); err != nil {
					log.Fatalf("Couldn't set CLOUDNS_AUTH_ID: %v", err)
				}
			} else {
				if err := os.Unsetenv("CLOUDNS_AUTH_ID"); err != nil {
					log.Fatalf("Couldn't unset CLOUDNS_AUTH_ID: %v", err)
				}
			}
			if test.userIDType != "" {
				if err := os.Setenv("CLOUDNS_AUTH_ID_TYPE", test.userIDType); err != nil {
					log.Fatalf("Couldn't set CLOUDNS_AUTH_ID_TYPE: %v", err)
				}
			} else {
				if err := os.Unsetenv("CLOUDNS_AUTH_ID_TYPE"); err != nil {
					log.Fatalf("Couldn't unset CLOUDNS_AUTH_ID_TYPE: %v", err)
				}
			}
			if test.userPassword != "" {
				if err := os.Setenv("CLOUDNS_AUTH_PASSWORD", test.userPassword); err != nil {
					log.Fatalf("Couldn't set CLOUDNS_AUTH_PASSWORD: %v", err)
				}
			} else {
				if err := os.Unsetenv("CLOUDNS_AUTH_PASSWORD"); err != nil {
					log.Fatalf("Couldn't unset CLOUDNS_AUTH_PASSWORD: %v", err)
				}
			}

			err := makeConfig()
			if err != nil && test.expectedError == "" {
				t.Errorf("got unexpected error: %s", err)
			} else if err == nil && test.expectedError != "" {
				t.Errorf("expected error %q but got nil", test.expectedError)
			} else if err != nil && test.expectedError != "" && err.Error() != test.expectedError {
				t.Errorf("got error %q, want %q", err.Error(), test.expectedError)
			}
			if err == nil && test.expectedErrorNil {
				t.Errorf("expected error but got nil")
			}
		})
	}
}

func makeConfig() error {
	envConfig := &Configuration{}
	if err := env.Set(envConfig); err != nil {
		return err
	}

	config, err := envConfig.ProviderConfig()
	if err != nil {
		return err
	}

	if _, err := NewClouDNSProvider(*config); err != nil {
		return err
	}

	return nil
}

func TestZoneFilter(t *testing.T) {
	zoneOne := mockZones[0]
	zoneTwo := mockZones[1]

	tests := []struct {
		name           string
		domainFilter   *endpoint.DomainFilter
		expectedZones  []cloudns.Zone
		expectingError bool
	}{
		{
			name:           "all zones",
			domainFilter:   endpoint.NewDomainFilterWithExclusions([]string{""}, []string{""}),
			expectedZones:  []cloudns.Zone{zoneOne, zoneTwo},
			expectingError: false,
		},
		{
			name:           "only test1, simple filter",
			domainFilter:   endpoint.NewDomainFilterWithExclusions([]string{"test1.com"}, []string{""}),
			expectedZones:  []cloudns.Zone{zoneOne},
			expectingError: false,
		},
		{
			name:           "only test2, with test1 excluded",
			domainFilter:   endpoint.NewDomainFilterWithExclusions([]string{}, []string{"test1.com"}),
			expectedZones:  []cloudns.Zone{zoneTwo},
			expectingError: false,
		},
		{
			name:           "all zones, with regexp",
			domainFilter:   endpoint.NewRegexDomainFilter(regexp.MustCompile("test[12].com"), regexp.MustCompile(``)),
			expectedZones:  []cloudns.Zone{zoneOne, zoneTwo},
			expectingError: false,
		},
		{
			name:           "only test1, with regexp",
			domainFilter:   endpoint.NewRegexDomainFilter(regexp.MustCompile(`test1\..*`), regexp.MustCompile(``)),
			expectedZones:  []cloudns.Zone{zoneOne},
			expectingError: false,
		},
		{
			name:           "only test2, with exclusion regexp",
			domainFilter:   endpoint.NewRegexDomainFilter(regexp.MustCompile(""), regexp.MustCompile(`test1\..*`)),
			expectedZones:  []cloudns.Zone{zoneTwo},
			expectingError: false,
		},
		{
			name:           "only test2, with two regexp",
			domainFilter:   endpoint.NewRegexDomainFilter(regexp.MustCompile(`.*\.com`), regexp.MustCompile(`test1\..*`)),
			expectedZones:  []cloudns.Zone{zoneTwo},
			expectingError: false,
		},
		{
			name:           "no zones, with non-matching regexp",
			domainFilter:   endpoint.NewRegexDomainFilter(regexp.MustCompile(`.*\.net`), regexp.MustCompile(``)),
			expectedZones:  []cloudns.Zone{},
			expectingError: false,
		},
	}

	oriListZones := listZones
	listZones = func(client *cloudns.Client, ctx context.Context) ([]cloudns.Zone, error) {
		return mockZones, nil
	}

	provider := &ClouDNSProvider{}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			provider.domainFilter = test.domainFilter
			zones, err := provider.Zones(context.Background())

			errExist := err != nil
			if test.expectingError != errExist {
				tt.Errorf("Expected error: %v, got: %v", test.expectingError, errExist)
			}

			if !reflect.DeepEqual(test.expectedZones, zones) {
				tt.Errorf("Error, return value expectation. Want: %+v, got: %+v", test.expectedZones, zones)
			}
		})
	}

	listZones = oriListZones
}

// func Test_Records(t *testing.T)
func TestZoneRecordMap(t *testing.T) {
	zoneOneRecordMap := make(cloudns.RecordMap)
	for _, record := range mockRecords[0] {
		zoneOneRecordMap[record.ID] = record
	}

	oneZoneRecordMap := make(map[string]cloudns.RecordMap)
	oneZoneRecordMap["test1.com"] = zoneOneRecordMap

	zoneTwoRecordMap := make(cloudns.RecordMap)
	for _, record := range mockRecords[1] {
		zoneTwoRecordMap[record.ID] = record
	}

	twoZoneRecordMap := make(map[string]cloudns.RecordMap)
	twoZoneRecordMap["test1.com"] = zoneOneRecordMap
	twoZoneRecordMap["test2.com"] = zoneTwoRecordMap

	tests := []struct {
		name           string
		expectedMap    map[string]cloudns.RecordMap
		expectingError bool
		mockFunc       func()
	}{
		{
			name:           "no zones",
			expectedMap:    map[string]cloudns.RecordMap{},
			expectingError: false,
			mockFunc: func() {
				listZones = func(client *cloudns.Client, ctx context.Context) ([]cloudns.Zone, error) {
					return []cloudns.Zone{}, nil
				}

				listRecords = func(client *cloudns.Client, ctx context.Context, zoneName string) (cloudns.RecordMap, error) {
					return nil, nil
				}
			},
		},
		{
			name:           "no records",
			expectedMap:    map[string]cloudns.RecordMap{},
			expectingError: false,
			mockFunc: func() {
				listZones = func(client *cloudns.Client, ctx context.Context) ([]cloudns.Zone, error) {
					return mockZones, nil
				}

				listRecords = func(client *cloudns.Client, ctx context.Context, zoneName string) (cloudns.RecordMap, error) {
					return nil, nil
				}
			},
		},
		{
			name:           "list zones error",
			expectedMap:    nil,
			expectingError: true,
			mockFunc: func() {
				listZones = func(client *cloudns.Client, ctx context.Context) ([]cloudns.Zone, error) {
					return nil, fmt.Errorf("list zones error")
				}
			},
		},
		{
			name:           "list records error",
			expectedMap:    nil,
			expectingError: true,
			mockFunc: func() {
				listZones = func(client *cloudns.Client, ctx context.Context) ([]cloudns.Zone, error) {
					return mockZones, nil
				}

				listRecords = func(client *cloudns.Client, ctx context.Context, zoneName string) (cloudns.RecordMap, error) {
					return nil, fmt.Errorf("list records error")
				}
			},
		},
		{
			name:           "one zone, five records",
			expectedMap:    oneZoneRecordMap,
			expectingError: false,
			mockFunc: func() {
				listZones = func(client *cloudns.Client, ctx context.Context) ([]cloudns.Zone, error) {
					return mockZones[0:1], nil
				}

				listRecords = func(client *cloudns.Client, ctx context.Context, zoneName string) (cloudns.RecordMap, error) {
					return zoneOneRecordMap, nil
				}
			},
		},
		{
			name:           "two zones, ten records",
			expectedMap:    twoZoneRecordMap,
			expectingError: false,
			mockFunc: func() {
				listZones = func(client *cloudns.Client, ctx context.Context) ([]cloudns.Zone, error) {
					return mockZones, nil
				}

				listRecords = func(client *cloudns.Client, ctx context.Context, zoneName string) (cloudns.RecordMap, error) {
					if zoneName == "test1.com" {
						return zoneOneRecordMap, nil
					}
					if zoneName == "test2.com" {
						return zoneTwoRecordMap, nil
					}
					return nil, nil
				}
			},
		},
	}

	oriListZones := listZones
	oriListRecords := listRecords

	provider := &ClouDNSProvider{}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			test.mockFunc()
			zoneRecordMap, err := provider.zoneRecordMap(context.Background())

			errExist := err != nil
			if test.expectingError != errExist {
				tt.Errorf("Expected error: %v, got: %v", test.expectingError, errExist)
			}

			if !reflect.DeepEqual(test.expectedMap, zoneRecordMap) {
				tt.Errorf("Error, return value expectation. Want: %+v, got: %+v", test.expectedMap, zoneRecordMap)
			}
		})
	}

	listZones = oriListZones
	listRecords = oriListRecords
}
