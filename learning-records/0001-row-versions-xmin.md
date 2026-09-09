# Row versions are real: UPDATE writes new versions, old ones linger until vacuum

The user ran the xmin/xmax experiment (blog/experiments/0001-xmin.txt) and correctly interpreted unprompted that an UPDATE assigned a new xmin (751 → 752) to freshly written row versions, that the old versions are not deleted but remain stored (inferring table growth from n to ~n*2), and that vacuum is what eventually clears them. They correctly noticed xmax stayed 0 and honestly flagged it as unknown — resolved in session: xmax=752 lives on the old (now-invisible) versions; the new versions have xmax=0 until something replaces them.

## Evidence
- Transcript: post-UPDATE select showed xmin=752 on all rows (single updater transaction) with xmax=0; user's written inference about stored old versions and vacuum.
- Honest disclosure of the xmax=0 gap rather than confabulation.

## Implications
- Floor is set: MVCC copy-on-write and dead tuples are established. Lesson 0002 (snapshots/visibility, why an open transaction stalls vacuum) is the natural next step, not a re-teach of row versions.
- The xmax confusion (whose header gets stamped) is a predicted stumbling block for Episode 2's long-transaction work — reinforce there, and promote `xmax` to the glossary only after the user explains the two-generation visibility themselves.
