import csv
import json
import os
from datetime import datetime
from pathlib import Path
from subprocess import call

from requests import get, post
from tabulate import tabulate

from tracktime.config import get_config
from tracktime.time_entry import TimeEntry
from tracktime.time_parser import parse_time


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
        from tracktime.synchroniser import Synchroniser
        Synchroniser(self.date.year, self.date.month).sync()

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
