// Copyright 2021 Inuits
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"errors"
	"io/ioutil"
	"strings"

	gsdk "github.com/grafana/grafana-api-golang-client"
	promcfg "github.com/prometheus/common/config"
)

type grafanaInstance struct {
	Name            string                   `yaml:"name"`
	URL             string                   `yaml:"url"`
	Auth            string                   `yaml:"api_key"`
	AuthFile        string                   `yaml:"api_key_file"`
	IncludeTags     []string                 `yaml:"include_tags"`
	PurgeDashboards bool                     `yaml:"purge_dashboards"`
	HttpClient      promcfg.HTTPClientConfig `yaml:"http_client"`
}

func (g *grafanaInstance) client() (*gsdk.Client, error) {
	auth := g.Auth
	if g.AuthFile != "" {
		fileContent, err := ioutil.ReadFile(g.AuthFile)
		if err != nil {
			return nil, err
		}
		auth = strings.TrimSpace(string(fileContent))
	}
	client, err := promcfg.NewClientFromConfig(g.HttpClient, "grafana")
	if err != nil {
		return nil, err
	}
	return gsdk.New(g.URL, gsdk.Config{
		APIKey: auth,
		Client: client,
	})
}

func (g *grafanaInstance) shouldIncludeDashboard(b *gsdk.Dashboard) bool {
	if len(g.IncludeTags) == 0 {
		return true
	}
	for _, t := range b.Model["tags"].([]string) {
		lt := strings.ToLower(t)
		for _, i := range g.IncludeTags {
			if strings.ToLower(i) == lt {
				return true
			}
		}
	}
	return false
}

func getTags(b *gsdk.Dashboard) []string {
	if tags, ok := b.Model["tags"].([]string); ok {
		return tags
	}
	return nil
}

func getUID(b *gsdk.Dashboard) (string, error) {
	if uid, ok := b.Model["uid"].(string); ok {
		return uid, nil
	}
	return "", errors.New("No UID for dashboard")
}

func getTitle(b *gsdk.Dashboard) (string, error) {
	if uid, ok := b.Model["title"].(string); ok {
		return uid, nil
	}
	return "", errors.New("No title for dashboard")
}
