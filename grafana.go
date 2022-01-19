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
	"reflect"
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

func extractDS(v reflect.Value) []string {
	var output []string
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			output = append(output, extractDS(v.Index(i))...)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			innerVal := v.MapIndex(k)
			output = append(output, extractDS(innerVal)...)
			if k.String() == "datasource" && innerVal.Kind() == reflect.Interface {
				innerInt := innerVal.Interface()
				if v, ok := innerInt.(map[string]interface{}); ok {
					if uid, ok := v["uid"]; ok {
						uidstr := uid.(string)
						if !strings.HasPrefix(uidstr, "-- ") {
							output = append(output, uidstr)
						}
					}
				}
			}
		}
	default:
	}

	keys := make(map[string]bool)
	var list []string
	for _, item := range output {
		if _, value := keys[item]; !value {
			list = append(list, item)
			keys[item] = true
		}
	}
	return list
}

func changeDS(v reflect.Value, equiv map[string]string) {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			changeDS(v.Index(i), equiv)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			innerVal := v.MapIndex(k)
			changeDS(innerVal, equiv)
			if k.String() == "datasource" && innerVal.Kind() == reflect.Interface {
				innerInt := innerVal.Interface()
				if v, ok := innerInt.(map[string]interface{}); ok {
					if uid, ok := v["uid"]; ok {
						uidstr := uid.(string)
						if !strings.HasPrefix(uidstr, "-- ") {
							if newUID, ok := equiv[uidstr]; ok {
								v["uid"] = newUID
							}
						}
					}
				}
			}
		}
	default:
	}
}

func getDatasources(b *gsdk.Dashboard) []string {
	return extractDS(reflect.ValueOf(b.Model))
}

func changeDatasources(b *gsdk.Dashboard, in, out []*gsdk.DataSource) {
	equiv := make(map[string]string)
	for _, inv := range in {
		for _, outv := range out {
			if inv.Name == outv.Name && inv.Type == outv.Type {
				equiv[inv.UID] = outv.UID
			}
		}
	}

	changeDS(reflect.ValueOf(b.Model), equiv)
}
