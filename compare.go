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
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"

	"github.com/go-test/deep"
	"github.com/grafana-tools/sdk"
)

type dashboardDiff struct {
	Source string `json:"source"`
	UID    string `json:"uid"`
	Action string `json:"action"`
	Title  string `json:"title"`
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
		dashboards, err := client.Search(context.TODO())
		if err != nil {
			return err
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
				localDashboard := &sdk.Board{}
				err = json.Unmarshal(data, localDashboard)
				if err != nil {
					return err
				}

				var found bool
				for _, d := range dashboards {
					if d.UID == localDashboard.UID {
						found = true
						break
					}
				}
				if !found {
					fmt.Printf("dashboard %s is new\n", localDashboard.UID)
					output[outputInstance.Name] = append(output[outputInstance.Name], dashboardDiff{
						Action: "new",
						Source: instance.Name,
						UID:    localDashboard.UID,
						Title:  localDashboard.Title,
					})
					return nil
				}

				board, _, err := client.GetDashboardByUID(context.TODO(), localDashboard.UID)
				if err != nil {
					return err
				}

				if !equalDashboards(*localDashboard, board) {
					fmt.Printf("dashboard %s (%s) is different\n", board.UID, board.Title)
					output[outputInstance.Name] = append(output[outputInstance.Name], dashboardDiff{
						Action: "modify",
						Source: instance.Name,
						UID:    localDashboard.UID,
						Title:  localDashboard.Title,
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

func equalDashboards(a, b sdk.Board) bool {
	reset := func(i sdk.Board) sdk.Board {
		i.ID = 0
		return i
	}
	dashboardA := reset(a)
	dashboardB := reset(b)
	if diff := deep.Equal(dashboardA, dashboardB); diff != nil {
		//fmt.Printf("diff: %v", diff)
		return false
	}
	return true
}
