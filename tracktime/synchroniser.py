"""Synchroniser module"""
import csv
from collections import defaultdict
from datetime import date
from pathlib import Path
from subprocess import PIPE, call, run
from urllib import parse

from requests import get, post

from tracktime import EntryList
from tracktime.config import get_config


class Synchroniser:
    def __init__(self, year, month):
        self.year = year
        self.month = month

        self.config = get_config()

        self.month_dir = Path(
            self.config['directory'],
            str(self.year),
            '{:02}'.format(self.month),
        )

    def _test_internet(self):
        """
        Tests whether or not the computer is currently connected to the
        internet.
        """
        command = ['ping', '-c', '1', '8.8.8.8']
        return run(command, stdout=PIPE, stderr=PIPE).returncode == 0

    def _make_request(self, rel_path, requester=post, params={}):
        params = parse.urlencode({
            'private_token': self.config.get('gitlab_api_key'),
            **params
        })
        rel_path = rel_path[1:] if rel_path.startswith('/') else rel_path
        path = parse.urljoin(self.config['gitlab_api_root'], rel_path)
        return requester(path, params)

    def sync(self):
        """Synchronize time entries with external services."""
        if not self.config['sync_time']:
            print('Time sync disabled in configuration file.')
            return

        if not self._test_internet():
            print('No internet connection. Skipping sync.')
            return

        # Create a dictionary of the total time tracked for each GitLab taskid.
        aggregated_time = defaultdict(int)
        for day in range(1, 32):
            path = Path(self.month_dir, '{:02}'.format(day))

            # Skip paths that don't exist
            if not path.exists():
                continue

            for entry in EntryList(date(self.year, self.month, day)).entries:
                # Skip any entries that are not GitLab entries.
                if entry.type not in ('gl', 'gitlab'):
                    continue
                # Skip any entries that don't have a project or task.
                if not entry.project or not entry.taskid:
                    continue
                # Skip any un-ended entries.
                if not entry.stop:
                    continue

                task_tuple = (entry.type, entry.project, entry.taskid)
                aggregated_time[task_tuple] += entry.duration()

        # Create a dictionary of all of the synchronised taskids.
        synced_time = defaultdict(int)
        synced_file_path = Path(self.month_dir, '.synced')
        if synced_file_path.exists():
            with open(synced_file_path, 'r') as f:
                for row in csv.DictReader(f):
                    task_tuple = (row['type'], row['project'], row['taskid'])
                    synced_time[task_tuple] = int(row['synced'])

        # Go through all of the aggredated time and determine how much time
        # needs to be syncrhonised over to GitLab for each taskid.
        for task_tuple, duration in aggregated_time.items():
            time_diff = duration - synced_time[task_tuple]

            # Skip tasks that don't have any change.
            if time_diff == 0:
                continue

            type, project, taskid = task_tuple
            print(f'Adding {time_diff}m to {project}#{taskid}')

            project = parse.quote(project).replace('/', '%2F')
            uri = f'/projects/{project}/issues/{taskid}/add_spent_time'
            params = {'duration': f'{time_diff}m'}
            result = self._make_request(uri, params=params)

            # If successful, update the amount that has been synced.
            if result.status_code == 201:
                synced_time[task_tuple] += time_diff

        # Update the .synced file with the updated amounts.
        with open(synced_file_path, 'w+') as f:
            fieldnames = ['type', 'project', 'taskid', 'synced']
            writer = csv.DictWriter(f, fieldnames=fieldnames)

            writer.writeheader()
            for task_tuple, synced in synced_time.items():
                writer.writerow({
                    'type': task_tuple[0],
                    'project': task_tuple[1],
                    'taskid': task_tuple[2],
                    'synced': synced,
                })
