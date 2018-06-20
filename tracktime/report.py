from calendar import Calendar, month_abbr, month_name
from datetime import date, datetime, timedelta


class Report:
    @staticmethod
    def create(filename, month, year, **kwargs):
        now = datetime.today().date()
        if not month:
            # Default to previous month
            if now.month == 1:  # It's January, default to last December
                start = date(now.year - 1, 12, 1)
            else:
                start = date(now.year, now.month - 1, 1)
        else:
            # If the month is specified as a string...
            abbrs = list(month_abbr)
            names = list(month_name)
            if month in abbrs:
                month = abbrs.index(month)

            if month in names:
                month = names.index(month)

            # If the month is specified as a numeric string...
            try:
                month = int(month)
            except ValueError as e:
                raise ValueError('You must specify the month as either the '
                                 'fully qualified month (December), an '
                                 'abbreviated month (Dec), or a number.') from e

            start = date(now.year, month, 1)

        days_in_month = max(Calendar().itermonthdays(start.year, start.month))
        end = date(start.year, start.month, days_in_month)

        print(start, end)
        print(filename, kwargs)
