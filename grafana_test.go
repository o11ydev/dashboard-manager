package main

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChangeDataSource(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/1/local/zgbO0T-7k.json")
	require.NoError(t, err)
	localDashboard := &FullDashboard{}
	err = json.Unmarshal(data, localDashboard)
	require.NoError(t, err)

	final, err := ioutil.ReadFile("testdata/1/final/zgbO0T-7k.json")
	require.NoError(t, err)
	expectedDashboard := &FullDashboard{}
	err = json.Unmarshal(final, expectedDashboard)
	require.NoError(t, err)

	changeDatasources(localDashboard.Dashboard, localDashboard.Datasources, expectedDashboard.Datasources)
	require.Equal(t, localDashboard.Dashboard, expectedDashboard.Dashboard)
}
