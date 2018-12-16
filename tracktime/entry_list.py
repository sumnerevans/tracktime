import csv
import os
from pathlib import Path

from tracktime.config import get_config
from tracktime.time_entry import TimeEntry
from datetime import timedelta
from tracktime.time_parser import parse_time


def get_path(date, makedirs=False):
    """
    Returns the path for a given date.

    Arguments:
    date: the date[time] object representing the date to get a path for.
    makedirs: (optional) whether or not to create the corresponding directory

    >>> from datetime import date
    >>> dir = get_config()['directory']
    >>> str(get_path(date(2018, 1, 1))) == f'{dir}/2018/01/01'
    True
    """
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
        self.filepath = get_path(date, makedirs=True)
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

    def add_entry(self, entry):
        # Determine where the entry does in the list.
        index = len(self.entries) + 1  # default to the end
        for i, e in enumerate(self.entries):
            if e.stop and e.start <= entry.start < e.stop:
                # The entry is being started in the middle of this one.
                entry.stop = e.stop
                e.stop = entry.start
                index = i + 1
                break

            if entry.start < e.start:
                # The entry is being started before this.
                entry.stop = e.start
                index = i
                break

            # There is an unended time entry. Stop it, and start the new one.
            if e.start <= entry.start and not e.stop:
                e.stop = entry.start
                index = i + 1
                break

        self.entries.insert(index, entry)

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
        from tracktime.synchronisers import Synchroniser
        Synchroniser(self.date.year, self.date.month).sync()

    def start(self, start, description, type, project, taskid, customer):
        time_entry = TimeEntry(
            start,
            description,
            type=type,
            project=project,
            taskid=taskid,
            customer=customer,
        )
        self.add_entry(time_entry)
        self.save_and_sync()

    def stop(self, stop):
        if len(self) == 0:
            raise Exception('No time entry to end.')

        self.entries[-1].stop = stop
        self.save_and_sync()

    def resume(self, start, entry):
        if len(self) == 0:
            yesterday = self.date - timedelta(days=1)
            old_entry = EntryList(yesterday).entries[-1]
        else:
            old_entry = self.entries[entry]

        self.start(
            start,
            old_entry.description,
            old_entry.type,
            old_entry.project,
            old_entry.taskid,
            old_entry.customer,
        )
