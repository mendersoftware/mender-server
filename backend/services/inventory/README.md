[![Build Status](https://gitlab.com/Northern.tech/Mender/inventory/badges/master/pipeline.svg)](https://gitlab.com/Northern.tech/Mender/inventory/pipelines)
[![Coverage Status](https://coveralls.io/repos/github/mendersoftware/inventory/badge.svg?branch=master)](https://coveralls.io/github/mendersoftware/inventory?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mendersoftware/inventory)](https://goreportcard.com/report/github.com/mendersoftware/inventory)
[![Docker pulls](https://img.shields.io/docker/pulls/mendersoftware/inventory.svg?maxAge=3600)](https://hub.docker.com/r/mendersoftware/inventory/)

Mender: Device Inventory Service
==============================================

Mender is an open source over-the-air (OTA) software updater for embedded Linux
devices. Mender comprises a client running at the embedded device, as well as
a server that manages deployments across many devices.

This repository contains the Mender Device Inventory Service, which is part of the
Mender server. The Mender server is designed as a microservices architecture
and comprises several repositories.

The Device Inventory Service stores attributes about devices reported by Mender clients, and supports
searching and sorting of attributes. In general, the reported attributes
are freely defined by the device vendor, and can be used to group devices - based on selected
criteria - into groups for which independent deployments can be rolled out.

![Mender logo](https://raw.githubusercontent.com/mendersoftware/mender/master/mender_logo.png)


## Getting started

To start using Mender, we recommend that you begin with the Getting started
section in [the Mender documentation](https://docs.mender.io/).


## Building from source

As the Mender server is designed as microservices architecture, it requires several
repositories to be built to be fully functional. If you are testing the Mender server it
is therefore easier to follow the getting started section above as it integrates these
services.

If you would like to build the Device Inventory service independently, you can follow
these steps:

```
git clone https://github.com/mendersoftware/inventory.git
cd inventory
go build
```

## Configuration

The service can be configured by:
* providing configuration file (supports JSON, TOML, YAML and HCL formatting).
The default configuration file is provided to be downloaded from [config.yaml](https://github.com/mendersoftware/inventory/blob/master/config.yaml).
* setting environment variables. The service will check for a environment variable
with a name matching the key uppercased and prefixed with "INVENTORY_".
Eg. for "listen" the variable name is "INVENTORY_LISTEN".

## Contributing

We welcome and ask for your contribution. If you would like to contribute to Mender, please read our guide on how to best get started [contributing code or
documentation](https://github.com/mendersoftware/mender/blob/master/CONTRIBUTING.md).

## License

Mender is licensed under the Apache License, Version 2.0. See
[LICENSE](https://github.com/mendersoftware/inventory/blob/master/LICENSE) for the
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
