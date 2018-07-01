import re
import time
from calendar import day_abbr, day_name, month_abbr, month_name
from datetime import date, datetime, timedelta


def parse_time(time_representation, date=datetime.now()):
    """Parse a time string. If a string is not passed, just return what was
    passed.

    Arguments:

    :param time_representation: the representation of the time to parse
    :param date: (optional) the date to use in the datetime
    :returns: a datetime representing the time and date
    """
    if not isinstance(time_representation, str):
        return time_representation

    time_formats = {
        r'\d{4}': '%H%M',
        r'\d\d:\d\d': '%H:%M',
    }

    for pattern, dateformat in time_formats.items():
        if re.match(pattern, time_representation):
            timestamp = time.strptime(time_representation, dateformat)
            return datetime(date.year, date.month, date.day,
                            timestamp.tm_hour, timestamp.tm_min)

    raise ValueError('Could not parse time.')


def parse_date(date_representation):
    """Parse a date string. If a string is not passed, just return what was
    passed.

    Arguments:

    :param date_representation: the representation of the time to parse
    :returns: a datetime representing midnight on the proper date
    """
    if not isinstance(date_representation, str):
        return date_representation

    date_formats = {
        r'\d{4}-\d{2}-\d{2}': '%Y-%m-%d',
        r'\d\d-\d\d': '%m-%d',
        r'\d\d': '%d',
    }

    for pattern, dateformat in date_formats.items():
        if re.match(pattern, date_representation):
            dt = datetime.strptime(date_representation, dateformat)

            # Set defaults
            now = datetime.now()
            if r'%Y' not in dateformat:
                dt = dt.replace(year=now.year)
            if r'%m' not in dateformat:
                dt = dt.replace(month=now.month)

            return dt

    if date_representation == 'today':
        return date.today()
    elif date_representation == 'yesterday':
        return date.today() - timedelta(1)
    elif date_representation in list(day_abbr):
        # Go back a week and iterate until you find a match.
        start = date.today() - timedelta(6)
        while start.strftime('%a') != date_representation:
            start += timedelta(1)
        return start
    elif date_representation in list(day_name):
        # Go back a week and iterate until you find a match.
        start = date.today() - timedelta(6)
        while start.strftime('%A') != date_representation:
            start += timedelta(1)
        return start

    raise ValueError('Could not parse date.')


def parse_month(month_representation, year=datetime.now().year):
    abbrs = list(month_abbr)  # Jan, Feb, Mar, ...
    names = list(month_name)  # January, February, March, ...
    if month_representation in abbrs:
        month = abbrs.index(month_representation)
    elif month_representation in names:
        month = names.index(month_representation)
    else:
        # The month is specified as a numeric string...
        try:
            month = int(month_representation)
        except ValueError as e:
            raise ValueError('You must specify the month as either the '
                             'fully qualified month (December), an '
                             'abbreviated month (Dec), or a number.') from e

    return month
