"""Time Entry class"""
import os
from datetime import datetime

from tracktime.entry_list import EntryList


class TimeEntry:
    def __init__(self, start, **kwargs):
        self.start = start
        self.stop = kwargs.get('stop', None)
        self.directory = kwargs.get('directory', None)
        self.type = kwargs.get('type', None)
        self.task = kwargs.get('task', None)
        self.description = kwargs.get('description', None)

    def __repr__(self):
        span = f'{self.start}-{self.stop}' if self.stop else self.start
        fields = ' '.join('{}={}'.format(f, getattr(self, f))
                          for f in ('type', 'task', 'description'))
        return f'<TimeEntry {span} {fields}>'

    @staticmethod
    def start(start, **kwargs):
        entries = EntryList.load(start.date())
        entries.append(TimeEntry(start, **kwargs))
        entries.save()

    @staticmethod
    def stop(stop, **kwargs):
        print(stop, kwargs)
        print(EntryList().list(datetime.now().date()))
