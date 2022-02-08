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
	"io/ioutil"
	"os"
	"path/filepath"

	gapi "github.com/grafana/grafana-api-golang-client"
)

func lazyMkdir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.Mkdir(path, os.ModePerm)
	}
	return err
}

type FullDashboard struct {
	Dashboard   *gapi.Dashboard `json:"board"`
	Datasources []*gapi.DataSource
	Folder      *gapi.Folder
}

func fetchDashboards(cfg *config) error {
	err := lazyMkdir(*fetchDirectory)
	if err != nil {
		return fmt.Errorf("error making base directory: %w", err)
	}

	for _, instance := range cfg.Input {
		client, err := instance.client()
		if err != nil {
			return err
		}
		basepath := filepath.Join(*fetchDirectory, instance.Name)
		err = lazyMkdir(basepath)
		if err != nil {
			return fmt.Errorf("error making directory for %s: %w", instance.Name, err)
		}

		dashboards, err := client.Dashboards()
		if err != nil {
			return err
		}

		clientDS := []*gapi.DataSource{}
		// Hard code limit to 50 for now.
		for i := int64(0); i < 50; i++ {
			ds, err := client.DataSource(i)
			if err == nil {
				clientDS = append(clientDS, ds)
			}
		}

		for _, d := range dashboards {
			board, err := client.DashboardByUID(d.UID)
			if err != nil {
				return fmt.Errorf("error fetching %s: %w", d.UID, err)
			}

			if !instance.shouldIncludeDashboard(board) {
				continue
			}

			folder, err := client.Folder(board.Meta.Folder)
			if err != nil {
				return fmt.Errorf("error fetching folder %d: %w", board.Meta.Folder, err)
			}

			dashboardDS := []*gapi.DataSource{}
			datasources := getDatasources(board)
			for _, ds := range clientDS {
				for _, v := range datasources {
					if ds.UID == v {
						dashboardDS = append(dashboardDS, &gapi.DataSource{
							UID:  ds.UID,
							Type: ds.Type,
							Name: ds.Name,
						})
					}
				}
			}

			data, err := json.MarshalIndent(FullDashboard{
				Dashboard:   board,
				Folder:      folder,
				Datasources: dashboardDS,
			}, "", " ")
			if err != nil {
				return err
			}

			folderPath := filepath.Join(basepath, folder.UID)
			err = lazyMkdir(folderPath)
			if err != nil {
				return fmt.Errorf("error making directory for %s / %s: %w", instance.Name, d.FolderUID, err)
			}

			filePath := filepath.Join(folderPath, d.UID+".json")
			err = ioutil.WriteFile(filePath, data, 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
