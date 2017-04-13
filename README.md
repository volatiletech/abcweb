# abcweb

ABCWeb is the only Go web framework you'll need, and was heavily inspired by Rails.

The 'abcweb new' command creates a new Go web app with a default directory
structure and configuration at the path you specify.

You can specify extra command-line arguments to be used every time 'abcweb new'
runs to tell the generator what features and packages you would like enabled.

The abcweb generated web app can function as a server-side rendered web app, 
a client-side rendered web app or a stand-alone web API server.

### Features

Most features can be removed on a per-app basis by specifying
flags to the `abcweb new` command.

* HTTP2 support
* Database ORM
* Database migrations
* Build system and task runner (Gulp)
* Automatic rebuild Go app on code change
* Automatic asset compile on asset change
* Automatic browser refreshing with LiveReload
* [Font-Awesome](http://fontawesome.io/) and [Twitter Bootstrap 4](https://v4-alpha.getbootstrap.com/)
* SCSS and LESS support
* Asset fingerprinting and minification
* Infinite environments in configuration
* Command-line and environment variable configuration
* Flexible routing (stdlib context.Context) 
* Flexible rendering (HTML, JSON, XML, text & binary data)
* Colored & leveled logging 
* TLS1.2/SSL support
* HTTP sessions (supports cookie, disk, memory and redis sessions) 
* Flash messages 
* Rendering interface to add support for any templating engine
* Vendored dependencies to ensure consistent compatibility 
* Many more features!

### Packages

* [sqlboiler](https://github.com/vattle/sqlboiler) *-- Database ORM generator*
* [mig](https://github.com/volatiletech/mig) *-- Database migrations*
* [abcsessions](https://github.com/volatiletech/abcsessions) *-- HTTP sessions*
* [abcmiddleware](https://github.com/volatiletech/abcmiddleware) *-- Zap logging and panic recovery middleware*
* [abcrender](https://github.com/volatiletech/abcrender) *-- Rendering interface to support other templating engines*
* [render](https://github.com/unrolled/render) *-- Template rendering*
* [cobra](https://github.com/spf13/cobra) *-- Command line arguments*
* [viper](https://github.com/spf13/viper) *-- Configuration loading*
* [zap](https://github.com/uber-go/zap) *-- Logging* 
* [chi](https://github.com/pressly/chi) *-- Routing*

### Configuration

This project loads configuration in the order of:

1. Command line argument default values
2. config.toml
3. Environment variables
4. Supplied command line arguments

This means that values passed into the command line will
override values passed in through the config.toml and env vars.

```
ABCWeb is a tool to help you scaffold and develop Go web applications.

Usage:
  abcweb [command]

Available Commands:
  build       Builds your abcweb binary and executes the gulp build task
  deps        Download and optionally update all abcweb dependencies
  dev         Runs your abcweb app for development
  gen         Generate your database models and migration files
  help        Help about any command
  migrate     Run migration tasks (up, down, redo, status, version)
  new         Generate a new abcweb app
  test        Runs the tests for your abcweb app

Flags:
      --version   Print the abcweb version

Use "abcweb [command] --help" for more information about a command.
```

### FAQ

### How do I install nodejs, npm and gulp?

Installing nodejs is system dependant. `nvm` is a nice option on some systems
but is not supported in some shells such as [fish shell](https://fishshell.com/).
NPM comes bundled with NodeJS.

[Download Node.js](https://nodejs.org/en/download/)
[Installing Node.js via package manager](https://nodejs.org/en/download/package-manager/)


ABCWeb uses Gulp 4 as its task runner and asset build system. Once nodejs and 
npm is installed you can install Gulp 4 using:

`abcweb deps -u`

if you get permission errors you can use the permissions fix described [here.](https://docs.npmjs.com/getting-started/fixing-npm-permissions)

### Why didn't you include something to combine asset files?

The HTTP2 spec specifies that concatenating files is no longer recommended
because HTTP2 supports multiplexing and retrieves files in parallel. Having
them as separate files provides speed advantages.

### I'm getting nodejs npm permission errors

These errors look something like: `npm WARN checkPermissions Missing write access to /usr/lib/node_modules`.
Instructions for fixing this can be found [here.](https://docs.npmjs.com/getting-started/fixing-npm-permissions)

### Dependencies

Bootstrap 4 dependencies (included by default):

[JQuery](https://jquery.com/)
[Tether](http://tether.io/)
[TransitionJS](http://transitionjs.org/)

