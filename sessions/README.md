### Available Session Storers

* Disk
* Memory
* Redis
* Cookie

### API Operations

### How do they work?

#### Disk

Disk sessions store the session as a text file on disk. By default they store in 
the systems temp directory under a folder that is randomly generated when you 
generate your app using abcweb app generator command. The file names are the UUIDs 
of the session. Each time the file is accessed (using Get, Put, or manually on 
disk) it will reset the access time of the file, which will push back the 
expiration defined by maxAge.

### RefreshExpiry middleware

When using the sessions.RefreshExpiry middleware it will reset the expiry of the 
session of the current user regardless of whether that session is being loaded 
or modified by your controller. This ensures that users navigating to pages 
that rely on no session activation (for example a simple about page) will 
keep their session alive.

Without this middleware, for all storers except the disk storer, your session will
only reset the expiry when modifying the session (using a Put call). For the disk
storer it will refresh on both a Put and a Get because it relies on file access time.

In the majority of cases the middleware is the best user-experience.
