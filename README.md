# abcweb

The 'abcweb new' command creates a new Go web app with a default directory
structure and configuration at the path you specify.

You can specify extra command-line arguments to be used every time 'abcweb new'
runs to tell the generator what features and packages you would like enabled.

The abcweb generated web app can function as a server-side rendered web app, 
a client-side rendered web app or a stand-alone web API server.

### Features
 
* Database ORM
* Database migrations
* LiveReload automatic browser refreshing
* Task runner
* Multiple environment configuration
* Command-line argument configuration
* Minification (HTML5/CSS3/JS/JSON/SVG/XML) 
* HTTP2 support
* Flexible routing (supports stdlib context.Context)
* Flexible rendering (HTML templates, JSON, XML, text & binary data)
* Colored & leveled logging
* TLS/SSL support (https://)
* HTTP sessions (supports cookie, disk, memory and redis sessions)

### Packages

* [sqlboiler](https://github.com/vattle/sqlboiler) *-- database ORM generator*
* [switch](https://github.com/nullbio/switch) *-- TOML configuration parser supporting multiple environments and environment var prefixes*
* [sql-migrate](https://github.com/rubenv/sql-migrate) *-- database migrations*
* [lrserver](https://github.com/jaschaephraim/lrserver) *-- LiveReload automatic browser refreshing for front-end development*
* [godo](https://github.com/go-godo/godo) and/or shell scripts *-- task runner for build tasks like minification*
* [cobra](https://github.com/spf13/cobra) *-- command line arguments*
* [minify](https://github.com/tdewolff/minify) *-- HTML/CSS/JS minification build task*
* [zap](https://github.com/uber-go/zap) & [zapcolors](https://github.com/aarondl/zapcolors) *-- colored logging*
* [chi](https://github.com/pressly/chi) *-- routing*
* [render](https://github.com/unrolled/render) *-- dynamic template rendering using render*
* [redis](https://github.com/go-redis/redis) *-- used for the optional redis session storer*

### Configuration

This project loads configuration in the order of:

1. command line argument default values
2. config.toml
3. environment variables
4. supplied command line arguments

This means that values passed into the command line will
override values passed in through the config.toml and env vars.

### FAQ

#### Why didn't you include something to combine asset files?

The HTTP2 spec specifies that concatenating files is no longer recommended
because HTTP2 supports multiplexing and retrieves files in parallel. Having
them as separate files provides speed advantages.



### Dependencies

Bootstrap 4 dependencies

https://jquery.com/
http://tether.io/
http://transitionjs.org/

