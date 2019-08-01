# GoCrack Changelog

## 2.1

New Features:

* Update Go to 1.12.6
* Support for Go Modules
* Updated all server dependencies
* Dropped support for Hashcat 3.6 for 5.1

## 2.0

New Features:

* Multiple cracking engines: GoCrack can easily leverage multiple password cracking engines simply by building a wrapper around the engine interface
* Multiple authentication providers: GoCrack can use third-party authentication providers and ships with LDAP and database backends
* Multiple database backends: Use whatever database you want! (as long as a backend plugin has been written). Ships with a boltdb (flatfile, in-memory) backend
* Improved Notifications: Emails will be sent out whenever a task you created has had a state change. Browser Notifications are now supported as well
* File Manager: Manage both task and engine files from the Web UI
* New Web UI: Refactored the UI to streamline task creation and added support for pagination and searching of tasks
* Auditing: Viewing passwords, task info, and modifications are now logged and viewable to administrators
* Entitlements: Access to tasks and files is controlled by a new entitlement system

Changes:

* Switched from a custom PUB/SUB framework to HTTP/HTTP2 for Server<->Worker comms
* Improved the build process by removing hardcoded references to CUDA and added build tags to customize the included features in the binary
* Improved support for debugging and logging to terminal
    * Added an embedded, configurable pprof for the worker and it's children
    * Pretty Print messages to stdout/sterr if running in an interactive terminal
* Improved test coverage
* Moved gocrack integration to a separate library, gocat
* One worker per server. In version 1.X, it was practice to run a worker per GPU incase you had N jobs each using 1 GPU. Now, only 1 worker is needed and will spawn processes as needed
* Improved the number of metrics exported by the server & worker

## 0.0 - 1.0

Internal releases, code not available.