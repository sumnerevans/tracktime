"""Linear synchroniser module."""
import json
from pathlib import Path
from subprocess import check_output
from typing import Any, Dict, Optional

import requests

from tracktime.time_entry import TimeEntry
from tracktime.synchronisers.base import ExternalSynchroniser


def get_path(obj: Dict[str, Any], *path: str) -> Any:
    v = obj
    for k in path:
        v = v.get(k, {})
    return v


class LinearSynchroniser(ExternalSynchroniser):
    name = "Linear"
    types = ("linear",)

    def gql_query(self, query: str) -> Dict[str, Any]:
        query = query.replace("\n", " ")
        return requests.post(
            "https://api.linear.app/graphql",
            json={"query": f"{{{query}}}"},
            headers={"Authorization": self.api_key},
        ).json()

    def __init__(self, config):
        linear_config = config.get("linear", {})
        self.default_org = linear_config.get("default_org")
        self.api_key = linear_config.get("api_key")
        if self.api_key and self.api_key.endswith("|"):
            self.api_key = check_output(self.api_key[:-1].split()).decode().strip()

    def get_formatted_task_id(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.project or not entry.taskid:
            return None
        return f"{entry.project}-{entry.taskid}"

    def get_task_link(self, entry: TimeEntry) -> Optional[str]:
        if entry.type not in self.types or not entry.project or not entry.taskid:
            return None

        issue_id = f"{entry.project}-{entry.taskid}"
        return f"https://linear.app/{self.default_org}/issue/{issue_id}"

    def get_task_description(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None
        if not self.api_key or not self.default_org:
            return None

        task_id = self.get_formatted_task_id(entry)
        if not task_id:
            return None

        cache_path = Path("~/.cache/tracktime").expanduser()
        cache_path.mkdir(parents=True, exist_ok=True)
        cache_file = cache_path.joinpath("linear.json")

        description_cache: Dict[str, str] = {}
        if cache_file.exists():
            with open(cache_file, "r") as f:
                try:
                    description_cache = json.load(f)
                except Exception:
                    pass

        if not description_cache.get(task_id):
            query = f'issue (id: "{task_id}") {{ title }}'
            description = (
                (self.gql_query(query).get("data") or {}).get("issue", {}).get("title")
            )
            if description:
                description_cache[task_id] = description
                with open(cache_file, "w+") as f:
                    json.dump(description_cache, f)

        return description_cache.get(task_id)
