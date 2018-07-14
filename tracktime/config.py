import os

import yaml


def get_config(filename=None):
    """Gets the configuration from ~/.config/tracktime/tracktimerc. If none
    exists, defaults are used.
    """
    defaults = {
        'customer_addresses': {},
        'customer_aliases': {},
        'directory': os.path.expanduser('~/.tracktime'),
        'gitlab_api_root': 'https://gitlab.com/api/v4/',
        'project_rates': {},
        'sync_time': False,
        'tableformat': 'simple',
    }

    if not filename:
        filename = os.path.expanduser('~/.config/tracktime/tracktimerc')

    if not os.path.exists(filename):
        return defaults

    with open(filename) as f:
        defaults.update(yaml.load(f) or {})
        return defaults
