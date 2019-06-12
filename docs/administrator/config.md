# Configuration Settings

GoCrack uses YAML based configuration files for it's Server and Workers. By default, the server looks for its config in `configs/server.yaml` and the worker looks in `configs/worker.yaml`. You can change the path via the `-config` flag on both binaries.

If you're setting up GoCrack for the first time, you may find the [installation considerations](install_considerations.md) list handy.

Configuration File Examples:

1. [Server](../default_server_config.yaml)
1. [Worker](../default_worker_config.yaml)

We **highly recommend** that you enable SSL for **all** communications. The only reason we have `ssl_enabled` options in the config is for disabling it if you're proxying requests through a webserver like nginx or apache2.

Example Configs for NGINX:

1. [Web Listener](../example-web.nginx.conf)
1. [RPC/Worker Listener](../example-rpc.nginx.conf)

# Caveats

1. At this time there is no way to migrate from storage backends. If you change one mid deployment, you will be starting from scratch.
1. Switching authentication drivers may have unattended consequences. If you go from database to LDAP or vice versa, you may find duplicate user records in the system.

## Server

### Top Level Options

    ---
    debug: optional bool (false by default)

1. `debug` Will enable a prettier output from GoCrack's logger as well as debug messages and information about the loaded GIN routes.

### Web Server (Server <-> Browser/Clients)

    web_server:
        listener:
            address: string (required)
            ssl_certificate: string (optional)
            ssl_private_key: string (optional)
            ssl_ca_certificate: string (optional)
            ssl_enabled: bool (optional)
        cors:
            allowed_origins: list of strings (optional)
        ui:
            static_path: string (optional)
            csrf_key: string (required if enabled)
            csrf_enabled: bool (optional)

1. `listener`
    * `address`: The FQDN or IP address with optional port where the API endpoint & UI should listen on. Example: `gocrack.local:1337`
    *  `ssl_certificate`: The SSL certificate that GoCrack should use.
    * `ssl_private_key`: The SSL private key for the certificate
    * `ssl_ca_certificate`: The SSL CA certificate
    * `ssl_enabled`: Indicates if GoCrack should use a TLS listener
