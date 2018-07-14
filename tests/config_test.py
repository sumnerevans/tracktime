import os

from tracktime import config


def test_default_config():
    defaults = {
        'customer_addresses': {},
        'customer_aliases': {},
        'directory': os.path.expanduser('~/.tracktime'),
        'gitlab_api_root': 'https://gitlab.com/api/v4/',
        'project_rates': {},
        'sync_time': False,
        'tableformat': 'simple',
    }

    # When the file doesn't exist, defaults should be used.
    assert config.get_config('does not exist') == defaults