import calendar

from datetime import timedelta
from pathlib import Path
from typing import DefaultDict, Tuple, Dict

import pdfkit
import tabulate

from docutils import core
from docutils.writers import html5_polyglot
from tracktime import EntryList, config


class EntrySet(set):
    @property
    def minutes(self):
        return sum(x.duration(False) for x in self)


class ReportDict(DefaultDict):
    @property
    def minutes(self):
        return sum(v.minutes for v in self.values())


class Report:
    def date_range(self, start, stop):
        current = start
        while current <= stop:
            yield current
            current += timedelta(days=1)

    def __init__(
        self,
        start_date,
        end_date,
        customer,
        project,
        task_grain,
        description_grain,
    ):
        self.start_date = start_date
        self.end_date = end_date
        self.customer = customer
        self.project = project
        self.task_grain = task_grain
        self.description_grain = description_grain
        self.configuration = config.get_config()
        self.max_customer_project_chars = 0
        self.max_task_chars = 0
        self.max_description_chars = 0

        # report_map[(customer, project)][task][description] = set(TimeEntry)
        self.report_map: ReportDict[Tuple[str, str], ReportDict[
            str, ReportDict[str, set]]] = ReportDict(
                lambda: ReportDict(lambda: ReportDict(EntrySet)))

        # Iterate through all of the days covered by this report.
        for day in self.date_range(start_date, end_date):
            for entry in EntryList(day):
                self.report_map[(
                    entry.customer,
                    entry.project,
                )][entry.taskid][entry.description.upper()].add(entry)

                self.max_customer_project_chars = max(
                    self.max_customer_project_chars,
                    len(
                        self.customer_project_str(
                            entry.customer,
                            entry.project,
                        )))
                self.max_task_chars = max(
                    self.max_task_chars,
                    len(entry.taskid),
                )
                self.max_description_chars = max(
                    self.max_description_chars,
                    len(entry.description),
                )

        self.rate_totals_map: Dict[Tuple[str, str], Tuple[float, float]] = {}
        for customer, project in self.report_map:
            rate = 0
            if customer:
                rate = self.configuration['customer_rates'].get(customer, rate)

            if project:
                rate = self.configuration['project_rates'].get(project, rate)

            total = self.report_map[(customer, project)].minutes / 60 * rate
            self.rate_totals_map[(customer, project)] = (rate, total)

    def customer_project_str(self, customer, project):
        if not customer and not project:
            return '<no project or customer>'
        if customer and project:
            return f'{customer}: {project}'
        return customer or project

    def to_hours(self, minutes):
        return minutes / 60

    def generate_textual_report(self, tablefmt):
        # Format the header.
        time_report_header = 'Time Report: {} - {}'.format(
            self.start_date, self.end_date)
        if self.start_date.year == self.end_date.year:
            if (self.start_date.month == 1 and self.start_date.day == 1
                    and self.end_date.month == 12 and self.end_date.day == 31):
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
            f"**User:** {self.configuration.get('fullname')}",
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
        grand_total = sum(rt[1] for rt in self.rate_totals_map.values())
        lines.append(f'**Grand Total:** ${grand_total:.2f}')
        lines.append('')

        # Include the report table
        def ellipsize(string, length=40):
            if len(string) > 40:
                return string[:37] + '...'
            return string

        def pad_tabulate(rows, headers=None, **kwargs):
            tabulate.PRESERVE_WHITESPACE = True
            real_headers = headers or ['', '', '', '']
            real_headers = [
                ellipsize(real_headers[0]),
                *(s.rjust(10) for s in real_headers[1:]),
            ]
            table = tabulate.tabulate(
                [[
                    ellipsize(desc).ljust(40),
                    self.to_hours(minutes),
                    rate,
                    total,
                ] for (desc, minutes, rate, total) in rows],
                tablefmt=tablefmt,
                floatfmt='.2f',
                numalign=None,
                colalign=('left', 'right', 'right', 'right'),
                headers=real_headers,
                **kwargs,
            )
            # Need to remove the headers if they weren't specified.
            if headers is None:
                lines = table.split('\n')
                table = '\n'.join([lines[0], *lines[3:]])

            return table

        def pad_entry(text, minutes, indent_level=0):
            return (
                ellipsize(' ' *
                          (1 + indent_level * 2) + ' * ' + text).ljust(40) +
                ' ' * 7 + '{:.2f}'.format(self.to_hours(minutes)).rjust(10))

        lines += [
            '**Detailed Time Report:**',
            '',
            pad_tabulate(
                [[
                    'TOTAL',
                    self.report_map.minutes,
                    '',
                    grand_total,
                ]],
                headers=['', 'Hours', 'Rate ($/h)', 'Total ($)'],
            ),
            '',
        ]

        for (i, ((customer, project),
                 tasks)) in enumerate(self.report_map.items()):
            if i > 0:
                lines.append('')
            lines.append(
                pad_tabulate([[
                    self.customer_project_str(customer, project),
                    tasks.minutes,
                    *self.rate_totals_map[(customer, project)],
                ]]))

            if not self.task_grain:
                continue

            lines.append('')

            for task_name, task_descriptions in tasks.items():
                task_name = task_name or '<NO TASK>'
                lines.append(pad_entry(task_name, task_descriptions.minutes))

                if not self.description_grain:
                    continue

                # Skip the <NO DESCRIPTION> if that's the only one
                if (len(task_descriptions) == 1
                        and '' in task_descriptions.keys()):
                    lines.append('')
                    continue

                lines.append('')

                for description, entries in task_descriptions.items():
                    description = description or '<NO DESCRIPTION>'
                    lines.append(
                        pad_entry(
                            description,
                            entries.minutes,
                            indent_level=1,
                        ))

                lines.append('')

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


class ReportExporter:
    def __init__(self, report: Report):
        self.report = report

    def export(self, path: Path):
        raise NotImplementedError(
            'Inheritors of ReportExporter must implement ``export``.')


class PDFExporter(ReportExporter):
    def export(self, path: Path):
        pass


class HTMLExporter(ReportExporter):
    def export(self, path: Path):
        pass


class RSTExporter(ReportExporter):
    def export(self, path: Path):
        pass


class StdoutExporter(ReportExporter):
    def export(self, path: Path):
        tablefmt = config.get_config()['tableformat']
        print(self.report.generate_textual_report(tablefmt))


report_exporters = {
    'pdf': PDFExporter,
    'html': HTMLExporter,
    'rst': RSTExporter,
    'stdout': StdoutExporter,
}
