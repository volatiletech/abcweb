### Available Session Storers

* Disk
* Memory
* Redis
* Cookie

### API Operations

Note, you should avoid mixing and matching API calls for key-value versus object.
Stick to one mode for every app you generate (either simple sessions using key-value
strings and the regular helpers or a custom sessions struct using the object helpers).

#### Key-Value string API

These session API endpoints are ideal for apps that only require simple session strings
to be stored and not a full custom sessions struct with a lot of differing variables
and types.

`Set(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, value string) error`

Set adds a key-value pair to your session. 

`Get(overseer Overseer, w http.ResponseWriter, r *http.Request, key string) (string, error)`

Get retrieves a key-value pair value stored in your session.

`Del(overseer Overseer, w http.ResponseWriter, r *http.Request, key string) error`

Del deletes a key-value pair from your session.

`AddFlash(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, value string) error`

AddFlash adds a flash message key-value pair to your session.

`GetFlash(overseer Overseer, w http.ResponseWriter, r *http.Request, key string) (string, error)`

GetFlash retrieves a flash message key-value pair value stored in your session. 
Note that flash messages are automatically deleted after being retrieved once.

#### Object API

These session API endpoints are ideal for apps that are best served with a sessions
struct opposed to individual strings. This flexibility allows you to define your
own custom sessions struct with a range of differing variable types. See [examples](#examples)
for Object API usage.

`SetObj(overseer Overseer, w http.ResponseWriter, r *http.Request, value interface{}) error`

SetObj adds any object or variable to your session.

`GetObj(overseer Overseer, w http.ResponseWriter, r *http.Request, value interface{}) error`

GetObj retrieves an object or variable stored in your session.

`AddFlashObj(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, value interface{}) error`

AddFlashObj adds an object as a flash message to your session.

`GetFlashObj(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, pointer interface{}) error`

GetFlashObj retrieves a object as a flash message stored in your session.
Note that flash messages are automatically deleted after being retrieved once.

### Overseer interface

The job of an Overseer is to interface with your storers and manage your session cookies.

`Resetter`

The Overseer interface imbeds the Resetter interface which defines the functions
used to reset the expiries of your sessions. This is handled automatically
by the ResetMiddleware and is not something that needs to be called directly.

`Set(w http.ResponseWriter, r *http.Request, value string) error`

Set creates or updates a session with a raw value.

`Get(w http.ResponseWriter, r *http.Request) (value string, err error)`

Get the raw value stored in a session.

`Del(w http.ResponseWriter, r *http.Request) error`

Delete a session and signal the browser to expire the session cookie.

`Regenerate(w http.ResponseWriter, r *http.Request) error`

Regenerate a new session ID for your session.

`SessionID(w http.ResponseWriter, r *http.Request) (id string, err error)`

SessionID returns the session ID for your session.

### Storer interface

In the case of session management, "keys" here are synonymous with session IDs.

`All() (keys []string, err error)`

All retrieves all keys in the store.

`Get(key string) (value string, err error)`

Get retrieves the value stored against a specific key.

`Set(key, value string) error`

Set a key-value pair in the store.

`Del(key string) error`

Delete a key-value pair from the store.

`ResetExpiry(key string) error`

Reset the expiry of a key in the store.

### Available Overseers

`NewStorageOverseer(opts CookieOptions, storer Storer) *StorageOverseer`

StorageOverseer is used for all server-side sessions (disk, memory, redis, etc).

`NewCookieOverseer(opts CookieOptions, secretKey [32]byte) *CookieOverseer`

CookieOverseer is used for client-side only cookie sessions.

### How does each Storer work?

#### Disk

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

#### Memory

Memory sessions are stored in memory in a mutex protected map[string]memorySession.
The key to the map is the session ID and the memorySession stores the value and expiry
of the session. The memory storer also has methods to start and stop a cleaner
go routine that will delete expired sessions on an interval that is defined when 
creating the memory session storer (cleanInterval).

#### Redis

Redis sessions are stored in a Redis database. Different databases can be used
by specifying a different database ID on creation of the storer. Redis handles
session expiration automatically.

#### Cookie

The cookie storer is intermingled with the CookieOverseer, so to use it you must
use the CookieOverseer instead of the StorageOverseer. Cookie sessions are stored
in encrypted form (AES-GCM encrypted and base64 encoded) in the clients browser.

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

### Error types

If an API operation fails, and you would like to check if it failed due to no session
existing (errNoSession type), or the key used in a map-key operation not existing
(errNoMapKey type), you can check the error types using the following functions 
against the errors returned:

```
IsNoSessionError(err error) bool
IsNoMapKeyError(err error) bool
```

### Examples

Using object helpers

```
type MySession struct {
	Username string
	AccessLevel int
	...
}

err := SetObj(o, w, r, MySession{...})
```

```
var mySess MySession

err := GetObj(o, w, r, &mySess)
if IsNoSessionError(err) {
	fmt.Printf("No session found")
}
```

Using object flash helpers

```
type FlashError struct {
	Error string
	Code int
}

err := AddFlashObj(o, w, r, "login", &FlashError{...})
```

```
var errorObj FlashError

err := GetFlashObj(o, w, r, "login", &errorObj)
```

Create an Overseer

```
cookieOverseer := NewCookieOverseer(NewCookieOptions(), []byte("secretkeyhere"))
```

```
// Uses default addr of localhost:6379
storer, err := NewDefaultRedisStorer("", "", 0)
if err != nil {
	panic(err)
}

redisOverseer := NewStorageOverseer(NewCookieOptions(), storer)
```

```
storer, _ := NewDefaultMemoryStorer()
memoryOverseer := NewStorageOverseer(NewCookieOptions(), storer)

// Start the cleaner go routine
memoryOverseer.StartCleaner()
```

