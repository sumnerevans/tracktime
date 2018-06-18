tracktime
=========

Time Tracking solution using CSVs.

Guiding Principles
------------------

- Filesystem based (want to be able to use Git to store this)
- Easy to edit manually (not a binary format)

Fields
------

- ``start`` - the start time for the time entry
- ``stop`` - the stop time for the time entry
- ``type`` - the type of entry (gitlab, github)
- ``task`` - the task ID (issue number)
- ``customer`` - the customer the entry is for
- ``notes`` - any notes about the time entry

Synchronisation with External Services
--------------------------------------
