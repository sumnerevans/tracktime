"""Synchroniser module"""
import re
import concurrent.futures
import threading

from urllib import parse
from typing import Optional

from requests import post, get
from tracktime.config import get_config
from tracktime.synchronisers.base import ExternalSynchroniser


class GitLabSynchroniser(ExternalSynchroniser):
    types = ('gl', 'gitlab')
    task_types = {'#': 'issue', '!': 'merge_request'}

    def __init__(self):
        config = get_config()
        self.api_key = config.get('gitlab').get('api_key')
        self.api_root = config.get('gitlab').get('api_root')
        self.executor = concurrent.futures.ThreadPoolExecutor(max_workers=50)

    def _make_request(self, rel_path, requester=post, params={}):
        params = parse.urlencode({'private_token': self.api_key, **params})
        rel_path = rel_path[1:] if rel_path.startswith('/') else rel_path
        path = parse.urljoin(self.api_root, rel_path)
        return requester(path, params)

    def get_name(self):
        return 'GitLab'

    def sync(self, aggregated_time, synced_time):
        """Synchronize time entries with GitLab."""
        # Go through all of the aggredated time and determine how much time
        # needs to be synchronised over to GitLab for each taskid.
        synced_time_lock = threading.Lock()

        def do_sync(task_tuple, duration):
            type_, project, taskid = task_tuple
            # Skip items which are not GitLab
            if type_.lower() not in ('gl', 'gitlab'):
                return

            time_diff = duration - synced_time[task_tuple]

            # Skip tasks that don't have any change.
            if time_diff == 0:
                return

            escaped_project = parse.quote(project).replace('/', '%2F')
            task_type = self.task_types[taskid[0]]
            task_number = taskid[1:]
            uri = f'/projects/{escaped_project}/{task_type}s/{task_number}/add_spent_time'
            params = {'duration': f'{time_diff}m'}
            result = self._make_request(uri, params=params)

            # If successful, update the amount that has been synced.
            if result.status_code == 201:
                print(f'[SUCCESS] Adding {time_diff}m to {project}{taskid}.')
                with synced_time_lock:
                    synced_time[task_tuple] += time_diff
            else:
                print(
                    f'[FAILED] Adding {time_diff}m to {project}{taskid}.\n' +
                    result.text
                )

        concurrent.futures.wait(
            [self.executor.submit(do_sync(k, v))
             for k, v in aggregated_time.items()],
            timeout=None,
            return_when=concurrent.futures.ALL_COMPLETED,
        )

        return synced_time

    def get_task_link(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return
        root = get_config().get('gitlab', {}).get('api_root')
        root_match = re.match('(.*)/api/v4/?', root)
        if not root_match:
            return
        root = root_match.group(1)

        task_type = self.task_types[entry.taskid[0]]
        return f'{root}/{entry.project}/{task_type}s/{entry.taskid[1:]}'

    def get_task_description(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return
        escaped_project = parse.quote(entry.project).replace('/', '%2F')
        task_type = self.task_types[entry.taskid[0]]
        task_number = entry.taskid[1:]
        uri = f'/projects/{escaped_project}/{task_type}s/{task_number}'
        try:
            return self._make_request(uri, requester=get).json().get('title')
        except Exception:
            return
