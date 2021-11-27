[![codecov](https://codecov.io/gh/orolia/prisma/branch/develop/graph/badge.svg?token=isXJvglRlX)](https://codecov.io/gh/orolia/prisma)
[![AWS Codebuild](https://codebuild.us-east-1.amazonaws.com/badges?uuid=eyJlbmNyeXB0ZWREYXRhIjoiZTViQWpLb2Zib1R2R2JoVVhzd0VmUW1LekpNUEhYNlNjSjBpRlJPTHVpYjhZcDFLVnZ1UkkvRWFIUXoxc1NJd0s2bHF0UExBMUdaUC85NkhVeTlxeXZrPSIsIml2UGFyYW1ldGVyU3BlYyI6InJKelFVYllMZTJsNll1Mm4iLCJtYXRlcmlhbFNldFNlcmlhbCI6MX0%3D&branch=develop)](https://console.aws.amazon.com/codesuite/codebuild/projects/prisma/history?region=us-east-1)

# Prisma C2 Platform

## Summary

The Prisma repository contains one main projects:

* TMS: Service daemons used for the collection and distribution of sensor data.

Documentation is located in the `docs` directory. Run `make serve-docs` to serve the docs locally at `127.0.0.1:8000`.

You may need to run the following two lines to get the docs building:

```bash tab="macOS"
brew install mkdocs
pip3 install mkdocs-material pymdown-extensions
```

```bash tab="Ubuntu"
sudo apt install mkdocs
pip install mkdocs-material pymdown-extensions
```

## TMS

### Development Setup

The easiest way to get started with TMS development is to use Vagrant to setup
a virtual machine. Install the following:

* [VirtualBox](https://www.virtualbox.org/wiki/VirtualBox)
* [Vagrant](https://www.vagrantup.com/)

Create the virtual machine by executing the following command from the
repository root:

```bash
vagrant up
```
For running one more machine use the next snippet:
```bash
vagrant up cc2
```

### Building

You can now edit the source files on your local system with your favorite
editor and then compile in the virtual machine with:

```bash
vagrant ssh
cd ~/go/src/prisma
make
```
For another machine:
```bash
vagrant ssh cc2
cd ~/go/src/prisma
make
```

Once complete, the entire installation will available under `bin`.

TMS executables should be available in the standard path. Symlinks are created
to the artifacts built in the `bin` directory.

When changing the protobuf files, use `make protobuf` to regenerate the
related source code. The default Makefile target does not perform this step.

### Makefile targets

| Target      | Action                                             |
|-------------|----------------------------------------------------|
| all         | Compiles the server code and documentation
| init        | Prepare the project to be run
| docs        | Generate static documentation site
| serve-docs  | Generate documentation site server on 127.0.0.1
| version     | Show current version of project
| protobuf    | Generate source code from protobuf definitions
| format      | Automatically format the source code
| test        | Run the test suite
| testbed     | Create deb package
| vet         | Check for programming or style errors
| dist        | Create a debian package using the existing build binaries
| clean       | Remove all generated files
| integration | Run the integration test suite
| acceptance  | Is the same as integration
| demo        | Run the demo data load script (tmsd must be running)
| cover       | Run the test suit and coverage outputs .cover.html file
| lint        | Run linter with errcheck, unused, deadcode, and gosec enabled. output stdout.


#### macOS Development Setup
Running on native macOS

Pros
*  Faster compile
*  Debugger

Cons
*  Complex setup
*  Greater configuration variance

Compile
```
brew install ejdb
```

**MongoDB**

Native
```
brew install mongodb
mkdir -p ~/data/db
mongod --dbpath ~/data/db --port 8201 --replSet replocal
mongo localhost:8201 etc/mongodb/replication.js
mongo localhost:8201/trident etc/mongodb/schema/trident.js
mongo localhost:8201/aaa etc/mongodb/schema/aaa.js
```

Docker (not working, replicaset port)
```
docker run --name prisma-mongo -d -p 127.0.0.1:8201:8201 mongo --replSet replocal
mongo localhost:8201 etc/mongodb/replication.js
mongo localhost:8201/trident etc/mongodb/schema/trident.js
mongo localhost:8201/aaa etc/mongodb/schema/aaa.js
```

Consul
```
brew install consul
consul agent -dev -bind 127.0.0.1 -client 127.0.0.1
```

Redis
```
brew install redis
redis-server /usr/local/etc/redis.conf
```

## Running

Use [tmsd](tms/cmd/daemons/tmsd/README.md) to manage the TMS processes:

| Command            | Use                              |
|--------------------|----------------------------------|
| `tmsd --start &`   | Start all processes              |
| `tmsd --info`      | Verify the processes are running |
| `tmsd --stop`      | Stop all processes               |
| `tmsd --help`      | For additional information       |

### REST Documentation

#### Postman (Preferred)

We currently provide two types of REST API documentation. For the most up to date
and best documentation experience, run POSTMAN app or see the POSTMAN hosted documentation at [Orolia POSTMAN C2 Documentation Site.](https://orolia-prisma-c2.postman.co/collections/352230-e1e86aaf-b2c3-4e3b-b718-a74f4b7a0281?workspace=44d15d8d-0a82-4c95-8e74-d196d18dd868)

#### Swagger

Additionally we also provide swagger documentation.

To view the JSON for the RESTful API:
http://localhost:8080/api/v2/apidocs.json

To view with the swagger UI:

1. [Install latest docker]( https://store.docker.com/editions/community/docker-ce-desktop-mac)
2. run
  ```docker pull swaggerapi/swagger-ui;
  docker run -p 80:8080 swaggerapi/swagger-ui
  ```
3. Goto - http://localhost/?url=https://localhost:8080/api/v2/apidocs.json

### Configuration

Processes managed by tmsd are configured in `/etc/trident/tmsd.conf` or
`/etc/trident/tmsd_cc2.conf` for remote machine.

On a clean install, no configuration is currently provided, but you can use the vagrant config as the basis for creating new configuration files.

### Logging

TMS processes use the system logger and are directed to `/var/log/tmsd.log`.
When running a process manually, use the `--stdlog` option to redirect
log messages to the console.

The minimum default log level is "info" for all the processes. Change the
logging level using `--log LEVEL` flag. The log levels are the same as those
used by syslog.

Tracer objects log at the alert level and are only enabled when passing in
`--trace NAME` on the command line where NAME is the given name of the tracer.

## Client

See https://github.com/orolia/prisma-ui for client project.

### Documentation for backend
To see documentation about packages you can run godoc
```bash
gorun -http :6060
```
Go to [http://localhost:6060/pkg/prisma]

### Documentations for daemons
- **[tanalyzed](tms/cmd/daemons/tanalyzed/README.md)**
- **[tauthd](tms/cmd/daemons/tauthd/README.md)**
- **[tdatabased](tms/cmd/daemons/tdatabased/README.md)**
- **[tfleetd](tms/cmd/daemons/tfleetd/README.md)**
- **[tgwad](tms/cmd/daemons/tgwad/README.md)**
- **[tmccd](tms/cmd/daemons//tmccd/README.md)**
- **[tmsd](tms/cmd/daemons/tmsd/README.md)**
- **[tnafexportd](tms/cmd/daemons/tnafexportd/README.md)**
- **[tnoid](tms/cmd/daemons/tnoid/README.md)**
- **[torbcommd](tms/cmd/daemons/torbcommd/README.md)**
- **[treportd](tms/cmd/daemons/treportd/README.md)**
- **[treportd](tms/cmd/daemons/treportd/README.md)**
- **[tselfsign](tms/cmd/daemons/tselfsign/README.md)**
- **[tspiderd](tms/cmd/daemons/tspiderd/README.md)**
- **[twebd](tms/cmd/daemons/twebd/README.md)**
- **[tauthd](tms/cmd/daemons/tauthd/README.md)**
- **[twebd](tms/cmd/daemons/twebd/README.md)**

### Documentations for tools
- **[datagramdownlinkrequest](tms/cmd/tools/datagramdownlinkrequest/README.md)**
- **[generator](tms/cmd/tools/generator/README.md)**
- **[nmea-lib-gen](tms/cmd/tools/nmea-lib-gen/README.md)**
- **[omnicom-iridium](tms/cmd/tools/omnicom-iridium/README.md)**
- **[omnicom-lib-gen](tms/cmd/tools/omnicom-lib-gen/README.md)**
- **[rsm](tms/cmd/tools/rsm/README.md)**
- **[sarsat-start](tms/cmd/tools/sarsat-start/README.md)**
- **[tdemo](tms/cmd/tools/tdemo/README.md)**
- **[tloggernmea](tms/cmd/tools/tloggernmea/README.md)**
- **[tmccrd](tms/cmd/tools/tmccrd/README.md)**
- **[tmongo](tms/cmd/tools/tmongo/README.md)**
- **[tping](tms/cmd/tools/tping/README.md)**
- **[tportald](tms/cmd/tools/tportald/README.md)**
- **[tsimulator](tms/cmd/tools/tsimulator/README.md)**
- **[tstarter-tnoids](tms/cmd/tools/tstarter-tnoids/README.md)**
- **[twatch](tms/cmd/tools/twatch/README.md)**

## Other topics

* [Enabling SSL/TLS security](https://documentation.mcmurdo.io/installation/ssl/)
