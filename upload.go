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
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/grafana-tools/sdk"
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
			if d.Dashboard.UID == dashboardUID {
				dashboard = *d
				found = true
			}
		}
		if !found {
			return fmt.Errorf("dasboard %s not found", dashboardUID)
		}

		params := sdk.SetDashboardParams{} //Overwrite: true}
		if dashboard.Properties.FolderID == 0 {
			params.FolderID = 0
		} else {
			folders, err := client.GetAllFolders(context.TODO())
			if err != nil {
				return err
			}
			var found bool
			for _, f := range folders {
				if f.Title == dashboard.Properties.FolderTitle {
					params.FolderID = f.ID
					found = true
					break
				}
			}
			if !found {
				//TODO create folder
				params.FolderID = 0
			}
		}

		if _, err = client.DeleteDashboard(context.TODO(), dashboard.Dashboard.UpdateSlug()); err != nil {
			log.Println(err)
			continue
		}
		_, err = client.SetDashboard(context.TODO(), dashboard.Dashboard, params)
		if err != nil {
			return fmt.Errorf("error uploading %s: %v", dashboardUID, err)
		}
	}
	return nil
}
