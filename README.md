> [!WARNING]
> This project is a work in progress and is not yet ready for use.

# Summary

gh-altar is a GitHub cli extension to be used with [altar](https://github.com/t-monaghan/altar) to display GitHub information on Awtrix powered displays, specifically the examples defined [here](https://github.com/t-monaghan/altar/tree/main/examples/github).

# Usage

Running `gh altar` and any sub-commands with the `--help` flag will provide the documentation for all commands and flags available.

## Configuration

`gh-altar` requires the IP address of your altar broker. You can define this in a config file `altar.yaml` in either `~/.config/` or `~/.config/altar` or in your current directory. The value `broker.address` should be set, with the port number, the default port for an altar broker is `25827`.

For example:
```yaml
broker:
  address: 127.0.0.1:25827
```
