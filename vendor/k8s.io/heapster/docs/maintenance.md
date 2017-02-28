Heapster Maintenance Procedures
===============================

Triage
------

Triage is done on a weekly rotation by members of the
@kubernetes/heapster-maintainers team.  Initially this will be limitted to
the following people:

- @piosz
- @directxman12

The on-duty triage maintainer is responsible for initial triage of bugs
and pull requests (labeling, assigning, etc), initially responding to
issues, merging pull requests approved by sink owners, and reviewing
non-sink-specific pull requests.

### Labels ###

Each issue and pull request should be assinged one of:

- bug
- enhancement
- question
- support
- testing
- invalid
- docs

A priority label should also be assigned (`Priority/P[0-3]`, with `Priority/P0`
being the highest priority) for `bug` and `enhancement` issues and pull
requests.

Duplicate bugs should be tagged with `duplicate`, and should have a comment
referencing the main bug.  `wontfix` may be applied as appropriate for issues
which won't be fixed.

Additionally, a sink label should be applied to sink-related issues and pull
requests (`sink/$SINK_NAME`).

Sink Maintenance
----------------

Each sink will have a set of one or more owners who are responsible for
responding to issues and pull requests, once triaged.  If a sink has no owners,
a call will be put out for owners, and if none are found, the sink may be
subject to deprecation and removal after a release.

See the [sink owners](sink-owners.md) reference file for more information.

Releases
--------

Releases will be performed by @piosz.  Any issues about releases should be
assigned to him.
