"""
Sourcehut synchroniser module
"""
import concurrent.futures
import pickle
import re
import threading
from collections import defaultdict
from pathlib import Path
from typing import Dict, Optional, Tuple
from urllib import parse

from requests import get, post, put

from tracktime.synchronisers.base import AggregatedTime, ExternalSynchroniser


class SourcehutSynchroniser(ExternalSynchroniser):
    name = "Sourcehut"
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
            username, tracker = project_str.split("/")
        else:
            username, tracker = self.username, project_str

        if not username.startswith("~"):
            username = "~" + username
        return username, tracker

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

    @staticmethod
    def format_duration(duration_minutes: int):
        hours = duration_minutes // 60
        minutes = duration_minutes % 60
        return (
            f"{hours} {SourcehutSynchroniser.pluralize('hour', hours)} "
            if hours > 0
            else ""
        ) + f"{minutes} {SourcehutSynchroniser.pluralize('minute', minutes)}"

    firstline_re = re.compile(
        r"\[tracktime\] ~\w+ has spent (\d+ hours )?\d+ minutes on this task.?"
    )
    month_line_re = re.compile(r"\s*\* (\d+)-(\d+): (?:(\d+) hours )?(\d+) minutes")

    @staticmethod
    def parse_comment(text: str) -> Optional[Dict[Tuple[int, int], int]]:
        """
        Parses comments of the form::

            [tracktime] ~sumner has spent 12 hours 48 minutes on this task.
            * 2020-10: 8 hours 12 minutes
            * 2020-11: 4 hours 36 minutes

        and returns a dictionary of this form::

            {
                (2020, 10): 492,
                (2020, 11): 276,
            }

        >>> SourcehutSynchroniser.parse_comment(
        ... '''[tracktime] ~sumner has spent 48 minutes on this task.
        ...  * 2020-11: 48 minutes''')
        {(2020, 11): 48}
        >>> SourcehutSynchroniser.parse_comment(
        ... '''[tracktime] ~sumner has spent 48 minutes on this task.
        ...  * 2020-1: 49 minutes''')
        {(2020, 1): 49}
        >>> SourcehutSynchroniser.parse_comment(
        ... '''[tracktime] ~sumner has spent 48 minutes on this task.
        ...  * 2020-01: 49 minutes''')
        {(2020, 1): 49}
        >>> SourcehutSynchroniser.parse_comment(
        ... '''[tracktime] ~sumner has spent 12 hours 48 minutes on this task.
        ...  * 2020-10: 8 hours 12 minutes
        ...  * 2020-11: 4 hours 36 minutes''')
        {(2020, 10): 492, (2020, 11): 276}
        """
        lines = text.split("\n")
        if not SourcehutSynchroniser.firstline_re.match(lines[0]):
            return None
        month_data = {}
        for month_line in lines[1:]:
            match = SourcehutSynchroniser.month_line_re.match(month_line)
            if match:
                vals = (int(x) if x is not None else 0 for x in match.groups())
                year, month, hours, minutes = vals
                month_data[(year, month)] = hours * 60 + minutes

        return month_data

    @staticmethod
    def generate_comment(month_data: Dict[Tuple[int, int], int], username: str) -> str:
        """
        Takes data of the form::

            {
                (2020, 10): 492,
                (2020, 11): 276,
            }
            [tracktime] ~sumner has spent 12 hours 48 minutes on this task.
            * 2020-10: 8 hours 12 minutes
            * 2020-11: 4 hours 36 minutes

        and returns a string of this form::

            [tracktime] ~sumner has spent 12 hours 48 minutes on this task.
            * 2020-10: 8 hours 12 minutes
            * 2020-11: 4 hours 36 minutes
        """
        total_duration_str = SourcehutSynchroniser.format_duration(
            sum(month_data.values())
        )
        lines = [f"[tracktime] {username} has spent {total_duration_str} on this task."]
        for ((year, month), duration) in sorted(month_data.items()):
            duration_str = SourcehutSynchroniser.format_duration(duration)
            lines.append(f" * {year}-{month:02}: {duration_str}")
        return "\n".join(lines)

    def sync(
        self,
        aggregated_time: AggregatedTime,
        synced_time: AggregatedTime,
        year_month: Tuple[int, int],
    ) -> AggregatedTime:
        """Synchronize time entries with Sourcehut."""
        if not self.username:
            return synced_time

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

            time_diff = duration - synced_time[task_tuple]

            # Skip tasks that don't have any change.
            if time_diff == 0:
                return

            ticket_uri = f"/user/{username}/trackers/{tracker}/tickets/{taskid}"

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
                comment_month_data = None

                for result in results:
                    if "comment" in result.get("event_type", []):
                        comment = result.get("comment", {})
                        comment_text = comment.get("text", "")

                        comment_month_data = (
                            SourcehutSynchroniser.parse_comment(comment_text) or {}
                        )

                        # We don't have to do anything if the duration is already
                        # correct.
                        if duration == comment_month_data.get(year_month):
                            with synced_time_lock:
                                synced_time[task_tuple] = duration
                            return

                        # This is the tracktime comment.
                        if comment_text.startswith(tracktime_prefix):
                            comment_id = comment.get("id")
                            break

                month_data = {**(comment_month_data or {}), year_month: duration}
                new_text = SourcehutSynchroniser.generate_comment(
                    month_data, self.username
                )

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

                duration_str = self.format_duration(duration)
                print("[SUCCESS]" if result.status_code == 200 else "[FAILED]", end=" ")
                print(
                    f"setting time spent on {project}#{taskid} in "
                    f"{year_month[0]}-{year_month[1]:02} to {duration_str}."
                )

                if result.status_code == 200:
                    with synced_time_lock:
                        synced_time[task_tuple] = duration

            except Exception:
                return None

        normalized_aggregated_time: AggregatedTime = defaultdict(int)
        for (type_, project, taskid), v in aggregated_time.items():
            username, tracker = self._extract_username_and_tracker(project)
            ticketid = taskid.strip("#")
            project = f"{username}/{tracker}"
            normalized_aggregated_time[(type_, project, ticketid)] += v

        concurrent.futures.wait(
            [
                self.executor.submit(do_sync(k, v))
                for k, v in normalized_aggregated_time.items()
            ],
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
                description = self._make_request(uri, requester=get).json().get("title")
            except Exception:
                return None

            description_cache[(entry.project, entry.taskid)] = description
            with open(cache_file, "wb+") as f:
                pickle.dump(description_cache, f)

        return description_cache[(entry.project, entry.taskid)]
