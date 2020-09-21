"""Time Entry class"""
from datetime import datetime


class TimeEntry:
    def __init__(
        self,
        start,
        description,
        stop=None,
        type=None,
        project=None,
        taskid=None,
        customer=None,
    ):
        self.start = start
        self.description = description
        self.stop = stop
        self.project = project
        self.taskid = taskid
        self.customer = customer

        self.type = {
            "gl": "gitlab",
            "gh": "github",
        }.get(type.lower() if type else None, type)

    @property
    def stop(self):
        return self._stop

    @stop.setter
    def stop(self, value):
        """Stops the time entry at the given stop time."""
        if value and value < self.start:
            raise Exception("Cannot stop a time entry before it was started.")
        self._stop = value

    def __repr__(self):
        """Return the string representation of a TimeEntry.

        >>> start = datetime(2018, 1, 1, 13, 11)
        >>> TimeEntry(
        ...     start, 'cool', type='gl', project='foo', taskid='#3', customer='bar'
        ... )  # doctest: +NORMALIZE_WHITESPACE
        <TimeEntry 13:11 project=foo type=gitlab taskid=#3 customer=bar
            description=cool>
        """
        start = "{:%H:%M}".format(self.start)
        span = f"{start}-{self.stop:%H:%M}" if self.stop else start
        fields = ("project", "type", "taskid", "customer", "description")
        fields = " ".join("{}={}".format(f, getattr(self, f)) for f in fields)
        return f"<TimeEntry {span} {fields}>"

    def duration(self, allow_unended=False):
        """Return the duration of this time entry.

        Arguments:
        allow_unended: (optional) whether or not to allow the entry to be
                       unended

        >>> start = datetime(2018, 1, 1, 10, 13)
        >>> stop = datetime(2018, 1, 1, 11, 23)
        >>> TimeEntry(start, '', stop=stop).duration()
        70
        """
        if not self.stop:
            if not allow_unended:
                raise Exception("Unended time entries cannot have a duration.")
            else:
                self.stop = datetime.now()

        return (self.stop - self.start).seconds // 60

    def __iter__(self):
        yield from {
            "start": self.start.strftime("%H:%M"),
            "stop": self.stop.strftime("%H:%M") if self.stop else None,
            "project": self.project,
            "type": self.type,
            "taskid": self.taskid,
            "customer": self.customer,
            "description": self.description,
        }.items()
