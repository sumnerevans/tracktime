import os
from subprocess import check_output

import yaml


def get_config(filename=None):
    """Gets the configuration from ~/.config/tracktime/tracktimerc. If none
    exists, defaults are used.
    """
    configuration_dict = {
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
        return configuration_dict

    with open(filename) as f:
        configuration_dict.update(yaml.load(f) or {})

    # If the API Key is a GPG file, decrypt it.
    api_key = configuration_dict.get('gitlab_api_key')
    if api_key and api_key.endswith('|'):
        configuration_dict['gitlab_api_key'] = check_output(
            api_key[:-1].split()).decode().strip()

    return configuration_dict
