# Mattermost Simple Lock Plugin [![Build Status](https://travis-ci.org/maruTA-bis5/mattermost-simple-lock-plugin.svg?branch=master)](https://travis-ci.org/maruTA-bis5/mattermost-simple-lock-plugin)

- This plugin based on [mattermost/mattermost-plugin-sample](https://github.com/mattermost/mattermost-plugin-sample).
- This plugin is rewrited version of [maruTA-bis5/mattermost-simple-lock](https://github.com/maruTA-bis5/mattermost-simple-lock) integration.

## Usage
Command syntax:
```
/lock TARGET_RESOURCE [MESSAGE]
```
- `TARGET_RESOURCE` is name of the resource you want to use.
- `MESSAGE` is optional message.

Example:
```
/lock Server-001 to use system test
```
![Screenshot](screenshot.png)

## Installation
1. Go to the [releases page of this Github repository](https://github.com/maruTA-bis5/mattermost-simple-lock-plugin/releases) and download the latest release for your Mattermost server.
2. Upload this file in the Mattermost System Console under **System Console > Plugins > Management** to install the plugin. To learn more about how to upload a plugin, [see the documentation](https://docs.mattermost.com/administration/plugins.html#plugin-uploads).
3. Enable `Mattermost Simple Lock Plugin`.

## Development
Build plugin:
```
make
```

This will produce a single plugin file (with support for multiple architectures) for upload to your Mattermost server:

```
dist/net.bis5.mattermost.simplelock.tar.gz
```

There is a build target to automate deploying and enabling the plugin to your server, but it requires configuration and [http](https://httpie.org/) to be installed:
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_USERNAME=admin
export MM_ADMIN_PASSWORD=password
make deploy
```

Alternatively, if you are running your `mattermost-server` out of a sibling directory by the same name, use the `deploy` target alone to  unpack the files into the right directory. You will need to restart your server and manually enable your plugin.

In production, deploy and upload your plugin via the [System Console](https://about.mattermost.com/default-plugin-uploads).

