tracktime
=========

Time Tracking solution using CSVs.

Guiding Principles
------------------

- Filesystem based (want to be able to use Git to keep track of my time entries)
- Easy to edit manually (not a binary format)

Protocol
--------

Directory Structure
^^^^^^^^^^^^^^^^^^^

::

    /<root>
    |-> 2017
    |   |-> 01
    |   |   |-> 01
    |   |   |-> 02
    |   |   |-> ...
    |   |-> 02
    |   |-> ...
    |-> 2018

In other words, the generic path is ``YEAR/MONTH/DAY`` where all three fields
are the numeric, zero-padded.

Each day with time tracked will have a corresponding file and have the file
format as described below.

File Format
^^^^^^^^^^^

The file will be a CSV with the following fields:

- ``start`` - the start time for the time entry
- ``stop`` - the stop time for the time entry
- ``type`` - the type of entry (gitlab, github)
- ``task`` - the task ID (issue number)
- ``customer`` - the customer the entry is for
- ``notes`` - any notes about the time entry

Synchronisation with External Services
--------------------------------------