1. `cors`
    * `allowed_origins`: A list of domains (origins) in which requests will be made from (UI). This is a [CORS]   (https://developer.mozilla.org/en-US/docs/Web/HTTP/Access_control_CORS) setting.
    * `max_preflight_age`: Indicates how long the response of a preflight request can be cached by the browser. By default, this is set to 24 hours. It must be in the format of a [duration string](https://golang.org/pkg/time/#ParseDuration)
1. `ui`
    * `static_path`: Path containing `index.html` and a folder called `static` that serve as the GoCrack User Interface.  The reference user interface can be found in the [gocrack-ui repository](https://github.com/fireeye/gocrack-ui).
    * `csrf_key`: A secure key that is used to sign the CSRF cookies. This should be set to a strong, random string
    * `csrf_enabled`: By default, this is true but on development instances this should be set to false.

### RPC (Server <-> Worker)

**Note**: For documentation on how to generate certificates for Workers, see [Worker Authentication](worker_authentication.md).

    rpc:
        listener:
            address: string (required)
            ssl_certificate: string (optional)
            ssl_private_key: string (optional)
            ssl_ca_certificate: string (optional)
            ssl_enabled: bool (optional)

1. `listener`
    * `address`: The FQDN or IP address with optional port where the RPC endpoint should listen on. Example: `rpc.gocrack.local:1338`
    * `ssl_certificate`: The SSL certificate that GoCrack should use.
    * `ssl_private_key`: The SSL private key for the certificate
    * `ssl_ca_certificate`: The SSL CA certificate
    * `ssl_enabled`: Indicates if GoCrack should use a TLS listener

### Database

    database:
        backend: string (required)
        connection_string: string (required)

1. `backend`: The database driver/backend to use for server storage. Current options are:
    * `bdb`: A flatfile database using BoltDB, a clone of LMDB

1. `connection_string`: The database connection string will vary depending on which driver you picked with (see the installation guide for supported options).
    * If you're using flatfile backend - it will be the path where the `.db` file is created. Example: `/opt/gocrack/storage.db`

### File Manager

    file_manager:
        task_file_path: string (required)
        engine_file_path: string (required)
        temp_path: string (required)
        task_max_upload_size: int (optional)
        import_path: string (optional)

1. `task_file_path`: The full path where you'd like to store task files (uncracked hashes).
1. `engine_file_path`: The full path where dictionaries, password masks, mangling rules, etc live
1. `temp_path`: The full path where temporary files are saved. It must live on the same volume as task and engine paths.
1. `task_max_upload_size`: The maximum size in bytes for task files. The default is `20mb`
1. `import_path`: Full path to a folder where shared files can be imported from. Sending a USR1 signal to the server process will trigger an import of all files from this directory

### Authentication

    authentication:
        backend: string (required)
        backend_settings: varies based on backend (required)
        token_expiry: string (optional)
        secret_key: string (required)

1. `backend`: The database backend you'd like to use for authentication. It's value will very depending on which authentication scheme you are going with.

    * LDAP: `ldap`
    * Database: `database`

1. `backend_settings`: The settings for the backend you selected.
    1. Database:
        * `bcrypt_cost`: The cost associated with generating a bcrypt hash of a users password. The value must be between `4` and `31`. It's default value is `10`. Setting this higher will generate increase CPU load on authentication.
    1. LDAP
        * `address`: The address & port of the LDAP server
        * `base_dn`: The distinguished name of an OU or even the base domain controller. Examples:

            1. `dc=myawesomedomain,dc=com`
            1. `ou=people,dc=myawesomedomain,dc=com`
        * `bind_dn`: The distinguished name of a **READ ONLY** user the system uses to bind to the LDAP server and search for users in.
        * `bind_password`: The password of the **READ ONLY** user the system uses to search for the user in
        * `root_ca`: The certificate of the LDAP server for verification
1. `token_expiry`: A [duration string](https://golang.org/pkg/time/#ParseDuration) that indicates how long a generated JWT is valid for in user authentication. It must be a positive duration and it's default value is 1 day. Examples:
    * `5h`
    * `60m`
    * `1d`
1. `secret_key`: A secure string that is used to sign the JWT's. It should be long (40+ alpha num.) and complicated.

### Notifications (Email)

    email_server:
        address: string (required)
        port: int (required)
        username: string (optional)
        password: string (optional)
        skip_invalid_cert: bool (optional)
        certificate: string (optional)
        server_name: string (optional)
    enabled: bool (required)
    from_address: string (required)
    public_address: string (required)

1. `email_server.address`: The FQDN or IP address of the mail server
1. `email_server.port`: The port of the mail server
1. `email_server.username`: An optional username to authenticate to the mail server with
1. `email_server.password`: The password to the optional user account
1. `email_server.skip_invalid_cert`: Skip validation of the SSL certificate on the mail server (false by default)
1. `enabled`: Should the notification be enabled? false by default
1. `from_address`: The email address to send notifications from
1. `public_address`: The FQDN of the GoCrack UI to embedd in notifications

## Worker

### Top Level Options

    ---
    engine_debug: bool (optional)
    save_task_file_path: string (required)
    save_engine_file_path: string (required)
    auto_cpu_assignment: bool (optional)

1. `engine_debug`: When set to true, the worker process will echo stdout and stderr from the child processes
1. `save_task_file_path`: The path where task files should temporarily be saved to
1. `save_engine_file_path`: The path where engine files should temporarily be saved to
1. `auto_cpu_assignment`: When set to true, the worker will automatically assign tasks to CPUs if no GPUs are available

### Server

    server:
        connect_to: string (required)
        ssl_certificate: string (required)
        ssl_ca_certificate: string (required)
        ssl_private_key: string (required)
        server_name: string (optional/required)

1. `connect_to`: The address/FQDN and port of the GoCrack RPC server.
1. `ssl_certificate`: The SSL certificate for the worker for mutual authentication
1. `ssl_ca_certificate`: The CA certification that signed the worker & server certificates for validation purposes
1. `ssl_private_key`: The private key for the SSL certificate
1. `server_name`: The FQDN/Subject Alternative Name on the Server Certificate. This will most likely be required for
validation purposes

### Intervals

*Note: All strings in this section must be parseable by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration)*

    intervals:
        beacon: string (optional)
        job_status: string (optional)
        termination_delay: string (optional)

1. `beacon`: This sets how frequently we beacon in to the GoCrack server. By default it's every 30 seconds.
1. `job_status`: This sets how frequently we return the engine status to the GoCrack server. By default it's every 5 seconds.
1. `termination_delay`: This sets how long the GoCrack process will wait for a graceful stop of the cracking engine when a stop was requested. By default it waits 30 seconds before sending a SIGKILL to the process.

### Hashcat Engine

    hashcat:
        log_path: string (required)
        potfile_path: string (optional)
        session_path: string
        shared_path: string

1. `log_path`: The path where hashcat can save log files for tasks at
1. `potfile_path`: The path where hashcat will save the potfile (list of previously cracked passwords)
1. `session_path`: The path where hashcat will save the checkpoint/restore files at
1. `shared_path`: The path where hashcat's shared files exist. This will most likely be `/usr/local/share/hashcat`.

### Device Assignment Settings

These values set the max # of GPUS that will be assigned to a task based on priority (if they are free).

    gpus_priority_limit:
        high: int (optional)
        normal: int (optional)
        low: int (optional)

1. `high`: Max # of GPUs assigned, if free to a high priority task.
1. `normal`: Max # of GPUs assigned, if free to a normal priority task.
1. `low`: Max # of GPUs assigned, if free to a low priority task.
