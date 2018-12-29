Changelog
#########

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
