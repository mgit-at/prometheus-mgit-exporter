prometheus-mgit-exporter
========================

A collection of useful monitoring for [Prometheus][1] by mgIT GmbH.

[1]: https://prometheus.io/

Available Checks
----------------

Each check must be enabled in the configuration file individually. The following
checks are currently available:

**certfile:** walks certain directories (like /etc/ssl, globbing configureable)
and exports the "not after" field of all certificates found.

**mcelog:** reports the size of the machine exception log (/var/log/mcelog)
indicating hardware errors.

Configuration
-------------

This is an example configuration file with all checks enabled:

    {
      "listen": ":9328",
      "certfile": {
        "enable": false,
        "globs": [
          "/etc/ssl/**/*.pem",
          "/etc/ssl/**/*.crt"
        ],
        "exclude_system": true
      },
      "mcelog": {
        "enable": false,
        "path": "/var/log/mcelog"
      },
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
