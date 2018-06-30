import csv
import os
from calendar import Calendar
from datetime import date, datetime, timedelta

import tracktime
from tracktime import EntryList
from tracktime.time_parser import parse_month


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

    def export_to_stdout(self):
        print('stdout export')
        print(self.entries)

    def export_to_pdf(self):
        print('PDF export')
        print(self.entries)
