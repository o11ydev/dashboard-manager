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

func uploadDashboards(cfg *config) error {
	var inputInstance grafanaInstance
	var outputInstance grafanaInstance
	var found bool
	for _, o := range cfg.Output {
		if o.Name == *uploadOutput {
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
		if i.Name == *uploadSource {
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

	clientDS := []*gsdk.DataSource{}
	// Hard code limit to 50 for now.
	for i := int64(0); i < 50; i++ {
		ds, err := client.DataSource(i)
		if err == nil {
			clientDS = append(clientDS, ds)
		}
	}

	basepath := filepath.Join(*uploadDirectory, inputInstance.Name)
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
	for _, dashboardUID := range *uploadDashboardsList {
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
			return fmt.Errorf("dashboard %s not found", dashboardUID)
		}

		if dashboard.Folder != nil && dashboard.Folder.ID != 0 {
			folders, err := client.Folders()
			if err != nil {
				return err
			}
			var found bool
			for _, f := range folders {
				if f.Title == dashboard.Folder.Title {
					dashboard.Dashboard.Meta.Folder = f.ID
					dashboard.Dashboard.Folder = f.ID
					found = true
					break
				}
			}
			if !found {
				folder, err := client.NewFolder(dashboard.Folder.Title)
				if err != nil {
					return err
				}
				dashboard.Dashboard.Meta.Folder = folder.ID
				dashboard.Dashboard.Folder = folder.ID
				dashboard.Dashboard.Meta.Slug = ""
			}
		}

		dashboard.Dashboard.Model["id"] = 0
		dashboard.Dashboard.Overwrite = true

		changeDatasources(dashboard.Dashboard, dashboard.Datasources, clientDS)

		_, err = client.NewDashboard(*dashboard.Dashboard)
		if err != nil {
			return fmt.Errorf("error uploading %s: %v", dashboardUID, err)
		}
	}
	return nil
}
