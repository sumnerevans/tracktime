Changelog
#########

0.9.10
======

- **Added config item:** ``editor_args`` - a comma separated list of arguments
  that should be passed to the ``editor`` when ``tt edit`` is run.
- **Reporting Improvements:**

  The main thing that was added is the ability to report time spent on
  individual tasks and descriptions.

  - **Breaking API change:** To specify a report output file, you now have to
    use the ``-o``/``--outfile`` argument.
  - Ability to report on both the project and customer gains at the same time.
  - By default, if you are reporting on 7 days or less, it will show both the
    task and description grains. If you are reporting on a month or less, it
    will show you the task grains. If you are reporting on longer than a month,
    it will show you neither of those grains.

    You can override the defaults using the ``--(no-)taskgrain`` and
    ``--(no-)descriptiongrain`` parameters on ``tt report``.
  - By default, the entries are sorted by time spent, descending. You can also
    sort alphabetically. You can explicitly specify the sort using the
    ``--sort`` argument. You can specify ``--asc``/``-a`` or ``--desc``/``-d``
    to force the sort to be ascending or descending, respectively.

- ``tt list`` now shows the entry numbers beside the entries for easier resume
  of previous tasks.

0.9.9
=====

- Windows Support!
- Allow for entry types other than GitLab and GitHub
- **Added config item:** ``editor`` - a string which specifies the editor to use
  when ``tt edit`` is run.
- Added default ``fullname`` config so that ``tt report`` works without setup.
- **Functionality change:** Added customer-specific billing rates. If a
  reporting group has both customer and project billing rates, the project rate
  is used.

0.9.8
=====

- Fixed bug preventing reporting on projects.
- Fixed bug where the GitLab synchroniser would try and sync GitHub entries.
- Fixed a few help formatting issues.
- Fixed documentation in README.

0.9.7
=====

- **License change:** Migrated from MIT to GPLv3. Positive in the Freedom
  Dimension, so to speak.
- **Deprecation Warning:** GitLab configuration moved to nested dictionary. See
  the new configuration example:
  https://gitlab.com/sumner/tracktime/snippets/1731133.
- Allowed resume across days.
- Better error message when trying to make a report with unended time entries.

0.9.6
=====

- Performance fix: configuration cached instead of reloaded every single time
  from disk.
- Added ``-v``/``--version`` flag to show version of the program.

0.9.5
=====

- Ability to report on projects
- Allow GitLab API Key config item to be an arbitrary shell command
- Added better logging for synchronizing time entries

0.9.4
=====

- Ability to resume time entries before the previous
- Added lots of unit tests
- Added code coverage statistics

0.9.3
=====

- **Emergency Bugfix:** Added the ``tracktime.synchronisers`` package by
  converting to use ``find_packages`` instead of hard-coding a list of packages.

0.9.2
=====

- **Emergency Bugfix:** Removed the ``flake8`` and ``flake8-pep3101``
  dependencies

0.9.1
=====

- **Bug Fix:** Added missing ``pyyaml`` dependency
- **Bug Fix:** ``tracktime`` no longer blows up when
  ``~/.config/tracktime/tracktimerc`` does not exist

- Changed Development Status to "Beta"
- Improved build process to include linting
- Moved ``edit`` functionality out to the CLI (#14)
- Added report export to reStructuredText
- Added a bunch of unit tests for critical code
- **Refactor:** pulled the GitLab synchroniser out to its own module and created
  a ``synchronisers`` module.
