tracktime
=========

.. image:: https://builds.sr.ht/~sumner/tracktime.svg
   :alt: Build Status
   :target: https://builds.sr.ht/~sumner?search=%7Esumner%2Ftracktime
.. image:: https://img.shields.io/pypi/v/tracktime?color=4DC71F&logo=python&logoColor=fff
   :alt: PyPi Version
   :target: https://pypi.org/project/tracktime/
.. image:: https://img.shields.io/aur/version/tracktime?logo=linux&logoColor=fff
   :alt: AUR Version
   :target: https://aur.archlinux.org/packages/tracktime/
.. image:: https://img.shields.io/liberapay/receives/sumner.svg?logo=liberapay
   :alt: LiberaPay Donation Status
   :target: https://liberapay.com/sumner/donate

tracktime is a filesystem-backed time tracking solution. It uses a sane
directory structure to organize CSV files that store time tracking data for each
day.

Features
--------

- CLI
- Start/stop/resume time entries
- List/edit time entries for a given day
- Generate rST, PDF, HTML reports for arbitrary date ranges (optionally
  restricted to a particular customer or project)
- Synchronise time spent to GitLab using the Time Tracking API

Installation
------------

Using PyPi::

    pip install --user tracktime

On Arch Linux, you can install the ``tracktime`` package from the AUR. For
example, if you use ``yay``::

    yay -S tracktime

Dependencies
^^^^^^^^^^^^

Report functionality requires ``wkhtmltopdf`` to be installed. If you install
using the AUR package, this will be installed automatically. Otherwise, you can
install it using your distribution's package manager or visit their `homepage`__
for installation instructions specific to your operating system.

Additionally, you will need to ensure that the ``wkhtmltopdf`` executable is in
your ``$PATH``.

__ https://wkhtmltopdf.org/

Guiding Principles
------------------

- Filesystem based (want to be able to use Git to keep track of my time entries)
- Easy to edit manually (not a binary format)
- Must be able to use offline

Configuration Options
---------------------

There are a number of configuration options that can be set in
``~/.config/tracktime/tracktimerc``. The ``tracktimerc`` file is in YAML format.
Here is a link to an `example configuration`_. Below is a list of each of the
options and what they do.

- ``fullname`` (``string``) - your full name. This is used for generating
  reports.
- ``sync_time`` (``boolean``, defaults to ``false``) - determines whether or not
  to synchronise with external services.
- ``editor`` (``string``) - specifies the editor to use when ``tt edit`` is run.
  If this value is not present, the ``EDITOR`` and ``VISUAL`` environment
  variables are used as fallback. If none are present, then ``vim`` (on
  non-Windows OSes) or ``notepad`` (on Windows) is used.
- ``editor_args`` (``string``) - a comma separated list of arguments that should
  be passed to the ``editor`` when ``tt edit`` is run.
- ``gitlab`` (``dictionary``) - configuration of GitLab parameters

  - ``api_root`` (``string``, defaults to ``'https://gitlab.com/api/v4/'``) -
    the GitLab API root to use.
  - ``api_key`` (``string``) - can be either your GitLab API Key in plain text
    or a shell command which returns the API key. This second option can be
    useful if you want to store your API key in a password manager. To indicate
    that it is a shell command, append a vertical bar (``|``) at the end of the
    command.

    .. note::

      You can create an API key here:
      https://gitlab.com/profile/personal_access_tokens. The API Key must be
      created with full API access. Used to sync with GitLab.

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
- ``external_synchroniser_files`` - a dictionary of
  ``synchroniser name -> synchroniser Python file``. Allows users to import
  third party synchronisers.
- ``day_worked_min_threshold`` - the number of minutes which must be worked in a
  day to consider it a work day. This is to avoid days where you work for a few
  minutes from skewing statistical results.

.. _example configuration: https://git.sr.ht/~sumner/tracktime/tree/master/examples/tracktimerc
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

Supported External Services
^^^^^^^^^^^^^^^^^^^^^^^^^^^

- GitLab

Contributing
------------

See the CONTRIBUTING.md_ document for details on how to contribute to the
project.

.. _CONTRIBUTING.md: https://git.sr.ht/~sumner/tracktime/tree/master/CONTRIBUTING.md

Unsupported Edge Cases
----------------------

- Daylight savings time (if you are needing to track time at 02:00 in the
  morning, I pitty you).
- Time entries that span multiple days (if you are working that late, create two
  entries).
- Timezones (only switch timezones between days, if you have to switch, just
  make sure that you keep the timezone consistent for a given day).
