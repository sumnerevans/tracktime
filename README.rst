tracktime
=========

tracktime is a filesystem-backed time tracking solution. It uses a sane
directory structure to organize CSV files that store the tasks worked on for
each day.

Installation
------------

Report functionality requires ``wkhtmltopdf`` to be installed. Install it using
your distribution's package manager or visit their `homepage`__ for installation
instructions specific to your operating system.

Guiding Principles
------------------

- Filesystem based (want to be able to use Git to keep track of my time entries)
- Easy to edit manually (not a binary format)
- Must be able to use offline

Structure
---------

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
- ``project`` - the project for the time entry
- ``type`` - the type of entry (gitlab, github, or none)
- ``taskid`` - the task ID (issue/PR/MR/story number)
- ``customer`` - the customer the entry is for
- ``notes`` - any notes about the time entry

The ``start`` and ``stop`` fields will be times, formatted in ``HH:MM`` where
``HH`` is 24-hour time. All other fields are text fields that can hold arbitrary
data.

Synchronisation with External Services
--------------------------------------

.. TODO
- How to deal with time entries that are not done through tracktime?
- How to deal with time entries that are edited outside of tracktime?
- How to determine whether or not a time entry has been added or not?
- What if there's an entry that intersects with the one that needs to be added?

Unsupported Edge Cases
----------------------

- Daylight savings time (if you are needing to track time at 02:00 in the
  morning, I pitty you)
- Time entries that span multiple days (if you are working that late, create two
  entries)
- Timezones (only switch timezones between days, if you have to switch, just
  make sure that you keep the timezone consistent for a given day)
