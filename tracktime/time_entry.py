"""Time Entry class"""
import os
import time
from datetime import datetime


class TimeEntry:
    def __init__(self,
                 start,
                 description,
                 stop=None,
                 type=type,
                 project=None,
                 taskid=None,
                 customer=None):
        self.start = start
        self.description = description
        self.stop = stop
        self.type = type
        self.project = project
        self.taskid = taskid
        self.customer = customer

        abbrs = {'gl': 'gitlab', 'gh': 'github'}
        if type in abbrs.keys():
            self.type = abbrs[type]

    @property
    def stop(self):
        return self._stop

    @stop.setter
    def stop(self, value):
        if value and value < self.start:
            raise Exception('Cannot stop a time entry before it was started.')
        self._stop = value

    def __repr__(self):
        start = '{:%H:%M}'.format(self.start)
        span = f'{start}-{self.stop:%H:%M}' if self.stop else start
        fields = ('project', 'type', 'taskid', 'customer', 'description')
        fields = ' '.join('{}={}'.format(f, getattr(self, f)) for f in fields)
        return f'<TimeEntry {span} {fields}>'

    def duration(self, allow_unended=False):
        if not self.stop:
            if not allow_unended:
                raise Exception('Unstopped time entries cannot have a duration.')
            else:
                self.stop = datetime.now()

        return (self.stop - self.start).seconds // 60

    def __iter__(self):
        yield from {
            'start': self.start.strftime('%H:%M'),
            'stop': self.stop.strftime('%H:%M') if self.stop else None,
            'project': self.project,
            'type': self.type,
            'taskid': self.taskid,
            'customer': self.customer,
            'description': self.description,
        }.items()
