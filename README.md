# ABCWeb

[![License](https://img.shields.io/badge/license-BSD-blue.svg)](https://github.com/volatiletech/abcweb/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/volatiletech/abcweb?status.svg)](https://godoc.org/github.com/volatiletech/abcweb)
[![Go Report Card](https://goreportcard.com/badge/volatiletech/abcweb)](http://goreportcard.com/report/volatiletech/abcweb)

ABCWeb was heavily inspired by Rails for its ease-of-use, flexibility, 
and development speed. The goal of this project is to make developing a 
web app in Go just as painless. We've tried to remain as unopinionated
as possible to make ABCWeb suited for many different application types,
and our big selling point is the packages we've chosen, which we believe
are the clear leaders in the eco-system.

Getting started is as simple as running the `abcweb new` command which will
generate you a new Go web app that comes working out of the box.

You can specify extra `abcweb new` command-line arguments to tell the generator 
what features and packages you would like enabled.

This customizability allows abcweb to suit your requirements, whether that be a 
server-side rendered web app (html templates), a client-side rendered 
web app (reactjs, angularjs) or a stand-alone web API server.

## Contents

- [ABCWeb](#abcweb)
    - [Features](#features)
    - [Packages](#packages)
    - [Getting Started](#getting-started)
    - [Usage](#usage)
    - [Configuration](#configuration)
    - [Gulp](#gulp)
    - [FAQ](#faq)

## Features

Most features can be removed on a per-app basis by specifying
flags to the `abcweb new` command.

* HTTP2 support
* Database ORM
* Database migrations
* Build system and task runner (Gulp 4)
* Automatic asset compilation on asset change
* Automatic browser refreshing with LiveReload
* Automatically rebuild Go app on code change
* Automatically run migrations against test database for testing
* SCSS and LESS support
* Asset fingerprinting, compilation, minification and gzip
* [Font-Awesome](http://fontawesome.io/) and [Twitter Bootstrap 4](https://v4-alpha.getbootstrap.com/)
* Infinite environments in configuration
* Command-line and environment variable configuration
* Flexible routing (stdlib context.Context) 
* Flexible rendering (HTML, JSON, XML, text & binary data)
* Colored and leveled logging 
* TLS1.2/SSL support
* HTTP sessions (supports cookie, disk, memory and redis sessions) 
* Flash messages 
* Rendering interface to easily add support for any templating engine
* Vendored dependencies to ensure consistent compatibility 
* Many more features!

## Packages

ABCWeb uses a collection of the very best open-source projects and packages the
Go community has to offer. These packages were chosen specifically because 
they are fast, intelligently designed, easy to use and modern. For a full list
see [PACKAGES.md](https://github.com/volatiletech/abcweb/blob/master/PACKAGES.md):

#### Database ORM

[SQLBoiler](https://github.com/vattle/sqlboiler) is one of our other core projects and was a natural 
fit for ABCWeb. It is the fastest ORM by far (on par with stdlib), it is featureful 
and has excellent relationship support, and its generation approach allows for type-safety and
editor auto-completion. We've made using SQLBoiler easy by bundling it into the `abcweb gen` command.

#### Database Migrations

[mig](https://github.com/volatiletech/mig) is our fork of [Goose](https://github.com/pressly/goose)
that patches some big issues and was tweaked to make it work better with ABCWeb. It does
everything you'd expect a migration tool to do, and has both a library and command-line component.
Mig supports MySQL and Postgres at present. It has been bundled into the `abcweb gen` 
and `abcweb migrate` commands.

#### Sessions, Cookies & Flash Messages

[ABCSessions](https://github.com/volatiletech/abcsessions) was designed to make working with 
HTTP sessions and cookies a breeze, and it also comes with a flash messages API. ABCSessions 
ships with disk, memory, redis and cookie storers, and the ability to easily add new storers using
our provided interfaces.

#### Rendering API

[Render](https://github.com/unrolled/render) is a package that provides functionality for 
easily rendering JSON, XML, text, binary data, and HTML templates. We have also written an interface
wrapper for Render ([ABCRender](https://github.com/volatiletech/abcrender)) that allows you to
easily add support for any templating engine you choose if Go's `html/template` is not enough for you.

#### Routing

[Chi](https://github.com/pressly/chi) is one of the fastest and most modern routers in the eco-system
and is starting to gain a cult following. Chi was built on the `context` package that was introduced
in Go 1.7. It's elegant API design and stdlib-only philosophy is what has Chi standing out from the rest.

#### Logging

[Zap](https://github.com/uber-go/zap) was written by Uber, and is widely regarded as the fastest and
most performant logging package, even rivaling the standard library. It is a structured, leveled &
colored logger.

#### Command Line and Configuration

ABCWeb comes with [Cobra](https://github.com/spf13/cobra) for handling command-line arguments and
[Viper](https://github.com/spf13/viper) for handling configuration loading. These packages are widely
known, widely used and widely enjoyed.

#### Vendoring

When ABCWeb generates your app it also includes a `vendor.json` file with all of its dependencies
in it. The `abcweb new` command will automatically sync your `vendor` folder on generation.
[Govendor](github.com/kardianos/govendor) was the tool of choice here. It's functional and easy to use.

#### Build System, Asset Pipeline & Task Running

ABCWeb uses [Gulp](https://github.com/gulpjs/gulp/tree/4.0) to handle asset compilation,
minification, task running and live reloading of the browser on changes to the asset files or
html templates. Read the [Gulp section of this readme](#gulp) if you'd like further information.
ABCWeb also uses [refresh](https://github.com/markbates/refresh) to rebuild your go web app on changes 
to your configuration or .go files. Refresh can be highly customized using the `watch.toml` config file.

[Read here](#why-did-you-choose-to-use-gulp) if you're wondering why we chose a NodeJS dependency. 
Also note that it's optional, but highly recommended due to the conveniences it provides.

## Getting Started

ABCWeb requires Go 1.8 or higher.

It's dead easy to generate a web app using ABCWeb.

### Step 1:
[Install NodeJS](#how-do-i-install-nodejs-npm-and-gulp).

This is required for the asset pipeline (Gulp). Using Gulp is optional (`--no-gulp`), 
but we highly recommend it because it makes the development process SO much easier.

### Step 2:
```shell
# download and install abcweb
go get -u github.com/volatiletech/abcweb

# install and upgrade all dependencies (including Gulp)
abcweb deps -u 

# generate your app (abcweb automatically uses the GOPATH to find your src folder)
abcweb new github.com/username/myapp
```

Your app has now been generated!

#### Where to from here?

```shell
# cd into your new project folder
cd $GOPATH/src/github.com/username/myapp 

# Run abcweb dev for auto-rebuild of app, assets and LiveReload.
# You can optionally set your default environment in config.toml
MYAPP_ENV=dev abcweb dev
```

Navigate your browser to your now running server at http://localhost:4000/, 
modify your `templates/main/home.html` template, and you should see your changes
automatically load. Awesome!

Note that changes to the `.go` files or `.toml` config files will automatically
rebuild your go web-server, however they will require a manual browser refresh
(automatically refreshing in the middle of server changes could put you in 
a pickle, so we decided against it).

## Usage

```
ABCWeb is a tool to help you scaffold and develop Go web applications.

Usage:
  abcweb [command]

Available Commands:
  build       Builds your abcweb binary and executes the gulp build task
  deps        Download and optionally update all abcweb dependencies
  dev         Runs your abcweb app for development
  dist        Dist creates a distribution bundle for easy deployment
  gen         Generate your database models and migration files
  help        Help about any command
  migrate     Run migration tasks (up, down, redo, status, version)
  new         Generate a new abcweb app
  revendor    Revendor updates your vendor.json file and resyncs your vendor folder
  test        Runs the tests for your abcweb app

Flags:
  -h, --help      help for abcweb
      --version   Print the abcweb version

Use "abcweb [command] --help" for more information about a command.
```

## Configuration

This project loads configuration in the order of:

1. Command line argument default values
2. config.toml
3. Environment variables
4. Supplied command line arguments

This means that values passed into the command line will
override values passed in through the config.toml and env vars.

* App configuration is found in `config.toml`.
* `abcweb dev` Go auto-rebuild configuration is found in `watch.toml`.
* Asset pipeline, task runner and build system configuration is found in `gulpfile.js`.

## Gulp

See the [FAQ](#faq) for installation instructions.

When ABCWeb generates your app it also includes a `gulpfile.js` for you that has been written
to perform all steps of the build process incrementally. Out of the box this includes
SCSS, LESS, CSS, JS, fonts, video, images and audio. The `gulpfile.js` has also been configured to
parse your CSS assets through [PostCSS Autoprefixer](https://github.com/postcss/autoprefixer) so
CSS vendor prefixes are a thing of the past. This is also a Bootstrap 4 dependency.

Your gulp file also comes with a watch task that can be run using `abcweb dev` that will 
watch all of your asset files and templates for changes, recompile them if necessary,
move them to your public assets folder and reload your browser automatically using LiveReload.
You can run this task manually using `gulp watch` if desired, but it works better through `abcweb dev`
so you also can take advantage of the Go app rebuilding.

Once you're ready to build your assets for production, it's as simple as calling `abcweb build` which will
build your Go binary for deployment and then run the gulp task called `build`. This build task
will first remove all files in the public assets directory, then it will compile,
minify, gzip and fingerprint all assets and then generate a `manifest.json` file
that will be loaded by your app in production mode. The assets manifest maps all of the
incoming file names to the fingerprinted asset file names, for example: `{"/css/main.css": "/css/main-a2e4fe.css"}`.

Once you've finishing building your binary and assets, all you need to do is deploy your binary,
your configuration files and your `public/assets` folder to your production server.

## FAQ

### Why did you choose to use Gulp?

We decided to use [Gulp](https://github.com/gulpjs/gulp/tree/4.0) for our build system and
task running. We realize that some people may not enjoy having a NodeJS dependency so we've
made this entirely optional (`abcweb new --no-gulp`), however we highly recommend using it
due to the conveniences it provides. Unfortunately there are no robust solutions in the Go
ecosystem for this problem yet, and when we started to [make our own](https://github.com/nullbio/pipedream)
we quickly realized that not only would it not work effectively for a multitude of reasons,
but it would also never be as flexible and simple to use as Gulp is due to the fact
that Go is compiled and all of the existing asset tools out there are either written in or
written for Javascript. With that being said, Gulp is extremely easy to use, and ABCWeb
makes it even easier to use.

### How do I install NodeJS, NPM and Gulp?

Installing NodeJS is system dependant. `nvm` is a nice option on some systems
but is not supported in some shells such as [fish shell](https://fishshell.com/).
NPM comes bundled with NodeJS.

* [Download Node.js](https://nodejs.org/en/download/)
* [Installing Node.js via package manager](https://nodejs.org/en/download/package-manager/)


ABCWeb uses [Gulp 4](https://github.com/gulpjs/gulp/tree/4.0) as its task runner and asset build system. Once NodeJS and 
NPM is installed you can install [Gulp](https://github.com/gulpjs/gulp/tree/4.0) using:

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

* [JQuery](https://jquery.com/)
* [Tether](http://tether.io/)
* [TransitionJS](http://transitionjs.org/)

### Where is the homepage?

The homepage for the ABCWeb Golang web app framework is located at: https://github.com/volatiletech/abcweb

