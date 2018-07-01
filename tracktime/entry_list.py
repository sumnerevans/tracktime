import csv
import json
import os
from collections import defaultdict
from datetime import datetime
from pathlib import Path
from subprocess import PIPE, call, run
from urllib import parse

from requests import get, post
from tabulate import tabulate

from tracktime.config import get_config
from tracktime.time_entry import TimeEntry
from tracktime.time_parser import parse_time


def _test_internet():
    """
    Tests whether or not the computer is currently connected to the internet.
    """
    command = ['ping', '-c', '1', '8.8.8.8']
    return run(command, stdout=PIPE, stderr=PIPE).returncode == 0


def _get_path(date, makedirs=False):
    directory = Path(get_config()['directory'])
    directory = directory.joinpath(str(date.year))
    directory = directory.joinpath('{:02}'.format(date.month))

    if makedirs:
        os.makedirs(directory, exist_ok=True)

    return directory.joinpath('{:02}'.format(date.day))


class EntryList:
    def __init__(self, date):
        self.date = date
        self.entries = []

        # Load entries from the file
        self.filepath = _get_path(date, makedirs=True)
        if os.path.exists(self.filepath):
            with open(self.filepath, 'r') as f:
                for row in csv.DictReader(f):
                    # Convert times to datetimes
                    for k in ('start', 'stop'):
                        if row[k]:
                            row[k] = parse_time(row[k])

                    self.entries.append(TimeEntry(**row))

    def __len__(self):
        return len(self.entries)

    def __getitem__(self, key):
        return self.entries[key]

    @property
    def total(self):
        total_minutes = sum(
            e.duration(allow_unended=True) for e in self.entries)
        return (total_minutes // 60, total_minutes % 60)

    def append(self, entry):
        self.entries.append(entry)

    def save(self):
        with open(self.filepath, 'w') as f:
            fieldnames = [
                'start', 'stop', 'type', 'project', 'taskid', 'customer',
                'description'
            ]
            writer = csv.DictWriter(f, fieldnames=fieldnames)

            writer.writeheader()
            for entry in self.entries:
                writer.writerow(dict(entry))

    def save_and_sync(self):
        self.save()
        self.sync()

    def sync(self):
        """Synchronize time entries with external services."""
        print('Syncronizing time entries...')
        if not _test_internet():
            print('No internet connection. Skipping sync.')
            return

        config = get_config()
        username = config.get('gitlab_username')

        def make_request(rel_path, requester=get, params={}):
            params = parse.urlencode({
                'private_token': config.get('gitlab_api_key'),
                **params
            })
            rel_path = rel_path[1:] if rel_path.startswith('/') else rel_path
            path = parse.urljoin(get_config()['gitlab_api_root'], rel_path)
            return requester(path, params)

        aggregated_time = defaultdict(int)
        for entry in self.entries:
            # Skip any entries that are not GitLab entries.
            if entry.type not in ('gl', 'gitlab'):
                continue
            # Skip any entries that don't have a project or task.
            if not entry.project or not entry.taskid:
                continue
            # Skip any un-ended entries.
            if not entry.stop:
                continue

            project = parse.quote(entry.project).replace('/', '%2F')
            uri = f'{project}/issues/{entry.taskid}'
            aggregated_time[uri] = (
                entry.duration() + aggregated_time.get(uri, 0))

        for proj, duration in aggregated_time.items():
            result = make_request(f'/projects/{project}/issues/{entry.taskid}/time_stats')
            print(result.text)

        # TODO Figure out how to determine what needs to be synced

    def start(self, start, description, type, project, taskid, customer):
        if len(self.entries) > 0:
            self.entries[-1].stop = start

        time_entry = TimeEntry(
            start,
            description,
            type=type,
            project=project,
            taskid=taskid,
            customer=customer,
        )
        self.entries.append(time_entry)
        self.save_and_sync()

    def stop(self, stop):
        entries = EntryList(stop.date())
        if len(entries) == 0:
            raise Exception('No time entry to end.')
        else:
            entries[-1].stop = stop
            entries.save_and_sync()

    def edit(self):
        """Open an editor to edit the time entries."""
        # Ensure the header exists.
        EntryList(self.date).save()

        # Edit the entries
        editor = os.environ['EDITOR'] or os.environ['VISUAL']
        call([editor, _get_path(self.date, makedirs=True)])

        # Reload and sync the time entries
        EntryList(self.date).sync()
