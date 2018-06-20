"""Time Entry class"""
import os
import time
from datetime import datetime

from tracktime.entry_list import EntryList


class TimeEntry:
    def __init__(self, start, **kwargs):
        self.start = start
        self.stop = kwargs.get('stop', None)

        if isinstance(self.start, str):
            timestamp = time.mktime(time.strptime(self.start, '%H:%M'))
            self.start = datetime.fromtimestamp(timestamp)

        if isinstance(self.stop, str) and len(self.stop) > 0:
            timestamp = time.mktime(time.strptime(self.stop, '%H:%M'))
            self.stop = datetime.fromtimestamp(timestamp)

        self.type = kwargs.get('type', None)
        self.task = kwargs.get('task', None)
        self.description = kwargs.get('description', None)
        self.customer = kwargs.get('customer', None)

    def __repr__(self):
        start = '{:%H:%M}'.format(self.start)
        span = f'{start}-{self.stop:%H:%M}' if self.stop else start
        fields = ' '.join('{}={}'.format(f, getattr(self, f))
                          for f in ('type', 'task', 'customer', 'description'))
        return f'<TimeEntry {span} {fields}>'

    def duration(self, allow_unended=False):
        if not self.stop:
            if not allow_unended:
                raise Exception('Unstopped time entries cannot have a duration.')
            else:
                self.stop = datetime.now()

        return self.stop - self.start

    def __iter__(self):
        yield from {
            'start': self.start.strftime('%H:%M'),
            'stop': self.stop.strftime('%H:%M') if self.stop else None,
            'type': self.type,
            'task': self.task,
            'customer': self.customer,
            'description': self.description,
        }.items()

    @staticmethod
    def start(start, **kwargs):
        entries = EntryList(start.date())

        # Stop the previous time entry if it exists.
        if len(entries) > 0:
            entries[-1].stop = start

        entries.append(TimeEntry(start, **kwargs))
        entries.save()

    @staticmethod
    def stop(stop, **kwargs):
        entries = EntryList(stop.date())
        if len(entries) == 0:
            print('No time entry to end.')
        else:
            entries[-1].stop = stop
            entries.save()
