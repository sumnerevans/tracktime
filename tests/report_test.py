from pathlib import Path
from subprocess import check_output

import pytest


@pytest.fixture(autouse=True)
def dummy_config():
    # Create a dummy configuration file
    config_directory = Path('~/.config/tracktime')
    config_directory.mkdir(parents=True, exist_ok=True)
    config_path = config_directory.joinpath('tracktimerc')
    if not config_path.exists():
        with open(config_path, 'w+') as cf:
            cf.writelines(f'{x}\n' for x in [
                'fullname: Sumner Evans',
                'sync_time: true',
                'gitlab_username: sumner',
                'tableformat: fancy_grid',
            ])


def test_report():
    # TODO validate that the output makes sense
    check_output('tt report'.split()).decode()
