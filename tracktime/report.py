import csv
import os
from calendar import Calendar
from collections import OrderedDict, defaultdict
from datetime import date, datetime, timedelta

import pdfkit
from docutils import core
from docutils.writers import html5_polyglot
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
        rates = self.configuration['project_rates']

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

    def generate_textual_report(self, tablefmt):
        time_report_header = 'Time Report - {:%B %Y}'.format(self.month)
        lines = [
            time_report_header,
            '=' * len(time_report_header),
            '',
            f'**User:** {self.fullname}',
            '',
        ]

        # If there's a customer, then add it to the report.
        if self.customer:
            aliases = self.configuration['customer_aliases']
            addresses = self.configuration['customer_addresses']
            addr_lines = [
                aliases.get(self.customer, self.customer),
                *addresses.get(self.customer, '').strip().split('\n'),
            ]
            customer = ''
            for line in addr_lines:
                customer += '    | {}\n'.format(line)

            lines += [
                f'**Customer:**',
                '',
                customer,
            ]

        # Include the report table
        lines += [
            '**Detailed Time Report:**',
            '',
            tabulate(
                self.report_table,
                headers='keys',
                floatfmt='.2f',
                tablefmt=tablefmt),
        ]

        return '\n'.join(lines)

    def generate_html_report(self):
        rst = self.generate_textual_report('rst')
        html = core.publish_string(rst, writer=html5_polyglot.Writer())
        return html.decode('utf-8')

    def export_to_stdout(self):
        tablefmt = self.configuration['tableformat']
        text = self.generate_textual_report(tablefmt)
        print(text.replace('| ', '').replace('**', ''))

    def export_to_html(self, filename):
        with open(filename, 'w+') as f:
            f.write(self.generate_html_report())

        print(f'HTML report exported to {filename}.')

    def export_to_pdf(self, filename):
        pdfkit.from_string(self.generate_html_report(), filename)
        print(f'PDF report exported to {filename}.')
