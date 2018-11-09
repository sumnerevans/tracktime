Contributing
############

Contributions are welcome! Please submit a MR.

Code Style
==========

`PEP-8`_ is to be followed **strictly**; however line lengths of up to 90 or
100 columns is OK if the purpose of the line is still cleanly conveyed.

.. _`PEP-8`: https://www.python.org/dev/peps/pep-0008/

The CI process uses Flake8 to lint the Python code. You can check the linting
locally by installing ``flake8`` and ``flake8-pep3101`` from ``pip``::

    pip install flake8 flake8-pep3101

Then you can run::

    flake8

to see all the linter errors.
