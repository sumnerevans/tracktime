import csv
import os

from pathlib import Path
from typing import Any, Dict, Union

from tracktime.time_entry import TimeEntry
from datetime import timedelta, date, datetime
from tracktime.time_parser import parse_time


def get_path(
    config: Dict[str, Any],
    date: Union[date, datetime],
    makedirs: bool = False,
) -> Path:
    """
    Returns the path for a given date.

    Arguments:
    date: the date[time] object representing the date to get a path for.
    makedirs: (optional) whether or not to create the corresponding directory

    Returns: ``Path`` for the date

    >>> from datetime import date
    >>> from tracktime.config import get_config
    >>> get_path(get_config(), date(2018, 1, 1)).parts[-3:] == (
    ... '2018', '01', '01')
    True
    """
    directory = Path(config["directory"])
    directory = directory.joinpath(str(date.year))
    directory = directory.joinpath("{:02}".format(date.month))

    if makedirs:
        os.makedirs(directory, exist_ok=True)

    return directory.joinpath("{:02}".format(date.day))


class EntryList:
    """
    A list of ``TimeEntry``s.

    Note: If customer is specified, the EntryList should only be used for
    display and saving is disabled due to the fact that it would override
    entries from other customers.
    """

    def __init__(self, config, date, customer=None):
        self.date = date
        self.config = config
        self.customer = customer
        self.entries = []

        # Load entries from the file
        self.filepath = get_path(self.config, date, makedirs=True)
        if os.path.exists(self.filepath):
            with open(self.filepath, "r") as f:
                for row in csv.DictReader(f):
                    # Convert times to datetimes
                    for k in ("start", "stop"):
                        if row[k]:
                            row[k] = parse_time(row[k])

                    if not self.customer or row["customer"] == self.customer:
                        self.entries.append(TimeEntry(**row))

    def __len__(self):
        return len(self.entries)

    def __getitem__(self, key):
        return self.entries[key]

    @property
    def total(self):
        total_minutes = sum(e.duration(allow_unended=True) for e in self.entries)
        return (total_minutes // 60, total_minutes % 60)

    def add_entry(self, entry):
        if self.customer:
            return

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
        if self.customer:
            return

        with open(self.filepath, "w", newline="") as f:
            fieldnames = [
                "start",
                "stop",
                "type",
                "project",
                "taskid",
                "customer",
                "description",
            ]
            writer = csv.DictWriter(f, fieldnames=fieldnames)

            writer.writeheader()
            for entry in self.entries:
                writer.writerow(dict(entry))

    def save_and_sync(self):
        if self.customer:
            return

        self.save()
        self.sync()

    def sync(self):
        if self.customer:
            return

        from tracktime.synchronisers import Synchroniser

        Synchroniser(self.config).sync(date(self.date.year, self.date.month, 1))

    def start(self, start, description, type, project, taskid, customer):
        if self.customer:
            return

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
        if self.customer:
            return

        if len(self) == 0 or self.entries[-1].stop:
            raise Exception("No time entry to end.")

        self.entries[-1].stop = stop
        self.save_and_sync()

    def resume(self, start, entry):
        if self.customer:
            return

        if len(self) == 0:
            yesterday = self.date - timedelta(days=1)
            old_entry = EntryList(self.config, yesterday).entries[-1]
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
