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

    If it's not a string, just pass it through.
    >>> parse_time(datetime(2018, 1, 1, 13, 34))
    datetime.datetime(2018, 1, 1, 13, 34)

    If it is a string, parse it, using ``date`` as the date for the resulting
    datetime.
    >>> parse_time('13:34', date=date(2018, 1, 1))
    datetime.datetime(2018, 1, 1, 13, 34)
    >>> parse_time('1334', date=date(2018, 1, 1))
    datetime.datetime(2018, 1, 1, 13, 34)

    If it is malformed, throw a ValueError
    >>> parse_time('foo', date=date(2018, 1, 1))
    Traceback (most recent call last):
        ...
    ValueError: Could not parse time.
    """
    if not isinstance(time_representation, str):
        return time_representation

    time_formats = {
        r'\d{4}': '%H%M',
        r'\d\d:\d\d': '%H:%M',
    }

    # Try to parse the time using the time formats.
    for pattern, dateformat in time_formats.items():
        if re.match(pattern, time_representation):
            timestamp = time.strptime(time_representation, dateformat)
            return datetime(date.year, date.month, date.day, timestamp.tm_hour,
                            timestamp.tm_min)

    raise ValueError('Could not parse time.')


def parse_date(date_representation):
    """Parse a date string. If a string is not passed, just return what was
    passed.

    Arguments:

    :param date_representation: the representation of the time to parse
    :returns: a datetime representing midnight on the proper date

    >>> now = datetime.now()
    >>> yesterday = datetime.now() - timedelta(days=1)

    If it's not a string, just pass it through.
    >>> parse_date(date(2018, 1, 1))
    datetime.date(2018, 1, 1)

    If it is a string, parse it.
    >>> d = parse_date('03')
    >>> assert (now.year, now.month, 3) == (d.year, d.month, d.day)
    >>> d = parse_date('03-03')
    >>> assert (now.year, 3, 3) == (d.year, d.month, d.day)
    >>> d = parse_date('2018-03-03')
    >>> assert (2018, 3, 3) == (d.year, d.month, d.day)
    >>> d = parse_date('today')
    >>> assert (now.year, now.month, now.day) == (d.year, d.month, d.day)
    >>> d = parse_date('yesterday')
    >>> assert ((yesterday.year, yesterday.month, yesterday.day) ==
    ...         (d.year,         d.month,         d.day))

    If it is malformed, throw a ValueError
    >>> parse_date('foo')
    Traceback (most recent call last):
        ...
    ValueError: Could not parse date.
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


def parse_month(month_representation):
    """Parse a month string.

    Arguments:

    :param month_representation: the representation of the month to parse
    :returns: a datetime representing midnight on the proper date

    If it's numeric, parse as is.
    >>> parse_month('01')
    1

    Otherwise, try to parse it.
    >>> parse_month('Jan')
    1
    >>> parse_month('December')
    12

    If it is malformed, throw a ValueError.
    >>> parse_month('foo')                     # doctest: +NORMALIZE_WHITESPACE
    Traceback (most recent call last):
        ...
    ValueError: You must specify the month as either the fully qualified month
        (December), an abbreviated month (Dec), or a number.
    """
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


def day_as_ordinal(day):
    """Gives the ordinal version of the date.

    Arguments:

    :param date_representation: the representation of the time to parse
    :returns: a string representing the date as an ordinal

    >>> day_as_ordinal(1), day_as_ordinal(2), day_as_ordinal(3)
    ('1st', '2nd', '3rd')
    >>> day_as_ordinal(11), day_as_ordinal(12), day_as_ordinal(13)
    ('11th', '12th', '13th')
    >>> day_as_ordinal(4), day_as_ordinal(20)
    ('4th', '20th')
    """
    suffix = ''
    if 4 <= day <= 20:
        suffix = 'th'
    elif day % 10 == 1:
        suffix = 'st'
    elif day % 10 == 2:
        suffix = 'nd'
    elif day % 10 == 3:
        suffix = 'rd'
    return f'{day}{suffix}'
