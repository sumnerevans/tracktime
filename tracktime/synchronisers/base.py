"""Synchroniser module"""
import csv
import importlib.util
import sys

from collections import defaultdict
from datetime import date
from pathlib import Path
from subprocess import PIPE, run
from typing import Optional, DefaultDict, Tuple

from tracktime import EntryList

AggregatedTime = DefaultDict[Tuple[str, str, str], int]


class ExternalSynchroniser:
    """
    Implementors of this class must handle parallelism and caching themselves.
    """

    def get_name(self):
        """
        Returns the human name for the external synchroniser.

        Returns:
        a string of the name of the external synchroniser.
        """
        raise NotImplementedError(
            'ExternalSynchroniser requires "get_name" to be implemented.'
        )

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
        raise NotImplementedError(
            'ExternalSynchroniser requires "sync" to be implemented.'
        )

    def get_formatted_task_id(self, entry) -> Optional[str]:
        """
        Gets the task ID formatted in a way that matches the way that the
        external service represents tasks.

        This is optional to implement. If ``None`` is returned, no formatting
        will be applied when displaying the task in reports.

        Arguments:
        :param entry: the ``TimeEntry`` to get a formatted task ID for.

        Returns:
        a string of the formatted task ID or ``None``
        """
        return None

    def get_task_link(self, entry) -> Optional[str]:
        """
        Gets a link to the task on the external service.

        This is optional to implement. If ``None`` is returned, the task
        descriptions will not link to the external service.

        Arguments:
        :param entry: the ``TimeEntry`` to get a task link for.

        Returns:
        a string of the URL of the task in the external service or ``None``
        """
        return None

    def get_task_description(self, entry) -> Optional[str]:
        """
        Get the description of a task from the external service.

        This is optional to implement. If ``None`` is returned, then the task ID
        will be shown without the description in reports.

        Arguments:
        :param entry: the ``TimeEntry`` to get a task description for.

        Returns:
        a string of the task description in the external service or ``None``
        """
        return None


class Synchroniser:
    def __init__(self, config):
        """Initialize the Synchroniser."""
        self.config = config
        self.synchronisers = None

    def _test_internet(self):
        """
        Tests whether or not the computer is currently connected to the
        internet.

        Uses ping of 8.8.8.8 to do this test. On Windows, it uses the ``-n``
        flag to specify to only do one ping. On POSIX OSes, it uses the ``-c``
        flag to specify the same.
        """
        is_win = sys.platform in ("win32", "cygwin")
        command = ["ping", "-n" if is_win else "-c", "1", "8.8.8.8"]
        return run(command, stdout=PIPE, stderr=PIPE).returncode == 0

    def get_synchronisers(self):
        parent = Path(__file__).parent
        synchronisers = {
            "gitlab": parent.joinpath("gitlab.py"),
            "sourcehut": parent.joinpath("sourcehut.py"),
        }
        synchronisers.update(self.config["external_synchroniser_files"])

        if self.synchronisers is None:
            self.synchronisers = []
            for module_name, file_path in synchronisers.items():
                spec = importlib.util.spec_from_file_location(module_name, file_path)
                module = importlib.util.module_from_spec(spec)
                spec.loader.exec_module(module)

                # Find the Synchroniser
                for element in filter(lambda x: not x.startswith("__"), dir(module)):
                    try:
                        item = getattr(module, element)
                        if not type(item) == type or item == ExternalSynchroniser:
                            continue
                        if isinstance(item(self.config), ExternalSynchroniser):
                            synchroniser = item(self.config)
                            break
                    except Exception:
                        pass
                else:
                    raise Exception(
                        f"Could not find valid synchroniser in {file_path}."
                    )

                self.synchronisers.append(synchroniser)

        return self.synchronisers

    def sync(self, first_of_month: date):
        """Synchronize time entries with external services."""
        year = first_of_month.year
        month = first_of_month.month
        month_dir = Path(
            self.config["directory"],
            str(year),
            "{:02}".format(month),
        )

        if not self.config["sync_time"]:
            print("Time sync disabled in configuration file.")
            return

        if not self._test_internet():
            print("No internet connection. Skipping sync.")
            return

        # Create a dictionary of the total time tracked for each GitLab taskid.
        aggregated_time: AggregatedTime = defaultdict(int)
        for day in range(1, 32):
            path = Path(month_dir, "{:02}".format(day))

            # Skip paths that don't exist
            if not path.exists():
                continue

            for entry in EntryList(
                self.config,
                date(year, month, day),
            ).entries:
                # Skip any entries that don't have a type, project, or taskid.
                if not entry.type or not entry.project or not entry.taskid:
                    continue
                # Skip any un-ended entries.
                if not entry.stop:
                    continue

                task_tuple = (entry.type, entry.project, entry.taskid)
                aggregated_time[task_tuple] += entry.duration()

        # Create a dictionary of all of the synchronised taskids.
        synced_time: AggregatedTime = defaultdict(int)
        synced_file_path = Path(month_dir, ".synced")
        if synced_file_path.exists():
            with open(synced_file_path, "r") as f:
                for row in csv.DictReader(f):
                    task_tuple = (row["type"], row["project"], row["taskid"])
                    synced_time[task_tuple] = int(row["synced"])

        for synchroniser in self.get_synchronisers():
            print(f"Syncronizing with {synchroniser.get_name()}.")
            synced_time.update(synchroniser.sync(aggregated_time, synced_time))

        # Update the .synced file with the updated amounts.
        with open(synced_file_path, "w+", newline="") as f:
            fieldnames = ["type", "project", "taskid", "synced"]
            writer = csv.DictWriter(f, fieldnames=fieldnames)

            writer.writeheader()
            for task_tuple, synced in synced_time.items():
                writer.writerow(
                    {
                        "type": task_tuple[0],
                        "project": task_tuple[1],
                        "taskid": task_tuple[2],
                        "synced": synced,
                    }
                )

    def get_formatted_task_id(self, entry) -> Optional[str]:
        for synchroniser in self.get_synchronisers():
            formatted = synchroniser.get_formatted_task_id(entry)
            if formatted:
                return formatted
        return entry.taskid

    def get_task_link(self, entry) -> Optional[str]:
        for synchroniser in self.get_synchronisers():
            task_link = synchroniser.get_task_link(entry)
            if task_link:
                return task_link
        return None

    def get_task_description(self, entry) -> Optional[str]:
        for synchroniser in self.get_synchronisers():
            task_description = synchroniser.get_task_description(entry)
            if task_description:
                return task_description
        return None
