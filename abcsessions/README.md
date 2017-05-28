# ABCSessions

[![License](https://img.shields.io/badge/license-BSD-blue.svg)](https://github.com/volatiletech/abcweb/abcsessions/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/volatiletech/abcweb/abcsessions?status.svg)](https://godoc.org/github.com/volatiletech/abcweb/abcsessions)
[![Go Report Card](https://goreportcard.com/badge/volatiletech/abcweb/abcsessions)](http://goreportcard.com/report/volatiletech/abcweb/abcsessions)

## Available Session Storers

* Disk
* Memory
* Redis
* Cookie

## API Operations

Note, you should avoid mixing and matching API calls for key-value versus object.
Stick to one mode for every app you generate (either simple sessions using key-value
strings and the regular helpers or a custom sessions struct using the object helpers).
This is the main API entry point (key-value or object API) to use.

###  Key-Value string API

These session API endpoints are ideal for apps that only require simple session strings
to be stored and not a full custom sessions struct with a lot of differing variables
and types.

```golang
// Set adds a key-value pair to your session. 
Set(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, value string) error

// Get retrieves a key-value pair value stored in your session.
Get(overseer Overseer, w http.ResponseWriter, r *http.Request, key string) (string, error)

// Del deletes a key-value pair from your session.
Del(overseer Overseer, w http.ResponseWriter, r *http.Request, key string) error

// AddFlash adds a flash message key-value pair to your session.
AddFlash(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, value string) error

// GetFlash retrieves a flash message key-value pair value stored in your session. 
// Note that flash messages are automatically deleted after being retrieved once.
GetFlash(overseer Overseer, w http.ResponseWriter, r *http.Request, key string) (string, error)
```

### Object API

These session API endpoints are ideal for apps that are best served with a sessions
struct opposed to individual strings. This flexibility allows you to define your
own custom sessions struct with a range of differing variable types. See [examples](#examples)
for Object API usage.

```golang
// SetObj adds any object or variable to your session.
SetObj(overseer Overseer, w http.ResponseWriter, r *http.Request, value interface{}) error

// GetObj retrieves an object or variable stored in your session.
GetObj(overseer Overseer, w http.ResponseWriter, r *http.Request, value interface{}) error

// AddFlashObj adds an object as a flash message to your session.
AddFlashObj(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, value interface{}) error

// GetFlashObj retrieves a object as a flash message stored in your session.
// Note that flash messages are automatically deleted after being retrieved once.
GetFlashObj(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, pointer interface{}) error
```

## Overseer interface

The job of an Overseer is to interface with your storers and manage your session cookies.
If you want to work with the session values you probably want to use the higher level
API functions listed above, but if you want to work with the session directly
you can use these overseer methods.

`Resetter`

The Overseer interface imbeds the Resetter interface which defines the functions
used to reset the expiries of your sessions. This is handled automatically
by the ResetMiddleware and is not something that needs to be called directly.

```golang
// Set creates or updates a session with a raw value.
Set(w http.ResponseWriter, r *http.Request, value string) error

// Get the raw value stored in a session.
Get(w http.ResponseWriter, r *http.Request) (value string, err error)

// Delete a session and signal the browser to expire the session cookie.
Del(w http.ResponseWriter, r *http.Request) error

// Regenerate a new session ID for your session.
Regenerate(w http.ResponseWriter, r *http.Request) error

// SessionID returns the session ID for your session.
SessionID(w http.ResponseWriter, r *http.Request) (id string, err error)

```

## Storer interface

In the case of session management, "keys" here are synonymous with session IDs. 
This is as lower level API and generally not used too much, because the above 
higher-level API functions call these functions. This can be used if you need
to deal with the storer directly for whatever reason.

```golang
// All retrieves all keys in the store.
All() (keys []string, err error)

// Get retrieves the value stored against a specific key.
Get(key string) (value string, err error)

// Set a key-value pair in the store.
Set(key, value string) error

// Delete a key-value pair from the store.
Del(key string) error

// Reset the expiry of a key in the store.
ResetExpiry(key string) error
```

## Available Overseers

```golang
// StorageOverseer is used for all server-side sessions (disk, memory, redis, etc).
NewStorageOverseer(opts CookieOptions, storer Storer) *StorageOverseer

//CookieOverseer is used for client-side only cookie sessions.
NewCookieOverseer(opts CookieOptions, secretKey [32]byte) *CookieOverseer
```

## How does each Storer work?

### Disk

Disk sessions store the session as a text file on disk. By default they store in 
the systems temp directory under a folder that is randomly generated when you 
generate your app using abcweb app generator command. The file names are the UUIDs 
of the session. Each time the file is accessed (using Get, Set, manually on 
disk, or by using the ResetMiddleware) it will reset the access time of the file, 
which will push back the expiration defined by maxAge. For example, if your
maxAge is set to 1 week, and your cleanInterval is set to 2 hours, then every 2
hours the cleaner will find all disk sessions files that have not been accessed 
for over 1 week and delete them. If the user refreshes a website and you're using
the ResetMiddleware then that 1 week timer will be reset. If your maxAge and 
cleanInterval is set to 0 then these disk session files will permanently persist, 
however the browser will still expire sessions depending on your cookieOptions 
maxAge configuration. In a typical (default) setup, cookieOptions will be set to 
maxAge 0 (expire on browser close), your DiskStorer maxAge will be set to 2 days,
and your DiskStorer cleanInterval will be set to 1 hour.

### Memory

Memory sessions are stored in memory in a mutex protected map[string]memorySession.
The key to the map is the session ID and the memorySession stores the value and expiry
of the session. The memory storer also has methods to start and stop a cleaner
go routine that will delete expired sessions on an interval that is defined when 
creating the memory session storer (cleanInterval).

### Redis

Redis sessions are stored in a Redis database. Different databases can be used
by specifying a different database ID on creation of the storer. Redis handles
session expiration automatically.

### Cookie

The cookie storer is intermingled with the CookieOverseer, so to use it you must
use the CookieOverseer instead of the StorageOverseer. Cookie sessions are stored
in encrypted form (AES-GCM encrypted and base64 encoded) in the clients browser.

## Middlewares

### Sessions Middleware

The sessions.Middleware is a required component when using this sessions package.
The sessions middleware converts the ResponseWriter that is being passed along to
your controller into a response type. This type implements the ResponseWriter
interface, but its implementation also takes care of writing the cookies in the
response objects buffer. These cookies are created as a result of creating/deleting
sessions using the sessions library.

### Sessions ResetMiddleware

When using the sessions.ResetMiddleware it will reset the expiry of the 
session of the current user regardless of whether that session is being loaded 
or modified by your controller. This ensures that users navigating to pages 
that rely on no session activation (for example a simple about page) will 
keep their session alive.

Without this middleware, for all storers except the disk storer, your session will
only reset the expiry when modifying the session (using a Set call). For the disk
storer it will refresh on both a Set and a Get because it relies on file access time.

In the majority of cases the middleware is the best user-experience, and we highly
recommend loading this middleware by default. You should load one instance of this
middleware for each session overseer you are using.

## Error types

If an API operation fails, and you would like to check if it failed due to no session
existing (errNoSession type), or the key used in a map-key operation not existing
(errNoMapKey type), you can check the error types using the following functions 
against the errors returned:

```golang
// errNoSession is a possible return value of Get and ResetExpiry
IsNoSessionError(err error) bool

// errNoMapKey is a possible return value of abcsessions.Get and abcsessions.GetFlash
// It indicates that the key-value map stored under a session did not have the 
// requested key
IsNoMapKeyError(err error) bool
```

## Examples

Using object helpers

```golang
// A session struct you've defined. Can contain anything you like.
type MySession struct {
	Username string
	AccessLevel int
	...
}

err := SetObj(o, w, r, MySession{...})
```

```golang
var mySess MySession

err := GetObj(o, w, r, &mySess)
if IsNoSessionError(err) {
	fmt.Printf("No session found")
}
```

Using object flash helpers

```golang
// A flash struct you've defined. Can contain anything you like.
type FlashError struct {
	Error string
	Code int
}

err := AddFlashObj(o, w, r, "login", &FlashError{...})
```

```golang
var errorObj FlashError

err := GetFlashObj(o, w, r, "login", &errorObj)
```

Create an Overseer

```golang
cookieOverseer := NewCookieOverseer(NewCookieOptions(), []byte("secretkeyhere"))
```

```golang
// Uses default addr of localhost:6379
storer, err := NewDefaultRedisStorer("", "", 0)
if err != nil {
	panic(err)
}

redisOverseer := NewStorageOverseer(NewCookieOptions(), storer)
```

```golang
storer, _ := NewDefaultMemoryStorer()
memoryOverseer := NewStorageOverseer(NewCookieOptions(), storer)

// Start the cleaner go routine
memoryOverseer.StartCleaner()
```

Handling unexpected errors

```golang
// o == overseer instance
val, err := abcsessions.Get(o, w, r, "coolkid")
if err != nil {
	if IsNoSessionError(err) {
		// Maybe you want to redirect? Maybe you want to create a new session?
	} else if IsNoMapKeyError(err) {
		// "coolkid" key didn't exist. Maybe you want to create it?
	} else {
		// Something crazy happened. Panic? Attempt to fix it?
		if err := o.Del(w, r); err != nil {
			panic(err)
		}
	}
}
```
