# Install Considerations

1. Pick your database implementation
    * Flatfile via BoltDB
        1. Pros
            * No dependency on a database server
            * Fast with small/medium data sets. Good for one off instances.
        1. Cons
            * May not scale well on large deployments
            * Memory consumption may be higher
            * Startup time might be higher
        1. Mixed
            * Backup is easy as a single file needs to be copied over. However to do this safely, you'd need to stop the server. We'll add a hook to create a backup via a signal in a future version.

1. Pick your authentication plugin
    * Database
        1. Pros
            * Separate credentials from your LDAP/corporate infrastructure
            * No dependency on an external auth server
        2. Cons
            * No centralized authentication
    * LDAP
        1. Pros
            * Centralized Authentication
        2. Cons
            * Depending on the location of your GoCrack server, allowing LDAP access might not be feasible
