"""Synchroniser module"""
from urllib import parse

from requests import post
from tracktime.config import get_config
from tracktime.synchronisers.base import ExternalSynchroniser


class GitLabSynchroniser(ExternalSynchroniser):
    def __init__(self):
        config = get_config()
        self.api_key = config.get('gitlab').get('api_key')
        self.api_root = config.get('gitlab').get('api_root')

    def _make_request(self, rel_path, requester=post, params={}):
        params = parse.urlencode({'private_token': self.api_key, **params})
        rel_path = rel_path[1:] if rel_path.startswith('/') else rel_path
        path = parse.urljoin(self.api_root, rel_path)
        return requester(path, params)

    def sync(self, aggregated_time, synced_time):
        """Synchronize time entries with GitLab."""
        # Go through all of the aggredated time and determine how much time
        # needs to be synchronised over to GitLab for each taskid.
        for task_tuple, duration in aggregated_time.items():
            time_diff = duration - synced_time[task_tuple]

            # Skip tasks that don't have any change.
            if time_diff == 0:
                continue

            type, project, taskid = task_tuple
            print(f'Adding {time_diff}m to {project}{taskid}...', end='')

            project = parse.quote(project).replace('/', '%2F')
            task_type = {'#': 'issue', '!': 'merge_request'}[taskid[0]]
            task_number = taskid[1:]
            uri = f'/projects/{project}/{task_type}s/{task_number}/add_spent_time'
            params = {'duration': f'{time_diff}m'}
            result = self._make_request(uri, params=params)

            # If successful, update the amount that has been synced.
            if result.status_code == 201:
                print(' SUCCESS')
                synced_time[task_tuple] += time_diff
            else:
                print(' FAILED')
                print(result.text)

        return synced_time
