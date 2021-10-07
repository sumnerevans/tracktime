"""GitHub synchroniser module."""
import pickle
from pathlib import Path
from typing import Any, Dict, Optional, Tuple

import requests

from tracktime.time_entry import TimeEntry
from tracktime.synchronisers.base import ExternalSynchroniser


def get_path(obj: Dict[str, Any], *path: str) -> Any:
    v = obj
    for k in path:
        v = v.get(k, {})
    return v


class GitHubSynchroniser(ExternalSynchroniser):
    name = "GitHub"
    types = ("gh", "github")

    def gql_query(self, query: str) -> Dict[str, Any]:
        query = query.replace("\n", " ")
        return requests.post(
            "https://api.github.com/graphql",
            json={"query": f"query{{{query}}}"},
            headers={"Authorization": f"bearer {self.access_token}"},
        ).json()

    def __init__(self, config):
        github_config = config.get("github", {})
        self.access_token = github_config.get("access_token")
        self.username = github_config.get("username")

    def get_formatted_task_id(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None
        return entry.taskid if entry.taskid.startswith("#") else f"#{entry.taskid}"

    def get_task_link(self, entry: TimeEntry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None

        # Prefix the project with the username if no namespace is specified.
        project = entry.project
        if "/" not in entry.project:
            if self.username is None:
                return None
            project = f"{self.username}/{entry.project}"

        task_id = entry.taskid
        if task_id.startswith("#"):
            task_id = task_id[1:]

        # Always link to /issues/ because it will redirect to /pull/ if necessary.
        return f"https://github.com/{project}/issues/{task_id}"

    def get_task_description(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None
        if not self.access_token:
            return None

        cache_path = Path("~/.cache/tracktime").expanduser()
        cache_path.mkdir(parents=True, exist_ok=True)
        cache_file = cache_path.joinpath("github.pickle")

        description_cache: Dict[Tuple[str, str], str] = {}
        if cache_file.exists():
            with open(cache_file, "rb") as f:
                try:
                    description_cache = pickle.load(f)
                except Exception:
                    pass

        # Prefix the project with the username if no namespace is specified.
        project: str = entry.project
        if "/" not in entry.project:
            if self.username is None:
                return None
            project = f"{self.username}/{entry.project}"

        owner, name = project.split("/")

        task_id: str = entry.taskid
        if task_id.startswith("#"):
            task_id = task_id[1:]

        if not description_cache.get((project, task_id)):
            task_types = ("issue", "pullRequest", "discussion")
            task_type_query_parts = [
                f"{t}(number: {task_id}) {{ title }}" for t in task_types
            ]
            query = f"""
            repository(owner: "{owner}", name: "{name}") {{
                {"".join(task_type_query_parts)}
            }}
            """
            repository = get_path(self.gql_query(query), "data", "repository")
            description = None
            for t in task_types:
                if value := repository.get(t):
                    description = value.get("title")
            if description is None:
                return None

            description_cache[(project, task_id)] = description
            with open(cache_file, "wb+") as f:
                pickle.dump(description_cache, f)

        return description_cache[(project, task_id)]
