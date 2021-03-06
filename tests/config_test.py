import os

from tracktime import config


def test_default_config():
    defaults = {
        "fullname": "<Not Specified>",
        "customer_addresses": {},
        "customer_aliases": {},
        "day_worked_min_threshold": 120,
        "directory": os.path.expanduser("~/.tracktime"),
        "gitlab": {
            "api_root": "https://gitlab.com/api/v4/",
        },
        "customer_rates": {},
        "project_rates": {},
        "report_statistics": True,
        "sync_time": False,
        "tableformat": "simple",
        "external_synchroniser_files": {},
    }

    # When the file doesn't exist, defaults should be used.
    assert config.get_config("does not exist") == defaults
