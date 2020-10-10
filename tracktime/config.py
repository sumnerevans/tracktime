import os
from subprocess import check_output
from typing import Any, Dict

import yaml

cached_config: Dict[str, Any] = {}


def get_config(filename=None) -> Dict[str, Any]:
    """
    Gets the configuration from the provided filename (or
    ~/.config/tracktime/tracktimerc if none is provided). If the file does not
    exist, defaults are used.
    """
    global cached_config
    if cached_config:
        return cached_config

    cached_config = {
        "fullname": "<Not Specified>",
        "customer_addresses": {},
        "customer_aliases": {},
        "customer_rates": {},
        "day_worked_min_threshold": 120,
        "directory": os.path.expanduser("~/.tracktime"),
        "gitlab": {
            "api_root": "https://gitlab.com/api/v4/",
        },
        "project_rates": {},
        "sync_time": False,
        "tableformat": "simple",
        "external_synchroniser_files": {},
    }

    if not filename:
        filename = (
            os.environ.get("XDG_CONFIG_HOME")
            or os.environ.get("APPDATA")
            or os.path.join(os.environ.get("HOME", os.path.expanduser("~")), ".config")
        )
        filename = os.path.join(filename, "tracktime/tracktimerc")

    if not os.path.exists(filename):
        return cached_config

    with open(filename) as f:
        cached_config.update(yaml.load(f, Loader=yaml.FullLoader) or {})

    # If the API Key is a shell command, execute it.
    # TODO (tracktime#15) move this to synchroniser code
    gitlab = cached_config.get("gitlab")
    if gitlab:
        api_key = gitlab.get("api_key")
        if api_key and api_key.endswith("|"):
            cached_config["gitlab"]["api_key"] = (
                check_output(api_key[:-1].split()).decode().strip()
            )
    sourcehut = cached_config.get("sourcehut")
    if sourcehut:
        access_token = sourcehut.get("access_token")
        if access_token and access_token.endswith("|"):
            cached_config["sourcehut"]["access_token"] = (
                check_output(access_token[:-1].split()).decode().strip()
            )

    if "gitlab_api_key" in cached_config:
        print(
            "\n".join(
                [
                    "DEPRECATION WARNING: GitLab configuration has been moved to a",
                    "    dictionary. See new example configuration here:",
                    "    https://gitlab.com/sumner/tracktime/snippets/1731133",
                ]
            )
        )
    return cached_config
