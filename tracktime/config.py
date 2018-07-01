import os

import yaml


def get_config():
    defaults = {
        'directory': os.path.expanduser('~/.tracktime'),
        'project_rates': {},
        'customer_aliases': {},
        'customer_addresses': {},
        'tableformat': 'simple',
    }
    with open(os.path.expanduser('~/.config/tracktime/tracktimerc')) as f:
        defaults.update(yaml.load(f) or {})
        return defaults
