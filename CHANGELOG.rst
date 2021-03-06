v0.10.0
=======

* Add Linear Synchroniser for making rich reports for Linear issues.
* Statistics now define a week as any set of five weekdays worked rather than
  any set of five days.
* Fixed bug with Sourcehut synchroniser where the month not correctly
  zero-padded.
* Internal changes:

  * Use ``ABC`` and ``abstractmethod`` for the ``ExternalSynchroniser`` class to
    make it easier to to know what needs to be implemented on a new
    synchroniser.

v0.9.20
=======

* Add GitHub Synchroniser for making rich reports for GitHub issues/PRs
* Add option ``report_statistics`` to config to enable/disable statistics
  reporting.
* Removed ``docutils`` from project dependencies since it is a dev dependency.

v0.9.19
=======

* Fixed bug with average time per week worked calculation when weeks worked < 1.
* Fixed bug with how the GitLab synchroniser was passing arguments to
  ``requests``.

* INFRASTRUCTURE

  * Convert README to Markdown
  * Convert ``build.yaml`` to use ``complete-build``

v0.9.18
=======

* Critical bug fix for sourcehut synchroniser preventing adding a new comment
  with the time data.

v0.9.17
=======

* Fixed an issue with the sourcehut synchroniser where time reporting on an
  issue was not correct when time was tracked across multiple months. (#18)

* INFRASTRUCTURE

  * Use ``poetry publish`` instead of twine for deploying to PyPi.

v0.9.16
=======

* Small bugfix for sourcehut synchroniser when time entries were specified in
  different but equivalent ways.
* Added an option to the ``jira_selenium_chrome.py`` example to specify the
  Chrome binary file.

* INFRASTRUCTURE

  * Converted to use ``poetry`` instead of ``setup.py`` for handling building.

v0.9.15
=======

* Use ``use nix`` with a ``shell.nix`` for consistent development environments.
* Fixed link to example configuration and brought it in to the repository.
* Added a synchroniser for sourcehut. [1]_

.. [1] https://sr.ht

v0.9.14
=======

* Fixed dependency issue where `argcomplete` was required as an
  `install_dependency`.
* Fixed issue with importing `importlib.util`.

* INFRASTRUCTURE

  * Fleshed out `CONTRIBUTING.md` document to help onboard contributors.
  * Migrated to sr.ht because of usability regressions in GitLab.
  * Migrated to using `black` for formatting instead of `yapf`.
  * Migrated from Pipenv to Poetry because Poetry is actually fast.
  * Added an `.envrc` and a `.editorconfig` file to help create consistent
    development environments for contributors.

v0.9.13
=======

* When listing entries for a day, you can now only show entries for a specific
  customer  using the ``-c``/``--customer`` parameter.
* Added statistics to the report output.
* The GitLab adapter now caches names of the task in a pickle store.

v0.9.12
=======

* **Added config item:** ``external_synchroniser_files`` - a dictionary of
  ``synchroniser name -> synchroniser Python file``. Allows users to import
  third-party synchronisers.
* **Added CLI parameter:** ``--config`` - the configuration file to use.
  Defaults to ``~/.config/tracktime/tracktimerc``.
* Added reference third-party synchroniser to JIRA (|jira_example|_)
* **Synchroniser API changes:**

  * New *required* function: ``get_name`` returns the human-understandable name
    for the synchroniser.
  * New *optional* functions: ``get_formatted_task_id``, ``get_task_link``, and
    ``get_task_description`` for getting a human-understandable task ID, task
    link, and task description, respectively, from the external service for
    reporting (see below).

* **Reporting Changes:**

  * If a Synchroniser implemented ``get_formatted_task_id`` for the task type,
    the formatted version will be used in reporting. For example, the reference
    JIRA synchroniser formats tasks as <project>-<task_id>.
  * If a Synchroniser implemented ``get_task_description`` for the task type,
    the description of the task will be shown in the report.
  * If a Synchroniser implemented ``get_task_link`` for the task type, when
    exporting to HTML or PDF, the task description will be a hyperlink to the
    task in the external service.

* **Infrastructure:**

  * integrated ``mypy`` into linting pipeline and added CoC settings for people
    using coc.nvim_.
  * Converted to use Pipenv for everything.
  * Started adding more integration tests, especially for reporting.

* Note: ``v0.9.11`` was skipped due to a deploy failure.

.. _coc.nvim: https://github.com/neoclide/coc.nvim
.. |jira_example| replace:: ``examples/jira.py``
.. _jira_example: https://gitlab.com/sumner/tracktime/blob/master/examples/jira.py

v0.9.10
=======

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

v0.9.9
======

- Windows Support!
- Allow for entry types other than GitLab and GitHub
- **Added config item:** ``editor`` - a string which specifies the editor to use
  when ``tt edit`` is run.
- Added default ``fullname`` config so that ``tt report`` works without setup.
- **Functionality change:** Added customer-specific billing rates. If a
  reporting group has both customer and project billing rates, the project rate
  is used.

v0.9.8
======

- Fixed bug preventing reporting on projects.
- Fixed bug where the GitLab synchroniser would try and sync GitHub entries.
- Fixed a few help formatting issues.
- Fixed documentation in README.

v0.9.7
======

- **License change:** Migrated from MIT to GPLv3. Positive in the Freedom
  Dimension, so to speak.
- **Deprecation Warning:** GitLab configuration moved to nested dictionary. See
  the new configuration example:
  https://gitlab.com/sumner/tracktime/snippets/1731133.
- Allowed resume across days.
- Better error message when trying to make a report with unended time entries.

v0.9.6
======

- Performance fix: configuration cached instead of reloaded every single time
  from disk.
- Added ``-v``/``--version`` flag to show version of the program.

v0.9.5
======

- Ability to report on projects
- Allow GitLab API Key config item to be an arbitrary shell command
- Added better logging for synchronizing time entries

v0.9.4
======

- Ability to resume time entries before the previous
- Added lots of unit tests
- Added code coverage statistics

v0.9.3
======

- **Emergency Bugfix:** Added the ``tracktime.synchronisers`` package by
  converting to use ``find_packages`` instead of hard-coding a list of packages.

v0.9.2
======

- **Emergency Bugfix:** Removed the ``flake8`` and ``flake8-pep3101``
  dependencies

v0.9.1
======

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
