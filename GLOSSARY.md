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
