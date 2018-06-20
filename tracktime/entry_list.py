import csv
import os
from datetime import datetime
from pathlib import Path
from subprocess import call

from tabulate import tabulate

import tracktime


def _get_path(date, makedirs=False):
    directory = Path(tracktime.root_directory)
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
                    self.entries.append(TimeEntry(**row))

    def __len__(self):
        return len(self.entries)

    def __getitem__(self, key):
        return self.entries[key]

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

    @property
    def total(self):
        total_seconds = sum(
            e.duration(allow_unended=True).seconds for e in self.entries)
        hours, minutes = total_seconds // 3600, (total_seconds // 60) % 60
        return f'{hours}:{minutes:02}'

    @staticmethod
    def list(date, **kwargs):
        """Gives you a list of ``TimeEntry``s for the given date."""
        if isinstance(date, str):
            date = datetime.strptime(date, '%Y-%m-%d')

        entry_list = EntryList(date)
        print(f'Entries for {date}')
        print('=' * 22)
        print()
        print(tabulate([dict(x) for x in entry_list], headers='keys'))

        print()
        print(f'Total: {entry_list.total}')

    @staticmethod
    def edit(date, **kwargs):
        """Open an editor to edit the time entries."""
        editor = os.environ['EDITOR'] or os.environ['VISUAL']
        call([editor, _get_path(date)])
        # TODO sync with external providers
