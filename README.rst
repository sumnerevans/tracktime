tracktime
=========

tracktime is a filesystem-backed time tracking solution. It uses a sane
directory structure to organize CSV files that store the tasks worked on for
each day.

Installation
------------

Using PyPi::

    pip install --user tracktime

On Arch Linux, you can install the ``tracktime`` package from the AUR. For
example, if you use ``aurman``::

    aurman -S tracktime

Dependencies
^^^^^^^^^^^^

Report functionality requires ``wkhtmltopdf`` to be installed. If you install
using the AUR package, this will be installed automatically. Otherwise, you can
install it using your distribution's package manager or visit their `homepage`__
for installation instructions specific to your operating system.

__ https://wkhtmltopdf.org/

Guiding Principles
------------------

- Filesystem based (want to be able to use Git to keep track of my time entries)
- Easy to edit manually (not a binary format)
- Must be able to use offline

Configuration Options
---------------------

There are a number of configuration options that can be set in
``~/.config/tracktime/tracktimerc``. The ``tracktimerc`` file is in YAML
format. Here is a link to an `example configuration`_. Below is a list of each
of the options and what they do.

- ``fullname`` (``string``) - your full name. This is used for generating reports.
- ``sync_time`` (``boolean``, defaults to ``false``) - determines whether or not
  to synchronise with external services.
- ``gitlab_username`` (``string``) - your GitLab username. Used to sync with GitLab.
- ``gitlab_api_key`` (``string``) - your GitLab API Key. Must be created with
  full API access. Used to sync with GitLab.
- ``tableformat`` (``string``, defaults to ``simple``) - the type of table to
  generate when exporting a report to stdout. (See the `tabulate documentation`_
  for details on what formats are supported.)
- ``project_rates`` (``dictionary``) - a dictionary of project-rate pairs. Used
  to calculate totals for the report export.
- ``customer_aliases`` (``dictionary``) - a dictionary of alias-full name
  pairs. Used to expand a name on the report export. Useful when a customer has
  a really long name.
- ``customer_addresses`` (``dictionary``) - a dictionary of name-address
  pairs. Used in the report export.

.. _example configuration: https://gitlab.com/sumner/tracktime/snippets/1731133
.. _tabulate documentation: https://bitbucket.org/astanin/python-tabulate#rst-header-table-format

Architecture
------------

Directory Structure
^^^^^^^^^^^^^^^^^^^

::

    /<root>
    |-> 2017
    |   |-> 01
    |   |   |-> .synced
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

The ``.synced`` file in each month's directory stores the amount of time that
has been reported to the external services.

Time Tracking File Format
^^^^^^^^^^^^^^^^^^^^^^^^^

All time tracking files will be CSVs with the following fields:

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

Synced Time File Format
^^^^^^^^^^^^^^^^^^^^^^^

All ``.synced`` files will be CSVs with the following fields:

- ``type`` - the type of taskid (gitlab, github, or none)
- ``project`` - the project that the taskid is associated with
- ``taskid`` - the task ID (issue/PR/MR/story number)
- ``synced`` - the amount of time that has been successfully pushed to the
  external service for this taskid

Synchronising to External Services
----------------------------------

tracktime can sync tracked time with external services. It does this by keeping
track of how much time it has been reported to the external service using the
``.synced`` file in each month's directory. Then, it pushes changes to the
external service.

**This is not a two-way sync! tracktime only pushes changes, it does not poll
for changes to the external services.**

Unsupported Edge Cases
----------------------

- Daylight savings time (if you are needing to track time at 02:00 in the
  morning, I pitty you)
- Time entries that span multiple days (if you are working that late, create two
  entries)
- Timezones (only switch timezones between days, if you have to switch, just
  make sure that you keep the timezone consistent for a given day)
