# es-archivist
Shipping logs or other data to Elasticsearch creates the never ending battle of removing your oldest indices and/or cleaning up unwanted data to keep from running out of space in your cluster. This tool is designed to automate this process in a smart and configurable way allowing you to spend your time doing other things like manually griding you coffee beans or learning a new programming language.

## Build
```
$ go build -o esa .
```

## Running
This utility is designed to run continuously as a daemon and monitor disk usage.

### Manually Run
To run manually, simple execute the binary from the command line.
```
$ esa
```

### SystemD Service
Simply create a file in `/etc/systemd/system/` called `esa.service` then run `systemctl daemon-reload` to reload unit files and `systemctl start esa.service` to start the service.

```
[Unit]
Description=ES Archivist Service
[Service]
User=storj
Group=storj
Restart=always
KillMode=control-group
ExecStart=/usr/local/bin/esa
```

## Configuration
Esa is configured by values in a JSON file named `config.json` located either in your current directory, or in the default directory which is `/etc/esa/config.json`.

### Config Values
| Key | Required | Default | Description |
| --- | -------- | ------- | ----------- |
| ESHost | NO | `localhost:9200` | Hostname and port of your ElasticSearch instance |
| MinStorageBytes | NO | 109951162777 | Minimum bytes available before archiving indices (not currently used) |
| SleepSeconds | NO | 5 | Number of seconds to sleep between storage available checks if all is going as expected |
| SleepAfterDeleteIndex | NO | 60 | Number of seconds to sleep after deleting an index |
| SleepAfterMainLoopSeconds | NO | 60 | Number of seconds to sleep if a problem is encountered and we restart the watcher |
| MinFreeSpacePercent | NO | 22 | Percentage of space to ensure is kept free |
| SnapshotRepositoryName | YES | N/A | Name of your snapshot repository |
| MinIndexCount | NO | 30 | Minimum number of indices that should exist. Processing will stop when the index count reaches this number |
| SnapDryRun | NO | true | With this set to true, no snapshots will actually be taken, you will only be notified that it would have happened |
| IndexDryRun | NO | true | With this set to true, no indices will be deleted, you will only be notified that it would have been deleted |
| IndexIncludePrefix | NO | [] | This is an array of strings which should include a prefix of indices that you wish to prune (I.E. Setting this value to ["logstash"] would prune indices like "logstash-2017-04-27" but not "kibana-index") |

### Example Configuration
```json
{
  "ESHost": "localhost:9200",
  "MinStorageBytes": 109951162777,
  "SleepSeconds": 5,
  "SleepAfterDeleteIndex": 60,
  "MinFreeSpacePercent": 22,
  "SnapshotRepositoryName": "snaps",
  "MinIndexCount": 5,
  "SnapDryRun": false,
  "IndexDryRun": false,
  "IndexIncludePrefix": [
    "logstash"
  ]
}
```
