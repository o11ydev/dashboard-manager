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
	"io/ioutil"
	"strings"

	"github.com/roidelapluie/sdk"
)

type grafanaInstance struct {
	Name            string   `yaml:"name"`
	URL             string   `yaml:"url"`
	Auth            string   `yaml:"api_key"`
	AuthFile        string   `yaml:"api_key_file"`
	IncludeTags     []string `yaml:"include_tags"`
	PurgeDashboards bool     `yaml:"purge_dashboards"`
}

func (g *grafanaInstance) client() (*sdk.Client, error) {
	auth := g.Auth
	if g.AuthFile != "" {
		fileContent, err := ioutil.ReadFile(g.AuthFile)
		if err != nil {
			return nil, err
		}
		auth = strings.TrimSpace(string(fileContent))
	}
	return sdk.NewClient(g.URL, auth, sdk.DefaultHTTPClient)
}

func (g *grafanaInstance) shouldIncludeDashboard(b sdk.Board) bool {
	if len(g.IncludeTags) == 0 {
		return true
	}
	for _, t := range b.Tags {
		lt := strings.ToLower(t)
		for _, i := range g.IncludeTags {
			if strings.ToLower(i) == lt {
				return true
			}
		}
	}
	return false
}
