"""
Sourcehut synchroniser module
"""
import concurrent.futures
import pickle
import re
import threading
from pathlib import Path
from typing import Dict, Optional, Tuple
from urllib import parse

from requests import get, post, put

from tracktime.synchronisers.base import ExternalSynchroniser


class SourcehutSynchroniser(ExternalSynchroniser):
    types = ("srht", "sr.ht", "sh", "sourcehut")

    def __init__(self, config):
        self.access_token = config.get("sourcehut", {}).get("access_token")
        self.api_root = config.get("sourcehut", {}).get("api_root")
        self.username = config.get("sourcehut", {}).get("username")
        self.executor = concurrent.futures.ThreadPoolExecutor(max_workers=50)

    def _make_request(self, rel_path, requester=post, params={}):
        rel_path = rel_path[1:] if rel_path.startswith("/") else rel_path
        path = parse.urljoin(self.api_root, rel_path)
        headers = {"Authorization": f"token {self.access_token}"}
        return requester(path, params, headers=headers)

    def _extract_username_and_tracker(self, project_str):
        if "/" in project_str:
            return project_str.split("/")
        else:
            return self.username, project_str

    @staticmethod
    def pluralize(string: str, number: int, pluralized_form: str = None) -> str:
        """
        Pluralize the given string given the count as a number.

        >>> SourcehutSynchroniser.pluralize('foo', 1)
        'foo'
        >>> SourcehutSynchroniser.pluralize('foo', 2)
        'foos'
        >>> SourcehutSynchroniser.pluralize('foo', 0)
        'foos'
        """
        if number != 1:
            return pluralized_form or f"{string}s"
        return string

    def format_duration(self, duration_minutes):
        hours = duration_minutes // 60
        minutes = duration_minutes % 60
        return (
            f"{hours} {SourcehutSynchroniser.pluralize('hour', hours)} "
            if hours > 0
            else ""
        ) + f"{minutes} {SourcehutSynchroniser.pluralize('minute', minutes)}"

    def get_name(self):
        return "Sourcehut"

    def sync(self, aggregated_time, synced_time):
        """Synchronize time entries with Sourcehut."""
        # Go through all of the aggredated time and determine how much time
        # needs to be synchronised over to GitLab for each taskid.
        synced_time_lock = threading.Lock()

        def do_sync(task_tuple, duration):
            type_, project, taskid = task_tuple
            # Skip items which are not GitLab
            if type_.lower() not in self.types:
                return

            username, tracker = self._extract_username_and_tracker(project)

            # Only do time tracking on repos that are mine.
            # TODO (#16) need to update this when the organization stuff is working.
            if username != self.username:
                return

            ticketid = taskid.strip("#")
            project = f"{username}/{tracker}"

            time_diff = duration - synced_time[(type_, project, taskid)]

            # Skip tasks that don't have any change.
            if time_diff == 0:
                return

            ticket_uri = f"/user/{username}/trackers/{tracker}/tickets/{ticketid}"

            try:
                # Try and find the existing comment and edit it.
                events_uri = f"{ticket_uri}/events"
                results = (
                    self._make_request(events_uri, requester=get)
                    .json()
                    .get("results", [])
                )
                comment_id = None
                tracktime_prefix = f"[tracktime] {self.username}"
                duration_str = self.format_duration(duration)
                new_text = f"{tracktime_prefix} has spent {duration_str} on this task"

                for result in results:
                    if "comment" in result.get("event_type", []):
                        comment = result.get("comment", {})
                        comment_text = comment.get("text", "")

                        # We don't have to do anything if the comment_text is already
                        # correct.
                        if comment_text == new_text:
                            with synced_time_lock:
                                synced_time[(type_, project, taskid)] = duration
                            return

                        if comment_text.startswith(tracktime_prefix):
                            comment_id = comment.get("id")
                            break

                if comment_id:
                    edit_url = f"{ticket_uri}/comments/{comment_id}"
                    result = self._make_request(
                        edit_url,
                        params={"text": new_text},
                        requester=put,
                    )
                else:
                    create_url = ticket_uri
                    result = self._make_request(
                        create_url,
                        params={"comment": new_text},
                        requester=put,
                    )

                print("[SUCCESS]" if result.status_code == 200 else "[FAILED]", end=" ")
                print(f"setting time spent on {project}#{ticketid} to {duration_str}.")

                if result.status_code == 200:
                    with synced_time_lock:
                        synced_time[(type_, project, taskid)] = duration

            except Exception:
                return None

        concurrent.futures.wait(
            [self.executor.submit(do_sync(k, v)) for k, v in aggregated_time.items()],
            timeout=None,
            return_when=concurrent.futures.ALL_COMPLETED,
        )

        return synced_time

    def get_formatted_task_id(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None
        return entry.taskid if entry.taskid.startswith("#") else f"#{entry.taskid}"

    def get_task_link(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None
        if not self.api_root:
            return None
        root_match = re.match("(.*)/api/?", self.api_root)
        if not root_match:
            return None
        root = root_match.group(1)

        username, tracker = self._extract_username_and_tracker(entry.project)
        return f"{root}/{username}/{tracker}/{entry.taskid.strip('#')}"

    def get_task_description(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None
        if not self.access_token or not self.api_root:
            return None

        cache_path = Path("~/.cache/tracktime").expanduser()
        cache_path.mkdir(parents=True, exist_ok=True)
        cache_file = cache_path.joinpath("sourcehut.pickle")

        description_cache: Dict[Tuple[str, str], str] = {}
        if cache_file.exists():
            with open(cache_file, "rb") as f:
                try:
                    description_cache = pickle.load(f)
                except Exception:
                    pass

        username, tracker = self._extract_username_and_tracker(entry.project)

        if not description_cache.get((entry.project, entry.taskid)):
            ticketid = entry.taskid.strip("#")
            uri = f"/user/{username}/trackers/{tracker}/tickets/{ticketid}"
            try:
                description = (
                    self._make_request(
                        uri,
                        requester=get,
                    )
                    .json()
                    .get("title")
                )
            except Exception:
                return None

            description_cache[(entry.project, entry.taskid)] = description
            with open(cache_file, "wb+") as f:
                pickle.dump(description_cache, f)

        return description_cache[(entry.project, entry.taskid)]
