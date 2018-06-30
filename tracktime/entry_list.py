import csv
import os
from datetime import datetime
from pathlib import Path
from subprocess import call

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

        from tracktime.time_entry import TimeEntry
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
        total_seconds = sum(
            e.duration(allow_unended=True).seconds for e in self.entries)
        return (total_seconds // 3600, (total_seconds // 60) % 60)

    def append(self, entry):
        self.entries.append(entry)

    def save(self):
        with open(self.filepath, 'w') as f:
            fieldnames = [
                'start', 'stop', 'type', 'task', 'customer', 'description'
            ]
            writer = csv.DictWriter(f, fieldnames=fieldnames)

            writer.writeheader()
            for entry in self.entries:
                writer.writerow(dict(entry))

        # TODO sync with external providers

    def start(self, start, type, description, customer, task):
        if len(self.entries) > 0:
            self.entries[-1].stop = start

        time_entry = TimeEntry(start, type, description, customer, task)
        self.entries.append(time_entry)
        self.save()

    def stop(self, stop):
        entries = EntryList(stop.date())
        if len(entries) == 0:
            raise Exception('No time entry to end.')
        else:
            entries[-1].stop = stop
            entries.save()

    def edit(self):
        """Open an editor to edit the time entries."""
        editor = os.environ['EDITOR'] or os.environ['VISUAL']
        call([editor, _get_path(self.date, makedirs=True)])
        # TODO sync with external providers
