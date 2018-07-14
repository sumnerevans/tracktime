Changelog
#########

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
