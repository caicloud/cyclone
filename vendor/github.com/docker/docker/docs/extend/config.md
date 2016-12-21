---
title: "Plugin config"
description: "How develop and use a plugin with the managed plugin system"
keywords: "API, Usage, plugins, documentation, developer"
---

<!-- This file is maintained within the docker/docker Github
     repository at https://github.com/docker/docker/. Make all
     pull requests against that repo. If you see this file in
     another repository, consider it read-only there, as it will
     periodically be overwritten by the definitive file. Pull
     requests which include edits to this file in other repositories
     will be rejected.
-->


# Plugin Config Version 0 of Plugin V2

This document outlines the format of the V0 plugin configuration. The plugin
config described herein was introduced in the Docker daemon in the [v1.12.0
release](https://github.com/docker/docker/commit/f37117045c5398fd3dca8016ea8ca0cb47e7312b).

Plugin configs describe the various constituents of a docker plugin. Plugin
configs can be serialized to JSON format with the following media types:

Config Type  | Media Type
------------- | -------------
config  | "application/vnd.docker.plugin.v0+json"


## *Config* Field Descriptions

Config provides the base accessible fields for working with V0 plugin format
 in the registry.

- **`description`** *string*

	description of the plugin

- **`documentation`** *string*

  	link to the documentation about the plugin

- **`interface`** *PluginInterface*

   interface implemented by the plugins, struct consisting of the following fields

    - **`types`** *string array*

      types indicate what interface(s) the plugin currently implements.

      currently supported:

      	- **docker.volumedriver/1.0**

      	- **docker.authz/1.0**

    - **`socket`** *string*

      socket is the name of the socket the engine should use to communicate with the plugins.
      the socket will be created in `/run/docker/plugins`.


- **`entrypoint`** *string array*

   entrypoint of the plugin, see [`ENTRYPOINT`](../reference/builder.md#entrypoint)

- **`workdir`** *string*

   workdir of the plugin, see [`WORKDIR`](../reference/builder.md#workdir)

- **`network`** *PluginNetwork*

   network of the plugin, struct consisting of the following fields

    - **`type`** *string*

      network type.

      currently supported:

      	- **bridge**
      	- **host**
      	- **none**

- **`mounts`** *PluginMount array*

   mount of the plugin, struct consisting of the following fields, see [`MOUNTS`](https://github.com/opencontainers/runtime-spec/blob/master/config.md#mounts)

    - **`name`** *string*

	  name of the mount.

    - **`description`** *string*

      description of the mount.

    - **`source`** *string*

	  source of the mount.

    - **`destination`** *string*

	  destination of the mount.

    - **`type`** *string*

      mount type.

    - **`options`** *string array*

	  options of the mount.

- **`env`** *PluginEnv array*

   env of the plugin, struct consisting of the following fields

    - **`name`** *string*

	  name of the env.

    - **`description`** *string*

      description of the env.

    - **`value`** *string*

	  value of the env.

- **`args`** *PluginArgs*

   args of the plugin, struct consisting of the following fields

    - **`name`** *string*

	  name of the env.

    - **`description`** *string*

      description of the env.

    - **`value`** *string array*

	  values of the args.

- **`linux`** *PluginLinux*

    - **`capabilities`** *string array*

          capabilities of the plugin (*Linux only*), see list [`here`](https://github.com/opencontainers/runc/blob/master/libcontainer/SPEC.md#security)

    - **`devices`** *PluginDevice array*

          device of the plugin, (*Linux only*), struct consisting of the following fields, see [`DEVICES`](https://github.com/opencontainers/runtime-spec/blob/master/config-linux.md#devices)

          - **`name`** *string*

	      name of the device.

          - **`description`** *string*

              description of the device.

          - **`path`** *string*

              path of the device.

## Example Config

*Example showing the 'tiborvass/no-remove' plugin config.*

```json
{
  "description": "A test plugin for Docker",
  "documentation": "https://docs.docker.com/engine/extend/plugins/",
  "entrypoint": ["plugin-no-remove", "/data"],
  "interface": {
    "types": ["docker.volumedriver/1.0"],
    "socket": "plugins.sock"
  },
  "network": {
    "type": "host"
  },
  "mounts": [
    {
      "source": "/data",
      "destination": "/data",
      "type": "bind",
      "options": ["shared", "rbind"]
    },
    {
      "destination": "/foobar",
      "type": "tmpfs"
    }
  ],
  "args": {
    "name": "args",
    "description": "command line arguments",
    "value": []
  },
  "env": [
    {
      "name": "DEBUG",
      "description": "If set, prints debug messages",
      "value": "1"
    }
  ],
  "linux": {
    "devices": [
      {
        "name": "device",
        "description": "a host device to mount",
        "path": "/dev/cpu_dma_latency"
      }
    ]
  }
}
```
