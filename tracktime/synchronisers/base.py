"""Synchroniser module"""
import csv
from collections import defaultdict
from datetime import date
from pathlib import Path
from subprocess import PIPE, run

from tracktime import EntryList
from tracktime.config import get_config


class ExternalSynchroniser:
    def sync(self, aggregated_time, synced_time):
        """
        Synchronise time over to the external service. All classes that inherit
        from ``ExternalSynchroniser`` must implement this function.

        Arguments:
        :param aggregated_time: a dictionary of (type, project, taskid) to
                                duration
        :param synced_time:     a dictionary of (type, project, taskid) to
                                duration

        Returns:
        a dictionary of (type, project, taskid) to duration
        """
        raise NotImplementedError('ExternalSynchroniser requires "sync" to be implemented.')


class Synchroniser:
    def __init__(self, year, month):
        """Initialize the Synchroniser.

        >>> s = Synchroniser(2018, 7)
        >>> assert (s.year, s.month) == (2018, 7)
        >>> str(s.month_dir)                               # doctest: +ELLIPSIS
        '.../2018/07'
        """
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
                # Skip any entries that don't have a type, project, or taskid.
                if not entry.type or not entry.project or not entry.taskid:
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

        from tracktime.synchronisers.gitlab import GitLabSynchroniser
        GitLabSynchroniser().sync(aggregated_time, synced_time)

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
