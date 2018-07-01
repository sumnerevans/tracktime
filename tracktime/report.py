import csv
import os
from calendar import Calendar
from collections import OrderedDict, defaultdict
from datetime import date, datetime, timedelta

from tabulate import tabulate

from tracktime import EntryList, config


class Report:
    class Project:
        def __init__(self, name=None):
            self.name = name
            self.customer = None
            self.rate = 0
            self.minutes = 0

        def __repr__(self):
            return '<Report.Project {} customer={} rate={} minutes={}>'.format(
                self.name, self.customer, self.rate, self.minutes)

        def add_minutes(self, minutes):
            self.minutes += minutes

        @property
        def total(self):
            return self.minutes / 60 * self.rate

        def get_dict(self, show_customer):
            details = OrderedDict()
            details['Project'] = self.name

            if show_customer:
                details['Customer'] = self.customer

            details['Hours'] = float(self.minutes / 60)
            details['Rate ($)'] = float(self.rate)
            details['Total ($)'] = self.total

            return details

    def __init__(self, start, customer):
        self.month = start
        self.customer = customer

        # Pull from config
        self.configuration = config.get_config()
        self.fullname = self.configuration['fullname']
        rates = self.configuration.get('project_rates', {})

        entry_groups = defaultdict(Report.Project)
        total_minutes = 0

        # Iterate through all of the days of the month for the report.
        for day in Calendar().itermonthdays(start.year, start.month):
            if day < 1:
                # Not sure why this happens, I think it's so that it can avoid
                # half-weeks.
                continue

            for entry in EntryList(date(start.year, start.month, day)):
                # Filter by customer. If customer is null, include everything.
                if customer and entry.customer != customer:
                    continue

                # Determine what group this entry belogs in
                if entry.project:
                    group = entry_groups[entry.project]

                    # Verify that the customer matches the previous entries.
                    if group.customer and group.customer != entry.customer:
                        raise Exception('Two entries with the same project but'
                                        ' different customers.')
                elif entry.customer:
                    group = entry_groups[entry.customer]
                else:
                    group = entry_groups[None]

                # Add the information about this entry to the appropriate
                # group.
                group.name = entry.project
                group.customer = entry.customer
                group.minutes += entry.duration()
                group.rate = rates.get(entry.project, 0)

                total_minutes += entry.duration()

        if len(entry_groups) == 0:
            raise Exception(f'No entries found for customer "{customer}".')

        self.report_table = [
            row.get_dict(self.customer is None)
            for row in entry_groups.values()
        ]

        # Total Line
        self.report_table.append({
            'Project': 'TOTAL',
            'Hours': total_minutes / 60,
            'Total ($)': sum(row['Total ($)'] for row in self.report_table),
        })

    def export_to_stdout(self):
        table = tabulate(
            self.report_table,
            headers='keys',
            floatfmt='.2f',
            tablefmt=self.configuration.get('tableformat', 'simple'))
        width = max(72, *(len(line) for line in table.splitlines()))

        print('TIME REPORT{}{:>20}'.format(' ' * (width - 31),
                                           self.month.strftime('%B %Y')))
        print('=' * width)
        print()
        print('User:', self.fullname)
        print()
        print('Customer:')
        if self.customer:
            addresses = self.configuration.get('customer_addresses', {})
            aliases = self.configuration.get('customer_aliases', {})
            lines = [
                aliases.get(self.customer, self.customer),
                *addresses.get(self.customer, '').split('\n'),
            ]
            for line in lines:
                print('    {}'.format(line))
        else:
            print('Internal Report\n')

        print('\nDetailed Time Report:\n')
        print(table)

    def export_to_pdf(self):
        print('PDF export')
        print(self)
