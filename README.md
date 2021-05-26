prometheus-mgit-exporter
========================

A collection of useful monitoring for [Prometheus][1] by mgIT GmbH.

[1]: https://prometheus.io/

Available Checks
----------------

Each check must be enabled in the configuration file individually. The following
checks are currently available:

### **certfile** ###

Walks certain directories (like /etc/ssl, globbing configureable) and exports the "not after" field of all certificates found.

**Parameters**:

- enable (bool) activate or deactivate this check
- globs ([]string) directory globbing patterns
- exclude_system (bool) exclude certs in /etc/ssl/certs/ if set to true

### **mcelog** ###

Reports the size of the machine exception log indicating hardware errors.

**Parameters:**

- enable (bool) activate or deactivate this check
- path (string) path to the mcelog 
  <br>Default: "/var/log/mcelog"

### **ptheartbeat** ###

Reports the replication lag via the pt-heartbeat command and indicates whether it has been collected successfully or not, by
executing the pt-heartbeat command with the following flags:
<br>--check, --database, --table, --defaults-file, --master-server-id, <br>--noinsert-heartbeat-row, --utc.

**Parameters:**
  - enable (bool) activate or deactivate this check
  - database (string) database parameter in the pt-heartbeat command
    <br>Default: "system"
  - table (string) table parameter in the pt-heartbeat command
    <br>Default: "pt_heartbeat"
  - defaultsFile (string) defaults-file parameter in the pt-heartbeat command
    <br>Default: "/etc/mysql/debian.cnf"
  - masterId (int) master-server-id in the pt-heartbeat command

### **fstab** ###

Reports for network file systems (nfs) whether the nofail flag is used, so the nfs does not hinder the booting process. Moreover it reports wheter the fstab information was collected successfully.

**Parameters:**
  - enable (bool) activate or deactivate this check

### **binlog** ###

Reports if all binlog files from an index (mysql-bin.index) are available.

**Parameters:**
  - enable (bool) activate or deactivate this check
  - path (string) path to the index directory
    <br>Default: "/var/log/mysql"

### **rasdaemon** ###

Reports the size of the mc-event log. 

**Parameters:**
  - enable (bool) activate or deactivate this check
  - path (string) path to the mc-event log
    <br>Default: "/var/lib/rasdaemon/ras-mc_event.db"

### **elk** ###

Reports the number of elk (elasticsearch) indices that are on a node longer than a given time duration. Usually these are indices that should have been moved from hot to cold storage but haven't.

**Parameters:**
  - enable (bool) activate or deactivate this check
  - duration (string) the duration after which the indices should have been moved
    <br>Default: "170h"
  - node (string) the name of the node
    <br>Default: "hot"

### **exec** ###

Runs your own executables (e.g collecting additional information, restarting a service etc) via the exec URL endpoint. You can't run the same executable in parallel multiple times. Configure the executables similarly to the example configuration. The script is executed when you visit the endpoint /exec/\<id>.

**Parameters:**

  - enable (bool) activate or deactivate this check
  - scripts (map[string]CmdOptions) configure your scripts with an id and options
  - id (string) a unique identifier for your script as string
    - command ([]string) command or path executable with arguments
    - dir (string) directory where the executable will be run
      <br>Default: working directory
    - timeout (string) timeout for your scripts
      <br>Default: "5s"

Configuration
-------------

This is an example configuration file with all checks enabled:

    {
      "listen": ":9328",
      "certfile": {
        "enable": true,
        "globs": [
          "/etc/ssl/**/*.pem",
          "/etc/ssl/**/*.crt"
        ],
        "exclude_system": true
      },
      "mcelog": {
        "enable": true,
        "path": "/var/log/mcelog"
      },
      "ptheartbeat": {
        "enable": true,
        "database": "system",
        "table": "pt_heartbeat",
        "defaultsFile": "/etc/mysql/debian.cnf",
        "masterId": 0
      },
      "fstab": {
        "enable": true,
      },
      "binlog": {
        "enable": true,
        "path": "var/log/mysql"
      },
      "rasdaemon": {
        "enable": true,
        "path": "/var/lib/rasdaemon/ras-mc_event.db"
      },
      "elk": {
        "enable": true,
        "duration": "170h",
        "node": "hot",
      }
      "exec": {
        "enable": true,
        "scripts": {
          "performance": {
            "command": ["./opt/check_performance.sh"],
            "timeout": "5s"
          }
        }
      }
    }

License
-------

prometheus-mgit-exporter is distributed under the Apache License.
See LICENSE for details.

> Copyright 2021 mgIT GmbH.
>
> Licensed under the Apache License, Version 2.0 (the "License");
> you may not use this file except in compliance with the License.
> You may obtain a copy of the License at
>
>     http://www.apache.org/licenses/LICENSE-2.0
>
> Unless required by applicable law or agreed to in writing, > software
> distributed under the License is distributed on an "AS IS" BASIS,
> WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
> See the License for the specific language governing permissions and
> limitations under the License.
> 