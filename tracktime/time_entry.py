"""Time Entry class"""
import os
import time
from datetime import datetime


class TimeEntry:
    def __init__(self,
                 start,
                 stop=None,
                 type=None,
                 task=None,
                 description=None,
                 customer=None):
        self.start = start
        self.stop = stop
        self.type = type
        self.task = task
        self.description = description
        self.customer = customer

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
