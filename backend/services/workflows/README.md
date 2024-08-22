[![Build Status](https://gitlab.com/Northern.tech/Mender/workflows/badges/master/pipeline.svg)](https://gitlab.com/Northern.tech/Mender/workflows/pipelines)
[![Coverage Status](https://coveralls.io/repos/github/mendersoftware/workflows/badge.svg?branch=master)](https://coveralls.io/github/mendersoftware/workflows?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mendersoftware/workflows)](https://goreportcard.com/report/github.com/mendersoftware/workflows)
[![Docker pulls](https://img.shields.io/docker/pulls/mendersoftware/workflows.svg?maxAge=3600)](https://hub.docker.com/r/mendersoftware/workflows/)

Mender: Workflows orchestrator
==============================

## Listing jobs

Current state of jobs can be checked via list-jobs commandline:
```shell script
#./workflows list-jobs --page 1 --perPage 4
all jobs: 750; page: 1/187 perPage:4
                  insert time                       id     status workflow
Thu, 16 Apr 2020 12:27:05 UTC 5e984f1958e80bdb83970d8a       done update_device_status
Thu, 16 Apr 2020 12:26:50 UTC 5e984f0a58e80bdb83970d89     failed provision_device
Thu, 16 Apr 2020 12:26:50 UTC 5e984f0a58e80bdb83970d88       done update_device_status
Thu, 16 Apr 2020 12:26:50 UTC 5e984f0a58e80bdb83970d87     failed provision_device
all jobs: 750; page: 1/187
```

## General

Mender is an open source over-the-air (OTA) software updater for embedded Linux
devices. Mender comprises a client running at the embedded device, as well as
a server that manages deployments across many devices.

This repository contains the Mender Workflows orchestrator, which is part of the
Mender server. The Mender server is designed as a microservices architecture
and comprises several repositories.

## Getting started

To start using Mender, we recommend that you begin with the Getting started
section in [the Mender documentation](https://docs.mender.io/).

## Building from source

As the Mender server is designed as microservices architecture, it requires several
repositories to be built to be fully functional. If you are testing the Mender server it
is therefore easier to follow the getting started section above as it integrates these
services.

If you would like to build the Workflows orchestrator independently, you can follow
these steps:

```
git clone https://github.com/mendersoftware/workflows.git
cd workflows
make build
```

## Contributing

We welcome and ask for your contribution. If you would like to contribute to Mender, please read our guide on how to best get started [contributing code or
documentation](https://github.com/mendersoftware/mender/blob/master/CONTRIBUTING.md).

## License

Mender is licensed under the Apache License, Version 2.0. See
[LICENSE](https://github.com/mendersoftware/workflows/blob/master/LICENSE) for the
full license text.

## Security disclosure

We take security very seriously. If you come across any issue regarding
security, please disclose the information by sending an email to
[security@mender.io](security@mender.io). Please do not create a new public
issue. We thank you in advance for your cooperation.

## Connect with us

* Join the [Mender Hub discussion forum](https://hub.mender.io)
* Follow us on [Twitter](https://twitter.com/mender_io). Please
  feel free to tweet us questions.
* Fork us on [Github](https://github.com/mendersoftware)
* Create an issue in the [bugtracker](https://northerntech.atlassian.net/projects/MEN)
* Email us at [contact@mender.io](mailto:contact@mender.io)
* Connect to the [#mender IRC channel on Libera](https://web.libera.chat/?#mender)
