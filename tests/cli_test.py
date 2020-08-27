import subprocess
from subprocess import check_output

import pytest


def test_cli():
    # TODO (#21): validate that the output makes sense
    check_output('tt'.split()).decode()

    # '\n'.join([
    #         'Entries for ...',
    #         '======================',
    #         '...',
    #         'Total: ...',
    #     ]))

    check_output('tt --help'.split()).decode()
    check_output('tt sync'.split()).decode()

    with pytest.raises(subprocess.CalledProcessError):
        check_output('tt sync 2018'.split()).decode()
