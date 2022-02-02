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
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"

	gsdk "github.com/grafana/grafana-api-golang-client"
)

func snapshotDashboards(cfg *config) error {
	var inputInstance grafanaInstance
	var outputInstance grafanaInstance
	var found bool
	for _, o := range cfg.Output {
		if o.Name == *snapshotOutput {
			outputInstance = o
			found = true
			break
		}
	}
	if !found {
		return errors.New("output instance not found")
	}

	found = false
	for _, i := range cfg.Input {
		if i.Name == *snapshotSource {
			inputInstance = i
			found = true
			break
		}
	}

	if !found {
		return errors.New("input instance not found")
	}

	client, err := outputInstance.client()
	if err != nil {
		return err
	}

	basepath := filepath.Join(*snapshotDirectory, inputInstance.Name)
	dashboards := []*FullDashboard{}
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
		dashboards = append(dashboards, localDashboard)
		return nil
	})
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

	for _, dashboardUID := range *snapshotDashboardsList {
		var dashboard FullDashboard
		var found bool
		for _, d := range dashboards {
			uid, err := getUID(d.Dashboard)
			if err != nil {
				return err
			}
			if uid == dashboardUID {
				dashboard = *d
				found = true
			}
		}
		if !found {
			return fmt.Errorf("dasboard %s not found", dashboardUID)
		}

		changeDatasources(dashboard.Dashboard, dashboard.Datasources, clientDS)

		resp, err := client.NewSnapshot(gsdk.Snapshot{
			Expires: int64(snapshotExpire.Seconds()),
			Model:   dashboard.Dashboard.Model,
		})

		if err != nil {
			return err
		}
		fmt.Printf("%s\n", resp.URL)
	}
	return nil
}
