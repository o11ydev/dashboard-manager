# Dashboard Manager

A command line utility to manage and promote dashboards between Grafana
instances, coupled with gitlab-ci.

## Example configuration

```
grafana_instances_output:
  - api_key_file: production-secret
    url: http://127.0.0.1:3000
    name: prod
grafana_instances_input:
  - api_key_file: dev-secret
    url: https://remote-dev.example.com/
    name: dev
    http_client:
        # Configured like alertmanager http client:
        # https://prometheus.io/docs/alerting/latest/configuration/#http_config
        tls_config:
            insecure_skip_verify: true
```

## Usage


```
Flags:
      --help                     Show context-sensitive help (also try
                                 --help-long and --help-man).
  -c, --config-file=CONFIG-FILE  Path to the configuration file.

Commands:
  help [<command>...]
    Show help.

  fetch --output-directory=OUTPUT-DIRECTORY
    Fetch dashboards from input grafana.

  compare --dashboards-directory=DASHBOARDS-DIRECTORY --results=RESULTS
    Compare dashboards.

  upload --dashboards-directory=DASHBOARDS-DIRECTORY --input-instance=INPUT-INSTANCE --output-instance=OUTPUT-INSTANCE --dashboards=DASHBOARDS
    Upload dashboards.

  snapshot --dashboards-directory=DASHBOARDS-DIRECTORY --input-instance=INPUT-INSTANCE --output-instance=OUTPUT-INSTANCE --dashboards=DASHBOARDS [<flags>]
    Upload snapshots.
```
