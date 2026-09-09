# pg-lab Glossary

Canonical language for PostgreSQL internals in this teaching workspace. Definitions are added **only after** the user has used a term correctly in a lesson or experiment — this is a record of compressed knowledge, not a dictionary to learn from.

## Terms

### Row versions & storage

**Row version (tuple)**:
One physical copy of a row in the heap, created by an INSERT or UPDATE; tables are collections of versions, not of rows.
_Avoid_: record, entry

**xmin**:
The transaction ID that inserted a row version, stored in the version's header.
_Avoid_: created-by, version id

**Dead tuple**:
A row version that no snapshot can see anymore but which still occupies heap space until VACUUM reclaims it.
_Avoid_: deleted row, garbage row

**VACUUM**:
The process that removes dead tuples, making their space reusable; autovacuum does this automatically in the background.
_Avoid_: cleanup job, defrag

### Snapshots & visibility

**xmax**:
The transaction ID that replaced or deleted a row version, stamped on the *old* version's header; 0 while the version is still current.
_Avoid_: deleted-by, expiry

**Snapshot**:
The list of transactions still in progress when a statement starts; by comparing each version's xmin/xmax against it, Postgres decides what that statement can see.
_Avoid_: view of data, read state

**backend_xmin**:
The oldest transaction ID a backend's snapshot might still need — its horizon floor; vacuum may not remove anything dead-at-or-after this value.
_Avoid_: session xid, lock value

**idle in transaction**:
A session state where a transaction is open but the client is sending no queries; its backend_xmin still pins the vacuum horizon, so it silently blocks dead-tuple cleanup.
_Avoid_: hanging session, stuck query
