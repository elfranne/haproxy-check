[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/sensu/haproxy-check)
![Go Test](https://github.com/sensu/haproxy-check/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/sensu/haproxy-check/workflows/goreleaser/badge.svg)

# HAProxy Check

## Overview
This is a Sensu check that checks the health and status of an HAProxy instance.

## Functionality

TODO

## Releases with Github Actions

To release a new version of this project, simply tag the target sha with a semver release without a `v`
prefix (ex. `1.0.0`). This will trigger the [GitHub action][5] workflow to [build and release][4]
the plugin with goreleaser. Register the asset with [Bonsai][8] to share it with the community!

***

# HAProxy Check

## Table of Contents
- [Overview](#overview)
- [Files](#files)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Check definition](#check-definition)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Usage examples

## Configuration

### Asset registration

```
sensuctl asset add sensu/haproxy-check
```

You can also find the asset on the [Bonsai Asset Index][https://bonsai.sensu.io/assets/sensu/haproxy-check].

### Check definition

```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: haproxy-check
  namespace: default
spec:
  command: haproxy-check --example example_arg
  subscriptions:
  - system
  runtime_assets:
  - sensu/haproxy-check
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the haproxy-check repository:

```
go build
```

## Additional notes

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[2]: https://github.com/sensu-community/sensu-plugin-sdk
[3]: https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md
[4]: https://github.com/sensu-community/check-plugin-template/blob/master/.github/workflows/release.yml
[5]: https://github.com/sensu-community/check-plugin-template/actions
[6]: https://docs.sensu.io/sensu-go/latest/reference/checks/
[7]: https://github.com/sensu-community/check-plugin-template/blob/master/main.go
[8]: https://bonsai.sensu.io/
[9]: https://github.com/sensu-community/sensu-plugin-tool
[10]: https://docs.sensu.io/sensu-go/latest/reference/assets/
