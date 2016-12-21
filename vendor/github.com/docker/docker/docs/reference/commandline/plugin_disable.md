---
title: "plugin disable"
description: "the plugin disable command description and usage"
keywords: "plugin, disable"
---

<!-- This file is maintained within the docker/docker Github
     repository at https://github.com/docker/docker/. Make all
     pull requests against that repo. If you see this file in
     another repository, consider it read-only there, as it will
     periodically be overwritten by the definitive file. Pull
     requests which include edits to this file in other repositories
     will be rejected.
-->

# plugin disable

```markdown
Usage:  docker plugin disable PLUGIN

Disable a plugin

Options:
      --help   Print usage
```

Disables a plugin. The plugin must be installed before it can be disabled,
see [`docker plugin install`](plugin_install.md).


The following example shows that the `no-remove` plugin is installed
and enabled:

```bash
$ docker plugin ls

NAME                  TAG                 DESCRIPTION                ENABLED
tiborvass/no-remove   latest              A test plugin for Docker   true
```

To disable the plugin, use the following command:

```bash
$ docker plugin disable tiborvass/no-remove

tiborvass/no-remove

$ docker plugin ls

NAME                  TAG                 DESCRIPTION                ENABLED
tiborvass/no-remove   latest              A test plugin for Docker   false
```

## Related information

* [plugin create](plugin_create.md)
* [plugin enable](plugin_enable.md)
* [plugin inspect](plugin_inspect.md)
* [plugin install](plugin_install.md)
* [plugin ls](plugin_ls.md)
* [plugin push](plugin_push.md)
* [plugin rm](plugin_rm.md)
* [plugin set](plugin_set.md)
