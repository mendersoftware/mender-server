[![Build Status](https://gitlab.com/Northern.tech/Mender/gui/badges/master/pipeline.svg)](https://gitlab.com/Northern.tech/Mender/gui/pipelines)
[![Docker pulls](https://img.shields.io/docker/pulls/mendersoftware/gui.svg?maxAge=3600)](https://hub.docker.com/r/mendersoftware/gui/)
[![code style: prettier](https://img.shields.io/badge/code_style-prettier-ff69b4.svg?style=flat-square)](https://github.com/prettier/prettier)
[![Coverage Status](https://coveralls.io/repos/github/mendersoftware/gui/badge.svg?branch=master)](https://coveralls.io/github/mendersoftware/gui?branch=master)

# Mender: GUI

Mender is an open source over-the-air (OTA) software updater for embedded Linux
devices. Mender comprises a client running at the embedded device, as well as
a server that manages deployments across many devices.

This repository contains the Mender Graphical User Interface, which is part of the
Mender server. The Mender server is designed as a microservices architecture
and comprises several repositories.

The GUI exposes the entirety of Mender's functionality to the user, including
deployment definition, scheduling and monitoring, device listing and admission, etc.

<img src="src/assets/img/loginlogo.svg" alt="Mender" width="33%" />

## Getting started

To start using Mender, we recommend that you begin with the Getting started
section in [the Mender documentation](https://docs.mender.io/).

## Contributing

We welcome and ask for your contribution. If you would like to contribute to Mender, please read our guide on how to best get started [contributing code or
documentation](https://github.com/mendersoftware/mender/blob/master/CONTRIBUTING.md).

## Development

In order to develop the GUI code to contribute, it is required to [nodejs](https://nodejs.org) and [npm](https://github.com/npm/cli) installed, followed
by an execution of `npm install` to install all the dependencies in package.json. Work on the files in 'src'. Running `npm run watch` will watch for your
changes as you develop and build as you go.
The `watch` command is intended to build however the content should be served from within the regular container running in your local mender setup.
To connect the two a typical approach is to create a `docker-compose.override.yml` in the folder you have cloned the integration repo into (to isolate the changes made to the running environment in a single file) with the following content:

```
version: '2.1'
services:
    mender-gui:
        volumes:
            - ../<path-to-clone-of-the-gui-repo>/dist:/var/www/mender-gui/dist
```

and then start the demo environment with this file like this: `./demo -f docker-compose.override.yml up` from the integration clone folder. The content should be accessible under https://localhost/ just as without any modifications. While running `npm run watch` the results of each incremental build should show up on page refresh.
To ensure committed changes pass the CI pipeline it is good to run `npm run lint` and `npm run test` before you submit your changes for review.

The project is equipped with commit checks powered by [husky](https://github.com/typicode/husky) that check for linter problems and run tests on the changed files. To also check your commit messages and fix potential problems with them, make sure to set `MENDER_TESTING` in your `ENV` to a local clone of the [mendertesting](https://github.com/mendersoftware/mendertesting) repository.

## License

Mender is licensed under the Apache License, Version 2.0. See
[LICENSE](https://github.com/mendersoftware/gui/blob/master/LICENSE) for the
full license text.

## Security disclosure

We take security very seriously. If you come across any issue regarding
security, please disclose the information by sending an email to
[security@mender.io](security@mender.io). Please do not create a new public
issue. We thank you in advance for your cooperation.

## Connect with us

- Join the [Mender Hub discussion forum](https://hub.mender.io)
- Follow us on [Twitter](https://twitter.com/mender_io). Please
  feel free to tweet us questions.
- Fork us on [Github](https://github.com/mendersoftware)
- Create an issue in the [bugtracker](https://northerntech.atlassian.net/projects/MEN)
- Email us at [contact@mender.io](mailto:contact@mender.io)
- Connect to the [#mender IRC channel on Libera](https://web.libera.chat/?#mender)

## Authors

Mender was created by the team at [Northern.tech AS](https://northern.tech), with many contributions from
the community. Thanks [everyone](https://github.com/mendersoftware/mender/graphs/contributors)!

[Mender](https://mender.io) is sponsored by [Northern.tech AS](https://northern.tech).
