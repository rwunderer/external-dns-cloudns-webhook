/*
 * Configuration - provider configuration
 *
 * Copyright 2023 Marco Confalonieri.
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
	"fmt"
	"regexp"
	"strings"

	cloudns "github.com/ppmathis/cloudns-go"

	"github.com/codingconcepts/env"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
)

// Configuration contains the ClouDNS provider's configuration.
type Configuration struct {
	AuthIDType           string   `env:"CLOUDNS_AUTH_ID_TYPE" default:"auth-id"`
	AuthID               int      `env:"CLOUDNS_AUTH_ID" required:"true"`
	AuthPassword         string   `env:"CLOUDNS_AUTH_PASSWORD" required:"true"`
	DryRun               bool     `env:"DRY_RUN" default:"false"`
	Debug                bool     `env:"CLOUDNS_DEBUG" default:"false"`
	BatchSize            int      `env:"BATCH_SIZE" default:"100"`
	DefaultTTL           int      `env:"DEFAULT_TTL" default:"3600"`
	OwnerID              string   `env:"OWNER_ID" default:""`
	DomainFilter         []string `env:"DOMAIN_FILTER" default:""`
	ExcludeDomains       []string `env:"EXCLUDE_DOMAIN_FILTER" default:""`
	RegexDomainFilter    string   `env:"REGEXP_DOMAIN_FILTER" default:""`
	RegexDomainExclusion string   `env:"REGEXP_DOMAIN_FILTER_EXCLUSION" default:""`
}

func NewConfiguration() (*Configuration, error) {
	cfg := &Configuration{}

	// Populate with values from environment.
	if err := env.Set(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetDomainFilter returns the domain filter from the configuration.
func GetDomainFilter(config Configuration) endpoint.DomainFilter {
	var domainFilter endpoint.DomainFilter
	createMsg := "Creating ClouDNS provider with "

	if config.RegexDomainFilter != "" {
		createMsg += fmt.Sprintf("Regexp domain filter: '%s', ", config.RegexDomainFilter)
		if config.RegexDomainExclusion != "" {
			createMsg += fmt.Sprintf("with exclusion: '%s', ", config.RegexDomainExclusion)
		}
		domainFilter = endpoint.NewRegexDomainFilter(
			regexp.MustCompile(config.RegexDomainFilter),
			regexp.MustCompile(config.RegexDomainExclusion),
		)
	} else {
		if len(config.DomainFilter) > 0 {
			createMsg += fmt.Sprintf("zoneNode filter: '%s', ", strings.Join(config.DomainFilter, ","))
		}
		if len(config.ExcludeDomains) > 0 {
			createMsg += fmt.Sprintf("Exclude domain filter: '%s', ", strings.Join(config.ExcludeDomains, ","))
		}
		domainFilter = endpoint.NewDomainFilterWithExclusions(config.DomainFilter, config.ExcludeDomains)
	}

	createMsg = strings.TrimSuffix(createMsg, ", ")
	if strings.HasSuffix(createMsg, "with ") {
		createMsg += "no kind of domain filters"
	}
	log.Info(createMsg)
	return domainFilter
}

// GetAuth returns an options object for authentication
func GetAuth(config Configuration) (cloudns.Option, error) {
	var auth cloudns.Option

	if config.AuthIDType == "auth-id" {
		auth = cloudns.AuthUserID(config.AuthID, config.AuthPassword)
	} else if config.AuthIDType == "sub-auth-id" {
		auth = cloudns.AuthSubUserID(config.AuthID, config.AuthPassword)
	} else {
		return nil, fmt.Errorf("CLOUDNS_AUTH_ID_TYPE is not valid. Expected one of 'auth-id' or 'sub-auth-id' but was: '%s'", config.AuthIDType)
	}

	return auth, nil
}

// ProviderConfig returns the configuration as expected by the provider
func (c *Configuration) ProviderConfig() (*ClouDNSConfig, error) {
	auth, err := GetAuth(*c)
	if err != nil {
		return nil, err
	}

	return &ClouDNSConfig{
		Auth:         auth,
		DomainFilter: GetDomainFilter(*c),
		ZoneIDFilter: provider.NewZoneIDFilter([]string{}),
		DefaultTTL:   c.DefaultTTL,
		OwnerID:      c.OwnerID,
		DryRun:       c.DryRun,
	}, nil
}
