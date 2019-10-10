import sys
import calendar
from collections import OrderedDict, defaultdict
from datetime import date, timedelta

import pdfkit
from docutils import core
from docutils.writers import html5_polyglot
from tabulate import tabulate
from tracktime import EntryList, config
from tracktime.time_parser import day_as_ordinal


class Report:
    class Project:
        def __init__(self, name=None):
            self.name = name
            self.customer = None
            self.rate = 0
            self.minutes = 0

        def __repr__(self):
            """Returns a string representation of the Project.

            >>> r = Report.Project(name='foo')
            >>> r.customer = 'bar'
            >>> r.rate = 30
            >>> r.minutes = 84
            >>> r
            <Report.Project foo customer=bar rate=30 minutes=84>
            """
            return '<Report.Project {} customer={} rate={} minutes={}>'.format(
                self.name, self.customer, self.rate, self.minutes)

        def add_minutes(self, minutes):
            """Add a specified number of minutes to the project's time.

            Arguments:
            minutes: the number of minutes to add

            >>> r = Report.Project()
            >>> r.add_minutes(10)
            >>> r.minutes
            10
            """
            self.minutes += minutes

        @property
        def total(self):
            """Calculates the total monetary amount for the project.

            >>> r = Report.Project()
            >>> r.add_minutes(90)
            >>> r.rate = 20
            >>> r.total
            30.0
            """
            return self.minutes / 60 * self.rate

        def get_dict(self, show_customer):
            """Gets the dictionary representation of this Project.

            Arguments:
            show_customer: whether or not to include the customer in the
                           dictionary

            >>> r = Report.Project(name='Test')
            >>> dict(r.get_dict(False))
            {'Project': 'Test', 'Hours': 0.0, 'Rate ($)': 0.0, 'Total ($)': 0.0}
            >>> dict(r.get_dict(True))
            {'Project': 'Test', 'Customer': None, 'Hours': 0.0, 'Rate ($)': 0.0, 'Total ($)': 0.0}
            """
            details = OrderedDict()
            details['Project'] = self.name

            if show_customer:
                details['Customer'] = self.customer

            details['Hours'] = float(self.minutes / 60)
            details['Rate ($)'] = float(self.rate)
            details['Total ($)'] = self.total

            return details

    def date_range(self, start, stop):
        current = start
        while current <= stop:
            yield current
            current += timedelta(days=1)

    def __init__(self, start_date, end_date, customer, project):
        self.start_date = start_date
        self.end_date = end_date
        self.customer = customer
        self.project = project

        if self.customer and self.project:
            raise Exception('You cannot specify both a customer and project '
                            'to report on.')

        # Pull from config
        self.configuration = config.get_config()
        self.fullname = self.configuration['fullname']
        project_rates = self.configuration['project_rates']
        customer_rates = self.configuration['customer_rates']

        entry_groups = defaultdict(Report.Project)
        total_minutes = 0

        # Iterate through all of the days covered by this report.
        for day in self.date_range(start_date, end_date):
            for entry in EntryList(day):
                # Filter by customer. If customer is null, include everything.
                if ((customer and entry.customer != customer)
                        or (project and entry.project != project)):
                    continue

                # Determine what group this entry belogs in.
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
                try:
                    group.minutes += entry.duration()
                except Exception:
                    print(f'Unended time entry on the {day_as_ordinal(day)}.',
                          file=sys.stderr)
                    sys.exit(1)
                group.rate = project_rates.get(
                    entry.project,
                    customer_rates.get(entry.customer, 0),
                )

                total_minutes += entry.duration()

        self.report_table = [
            row.get_dict(self.customer is None)
            for row in entry_groups.values()
        ]

        # Total Line
        self.grand_total = sum(row['Total ($)'] for row in self.report_table)
        self.report_table.append({
            'Project': 'TOTAL',
            'Hours': total_minutes / 60,
            'Total ($)': self.grand_total,
        })

    def generate_textual_report(self, tablefmt):
        time_report_header = 'Time Report: {} - {}'.format(
            self.start_date, self.end_date)
        if self.start_date.year == self.end_date.year:
            if (self.start_date.month == 1 and self.start_date.day == 1
                    and self.end_date.month == 12
                    and self.end_date.day == 31):
                # Reporting on the whole year.
                time_report_header = f'Time Report: {self.start_date.year}'
            elif self.start_date.month == self.end_date.month:
                if (self.start_date.day == 1
                        and self.end_date.day == calendar.monthrange(
                            self.start_date.year, self.start_date.month)[1]):
                    # Reporting on a single month.
                    time_report_header = 'Time Report - {:%B %Y}'.format(
                        self.start_date)
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

        # Include the Grand Total
        lines.append(f'**Grand Total:** ${self.grand_total:.2f}')
        lines.append('')

        # Include the report table
        lines += [
            '**Detailed Time Report:**',
            '',
            tabulate(self.report_table,
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
        pdfkit.from_string(self.generate_html_report(), str(filename))
        print(f'PDF report exported to {filename}.')

    def export_to_rst(self, filename):
        with open(filename, 'w+') as f:
            f.write(self.generate_textual_report('rst'))
