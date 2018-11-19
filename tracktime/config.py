import os
from subprocess import check_output
from typing import Any, Dict

import yaml

cached_config: Dict[str, Any] = {}


def get_config(filename=None) -> Dict[str, Any]:
    """Gets the configuration from ~/.config/tracktime/tracktimerc. If none
    exists, defaults are used.
    """
    global cached_config
    if cached_config:
        return cached_config

    cached_config = {
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
        return cached_config

    with open(filename) as f:
        cached_config.update(yaml.load(f) or {})

    # If the API Key is a GPG file, decrypt it.
    api_key = cached_config.get('gitlab_api_key')
    if api_key and api_key.endswith('|'):
        cached_config['gitlab_api_key'] = check_output(
            api_key[:-1].split()).decode().strip()

    return cached_config
