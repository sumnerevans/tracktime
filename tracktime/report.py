import csv
import os
from calendar import Calendar, month_abbr, month_name
from datetime import date, datetime, timedelta

import tracktime
from tracktime import EntryList


class Report:
    def __init__(self, start, customer):
        self.entries = []
        # Iterate through all of the days of the month for the report.
        for day in Calendar().itermonthdays(start.year, start.month):
            if day < 1:
                # Not sure why this happens, I think it's so that it can avoid
                # half-weeks.
                continue

            day = datetime(start.year, start.month, day)
            for x in EntryList(day):
                # Filter by customer. If customer is null, include everything.
                if customer and x.customer != customer:
                    continue
                self.entries.append(x)

    @staticmethod
    def parse_month(month_str):
        # If the month is specified as a string...
        abbrs = list(month_abbr)
        names = list(month_name)
        if month_str in abbrs:
            month = abbrs.index(month_str)

        elif month_str in names:
            month = names.index(month_str)
        else:
            # If the month is specified as a numeric string...
            try:
                month = int(month_str)
            except ValueError as e:
                raise ValueError('You must specify the month as either the '
                                 'fully qualified month (December), an '
                                 'abbreviated month (Dec), or a number.') from e

        return month

    def export_to_stdout(self):
        print('stdout export')
        print(self.entries)

    def export_to_pdf(self):
        print('PDF export')
        print(self.entries)

    @staticmethod
    def create(filename, month, year, customer, **kwargs):
        if year and not month:
            raise ValueError('Must specify month when specifying year for report.')

        if year:
            start = date(int(year), Report.parse_month(month), 1)
        else:
            now = datetime.today().date()
            if not month:
                # Default to previous month
                if now.month == 1:  # It's January, default to last December
                    start = date(now.year - 1, 12, 1)
                else:
                    start = date(now.year, now.month - 1, 1)
            else:
                start = date(now.year, Report.parse_month(month), 1)

        report = Report(start, customer)
        if filename:
            report.export_to_pdf(filename)
        else:
            report.export_to_stdout()
