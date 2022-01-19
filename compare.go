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
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/go-test/deep"
	"github.com/google/go-cmp/cmp"
	gsdk "github.com/grafana/grafana-api-golang-client"
)

type dashboardDiff struct {
	Source string   `json:"source"`
	UID    string   `json:"uid"`
	Action string   `json:"action"`
	Title  string   `json:"title"`
	Tags   []string `json:"tags"`
	Diff   string   `json:"diff"`
}

type diff map[string][]dashboardDiff

func compareDashboards(cfg *config) error {
	output := make(diff, 0)
	for _, outputInstance := range cfg.Output {
		output[outputInstance.Name] = []dashboardDiff{}
		client, err := outputInstance.client()
		if err != nil {
			return err
		}
		dashboards, err := client.Dashboards()
		if err != nil {
			return err
		}

		clientDS := []*gsdk.DataSource{}
		// Hard code limit to 50 for now.
		for i := int64(0); i < 50; i++ {
			ds, err := client.DataSource(i)
			if err == nil {
				clientDS = append(clientDS, ds)
			}
		}

		for _, instance := range cfg.Input {
			basepath := filepath.Join(*compareDirectory, instance.Name)
			err = filepath.Walk(basepath, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				localDashboard := &FullDashboard{}
				err = json.Unmarshal(data, localDashboard)
				if err != nil {
					return err
				}

				changeDatasources(localDashboard.Dashboard, localDashboard.Datasources, clientDS)

				tags := sanitizeTags(getTags(localDashboard.Dashboard))

				if !outputInstance.shouldIncludeDashboard(localDashboard.Dashboard) {
					return nil
				}

				var found bool
				for _, d := range dashboards {
					uid, err := getUID(localDashboard.Dashboard)
					if err != nil {
						return err
					}
					if d.UID == uid {
						found = true
						break
					}
				}
				if !found {
					title, err := getTitle(localDashboard.Dashboard)
					if err != nil {
						return err
					}
					uid, err := getUID(localDashboard.Dashboard)
					if err != nil {
						return err
					}

					fmt.Printf("Dashboard %s (%s) is new.\n", title, uid)
					output[outputInstance.Name] = append(output[outputInstance.Name], dashboardDiff{
						Action: "new",
						Source: instance.Name,
						UID:    uid,
						Title:  title,
						Tags:   tags,
					})
					return nil
				}

				uid, err := getUID(localDashboard.Dashboard)
				if err != nil {
					return err
				}
				title, err := getTitle(localDashboard.Dashboard)
				if err != nil {
					return err
				}
				board, err := client.DashboardByUID(uid)
				if err != nil {
					return err
				}
				folder, err := client.Folder(board.Meta.Folder)
				if err != nil {
					return err
				}

				outputDashboard := FullDashboard{Dashboard: board, Folder: folder}
				if !equalDashboards(*localDashboard, outputDashboard) {
					fmt.Printf("Dashboard %s (%s) is different.\n", title, uid)
					output[outputInstance.Name] = append(output[outputInstance.Name], dashboardDiff{
						Action: "modify",
						Source: instance.Name,
						UID:    uid,
						Title:  title,
						Tags:   tags,
						Diff:   cmp.Diff(*localDashboard, outputDashboard),
					})
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("error comparing dashboards: %w", err)
			}
		}
	}
	data, err := json.MarshalIndent(output, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(*compareResults, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func equalDashboards(a, b FullDashboard) bool {
	reset := func(i gsdk.Dashboard) gsdk.Dashboard {
		i.Model["id"] = 0
		i.Model["slug"] = ""
		i.Model["version"] = 1
		return i
	}
	dashboardA := reset(*a.Dashboard)
	dashboardB := reset(*b.Dashboard)
	if diff := deep.Equal(dashboardA.Model, dashboardB.Model); diff != nil {
		return false
	}
	if a.Folder.Title != b.Folder.Title {
		return false
	}
	return true
}

func sanitizeTags(tags []string) []string {
	sanitizedTags := make([]string, len(tags))
	for i, s := range tags {
		sanitizedTags[i] = strings.TrimSpace(strings.ToLower(s))
	}
	return sanitizedTags
}
