import calendar

from collections import defaultdict
from datetime import date, timedelta
from pathlib import Path
from typing import DefaultDict, Tuple, Dict

import pdfkit  # type: ignore
import tabulate

from tracktime import EntryList
from tracktime.synchronisers.base import Synchroniser


class EntrySet(set):
    @property
    def minutes(self):
        return sum(x.duration(False) for x in self)


class ReportDict(DefaultDict):
    def __init__(self, default_factory, sort, reverse):
        super().__init__(default_factory)
        self.sort = sort
        self.reverse = reverse

    @property
    def minutes(self):
        return sum(v.minutes for v in self.values())

    def items(self):
        def sorter(kvp):
            if self.sort == Report.SortType.ALPHABETICAL:
                return "".join(kvp[0]).lower()
            else:
                return kvp[1].minutes

        yield from sorted(super().items(), key=sorter, reverse=self.reverse)


class ReportTimeStatistics:
    days_worked: int
    average_time_per_day_worked: float
    average_time_per_weekday_worked: float
    average_time_per_week_worked: float

    def __init__(self, report):
        def mean(numbers):
            return sum(numbers) / len(numbers)

        # Only considers days where more than the day_worked_min_threshold for
        # minutes worked was surpassed. This avoids counting days where you work
        # for a few minutes. The same thing for weeks, except multiplied by 7.
        day_worked_threshold = report.config.get("day_worked_min_threshold")

        # Per day stats
        days_worked = {
            d: m for d, m in report.day_stats.items() if m >= day_worked_threshold
        }
        total_minutes_worked = sum(report.day_stats.values())

        self.days_worked = len(days_worked)
        self.weekdays_worked = len(
            [1 for d, m in days_worked.items() if d.weekday() < 5]
        )
        self.average_time_per_day_worked = (
            0 if self.days_worked == 0 else (total_minutes_worked / self.days_worked)
        )
        self.average_time_per_weekday_worked = (
            0
            if self.weekdays_worked == 0
            else (total_minutes_worked / self.weekdays_worked)
        )

        # Per week stats
        week_stats = defaultdict(int)
        for d, m in report.day_stats.items():
            week_stats[d.isocalendar()[1]] += m

        self.weeks_worked = len(days_worked) / 5
        self.average_time_per_week_worked = (
            0 if self.weeks_worked == 0 else (total_minutes_worked / self.weeks_worked)
        )

    def _format_hours(self, minutes: float) -> str:
        minutes = round(minutes)
        hours = minutes // 60
        minutes = minutes % 60
        minutes_str = f"0{minutes}" if minutes < 10 else str(minutes)
        return f"{hours}:{minutes_str}"

    @property
    def avg_per_day(self):
        return self._format_hours(self.average_time_per_day_worked)

    @property
    def avg_per_weekday(self):
        return self._format_hours(self.average_time_per_weekday_worked)

    @property
    def avg_per_week(self):
        return self._format_hours(self.average_time_per_week_worked)

    @property
    def statistics_dictionary(self):
        return {
            "Days worked": self.days_worked,
            "Average time per day worked": self.avg_per_day,
            "Average time per weekday worked": self.avg_per_weekday,
            "Weeks* worked": self.weeks_worked,
            "Average time per week* worked": self.avg_per_week,
        }


class Report:
    class SortType:
        ALPHABETICAL = 0
        TIME_SPENT = 1

    class SortDirection:
        ASCENDING = 0
        DESCENDING = 1

    def date_range(self, start, stop):
        current = start
        while current <= stop:
            yield current
            current += timedelta(days=1)

    def __init__(
        self,
        config,
        start_date,
        end_date,
        sort,
        sort_direction,
        customer,
        project,
        task_grain,
        description_grain,
    ):
        self.start_date = start_date
        self.end_date = end_date
        self.customer = customer
        self.project = project
        self.sort = sort
        self.reverse = sort_direction == Report.SortDirection.DESCENDING
        self.task_grain = task_grain or description_grain
        self.description_grain = description_grain
        self.config = config
        self.synchroniser = Synchroniser(self.config)

        # report_map[(customer, project)][task][description] = set(TimeEntry)
        self.report_map: ReportDict[
            Tuple[str, str], ReportDict[str, ReportDict[str, EntrySet]]
        ] = ReportDict(
            lambda: ReportDict(
                lambda: ReportDict(
                    EntrySet,
                    self.sort,
                    self.reverse,
                ),
                self.sort,
                self.reverse,
            ),
            self.sort,
            self.reverse,
        )

        self.day_stats: DefaultDict[date, int] = defaultdict(int)

        # Iterate through all of the days covered by this report.
        for day in self.date_range(start_date, end_date):
            for entry in EntryList(self.config, day):
                if not entry.stop:
                    raise Exception(f"ERROR: Unended time entry on {day}")

                if self.customer and entry.customer != self.customer:
                    continue
                if self.project and entry.project != self.project:
                    continue

                self.day_stats[day] += entry.duration()

                self.report_map[(entry.customer, entry.project,)][entry.taskid][
                    entry.description.upper()
                ].add(entry)

        self.rate_totals_map: Dict[Tuple[str, str], Tuple[float, float]] = {}
        for customer, project in self.report_map:
            rate = 0
            if customer:
                rate = self.config["customer_rates"].get(customer, rate)

            if project:
                rate = self.config["project_rates"].get(project, rate)

            total = self.report_map[(customer, project)].minutes / 60 * rate
            self.rate_totals_map[(customer, project)] = (rate, total)

        self.stats = ReportTimeStatistics(self)

    def customer_project_str(self, customer, project, html=False):
        if not customer and not project:
            return (
                "<no project or customer>"
                if not html
                else "<i>no project or customer</i>"
            )
        if customer and project:
            return f"{customer}: {project}"
        return customer or project

    def to_hours(self, minutes):
        return minutes / 60

    def round(self, val) -> str:
        return "{:.2f}".format(round(val, 2))

    @property
    def header_text(self) -> str:
        time_report_header = "Time Report: {} - {}".format(
            self.start_date, self.end_date
        )
        if self.start_date.year == self.end_date.year:
            if (
                self.start_date.month == 1
                and self.start_date.day == 1
                and self.end_date.month == 12
                and self.end_date.day == 31
            ):
                # Reporting on the whole year.
                time_report_header = f"Time Report: {self.start_date.year}"
            elif self.start_date.month == self.end_date.month:
                if (
                    self.start_date.day == 1
                    and self.end_date.day
                    == calendar.monthrange(self.start_date.year, self.start_date.month)[
                        1
                    ]
                ):
                    # Reporting on a single month.
                    time_report_header = "Time Report: {:%B %Y}".format(self.start_date)
                elif self.start_date.day == self.end_date.day:
                    time_report_header = f"Time Report: {self.start_date}"
        return time_report_header

    @property
    def grand_total(self) -> float:
        return sum(rt[1] for rt in self.rate_totals_map.values())

    @property
    def address_lines(self):
        aliases = self.config["customer_aliases"]
        addresses = self.config["customer_addresses"]
        return [
            aliases.get(self.customer, self.customer),
            *addresses.get(self.customer, "").strip().split("\n"),
        ]

    def generate_textual_report(self, tablefmt):
        # Format the header.
        lines = [
            self.header_text,
            "=" * len(self.header_text),
            "",
            f"**User:** {self.config.get('fullname')}",
            "",
        ]

        # If there's a customer, then add it to the report.
        if self.customer:
            customer = ""
            for line in self.address_lines:
                customer += "    | {}\n".format(line)

            lines += [
                "**Customer:**",
                "",
                customer,
            ]

        # Include the Grand Total
        lines.append(f"**Grand Total:** ${self.round(self.grand_total)}")
        lines.append("")

        lines += [
            "**Statistics:**",
            "",
        ]
        statistics = self.stats.statistics_dictionary
        max_desc_length = max(map(len, statistics.keys()))
        for desc, val in statistics.items():
            desc = desc + ":"
            lines.append(f"    | {desc.ljust(max_desc_length+2)}{val}")
        lines.append("")
        lines.append(
            "* a week is any set of five days (not necessarily within the same "
            "calendar week)"
        )
        lines.append("")

        # Include the report table
        def ellipsize(string, length=40):
            if len(string) > 40:
                return string[:37] + "..."
            return string

        def pad_tabulate(rows, headers=None, **kwargs):
            tabulate.PRESERVE_WHITESPACE = True
            real_headers = headers or ["", "", "", ""]
            real_headers = [
                ellipsize(real_headers[0]),
                *(s.rjust(10) for s in real_headers[1:]),
            ]
            table = tabulate.tabulate(
                [
                    [ellipsize(desc).ljust(40), self.to_hours(minutes), rate, total]
                    for (desc, minutes, rate, total) in rows
                ],
                tablefmt=tablefmt,
                floatfmt=".2f",
                numalign=None,
                colalign=("left", "right", "right", "right"),
                headers=real_headers,
                **kwargs,
            )
            # Need to remove the headers if they weren't specified.
            if headers is None:
                lines = table.split("\n")
                table = "\n".join([lines[0], *lines[3:]])

            return table

        def pad_entry(text, minutes, indent_level=0):
            return (
                ellipsize(" " * (1 + indent_level * 2) + " * " + text).ljust(40)
                + " " * 7
                + self.round(self.to_hours(minutes)).rjust(10)
            )

        lines += [
            "**Detailed Time Report:**",
            "",
            pad_tabulate(
                [["TOTAL", self.report_map.minutes, "", self.grand_total]],
                headers=["", "Hours", "Rate ($/h)", "Total ($)"],
            ),
            "",
        ]

        for (i, ((customer, project), tasks)) in enumerate(self.report_map.items()):
            if i > 0:
                lines.append("")
            lines.append(
                pad_tabulate(
                    [
                        [
                            self.customer_project_str(customer, project),
                            tasks.minutes,
                            *self.rate_totals_map[(customer, project)],
                        ]
                    ]
                )
            )

            if not self.task_grain:
                continue

            lines.append("")

            for task_name, task_descriptions in tasks.items():
                first_entry = list(list(task_descriptions.values())[0])[0]
                task_name = (
                    self.synchroniser.get_formatted_task_id(first_entry) or "<NO TASK>"
                )
                desc = self.synchroniser.get_task_description(first_entry)
                if desc:
                    task_name += f": {desc.upper()}"
                lines.append(pad_entry(task_name, task_descriptions.minutes))

                if not self.description_grain:
                    continue

                # Skip the <NO DESCRIPTION> if that's the only one
                if len(task_descriptions) == 1 and "" in task_descriptions.keys():
                    lines.append("")
                    continue

                lines.append("")

                for description, entries in task_descriptions.items():
                    description = description or "<NO DESCRIPTION>"
                    lines.append(
                        pad_entry(description, entries.minutes, indent_level=1)
                    )

                lines.append("")

        return "\n".join(lines)

    def generate_html_report(self):
        styles = """
        body {
          background-color: white;
        }

        .content {
          max-width: 900px;
          margin: 0 auto;
        }

        .customer-address {
          padding: 10px 40px;
        }

        table {
          border-collapse: collapse;
        }

        table.statistics-table {
          margin-left: 40px;
        }

        thead th, tbody tr.customer-project td, tbody tr.total td {
          font-weight: bold;
          min-width: 100px;
          padding: 10px;
          border: 1px solid black;
        }

        tbody tr.spacer td {
            padding: 5px;
        }

        tbody td {
          padding: 3px 10px 0 0;
        }

        tbody tr td h3 {
          margin: 0;
        }

        tbody.detailed-time-report-body tr td:first-child {
          width: 526px;
          max-width: 526px;
        }

        tbody td li {
          list-style-position:inside;
          overflow: hidden;
          white-space: nowrap;
          text-overflow: ellipsis;
        }
        """

        # If there's a customer, then add it to the report.
        customer_html = ""
        if self.customer:
            customer_html = f"""
            <tr><td><b>Customer:</b></td></tr>
            <tr>
              <td colspan="2" class="customer-address">
                {'<br/>'.join(self.address_lines)}
              </td>
            </tr>
            """

        statistics_html = """
        <h3>Statistics</h3>
        <table class="statistics-table">
        """
        for desc, val in self.stats.statistics_dictionary.items():
            statistics_html += f"""
            <tr>
                <td><b>{desc}:</b></td>
                <td>{val}</td>
            </tr>
            """
        statistics_html += "</table>"
        statistics_html += """<i>
            * a week is any set of five days (not necessarily within the same
            calendar week)
        </i>"""

        data = [
            (
                "total",
                "<b>TOTAL</b>",
                self.round(self.to_hours(self.report_map.minutes)),
                "",
                self.round(self.grand_total),
            ),
        ]

        for (i, ((customer, project), tasks)) in enumerate(self.report_map.items()):
            rate, total = self.rate_totals_map[(customer, project)]
            data.append(("spacer", "", ""))
            data.append(
                (
                    "customer-project",
                    self.customer_project_str(customer, project, html=True),
                    self.round(self.to_hours(tasks.minutes)),
                    rate,
                    self.round(total),
                )
            )

            if not self.task_grain:
                continue

            for task_name, task_descriptions in tasks.items():
                first_entry = list(list(task_descriptions.values())[0])[0]
                task_name = (
                    self.synchroniser.get_formatted_task_id(first_entry)
                    or "<i>NO TASK</i>"
                )
                desc = self.synchroniser.get_task_description(first_entry)
                if desc:
                    task_name += f": {desc.upper()}"
                link = self.synchroniser.get_task_link(first_entry)
                if link:
                    task_name = f'<a href={link} target="_blank">{task_name}</a>'

                data.append(
                    (
                        "task",
                        f"""<ul style="margin: 0; padding-left: 30px;">
                          <li>{task_name}</li>
                        </ul>""",
                        self.round(self.to_hours(task_descriptions.minutes)),
                    )
                )

                if not self.description_grain:
                    continue

                # Skip the <NO DESCRIPTION> if that's the only one
                if len(task_descriptions) == 1 and "" in task_descriptions.keys():
                    continue

                for description, entries in task_descriptions.items():
                    description = description or "<i>NO DESCRIPTION</i>"
                    data.append(
                        (
                            "description",
                            f"""<ul style="margin: 0; padding-left: 50px;">
                              <li title="{description}">{description}</li>
                            </ul>""",
                            self.round(self.to_hours(entries.minutes)),
                        )
                    )

        table_body = ""

        for c, *cells in data:
            table_body += f'<tr class="{c}">'
            for i, cell in enumerate(cells):
                align = "right" if i > 0 else "left"
                table_body += f'<td style="text-align: {align};">{cell}</td>'
            table_body += "</tr>"

        return f"""<!doctype html>
        <html>
          <head>
            <title>{self.header_text}</title>
            <style type="text/css">{styles}</style>
          </head>
          <body>
            <div class="content">
              <h1 style="text-align: center;">{self.header_text}</h1>
              <table>
                <tr>
                  <td><b>User:</b></td>
                  <td>{self.config.get('fullname')}</td>
                </tr>
                {customer_html}
                <tr>
                  <td><h3>Grand Total:</h3></td>
                  <td><h3>${self.round(self.grand_total)}</h3></td>
                </tr>
              </table>
              {statistics_html}
              <h3>Detailed Time Report</h3>
              <table>
                <thead>
                  <th></th>
                  <th style="text-align: right;">Hours</th>
                  <th style="text-align: right;">Rate ($/h)</th>
                  <th style="text-align: right;">Total ($)</th>
                </thead>
                <tbody class="detailed-time-report-body">
                  {table_body}
                </tbody>
              </table>
            </div>
          </body>
        </html>
        """


class ReportExporter:
    def __init__(self, config, report: Report):
        self.config = config
        self.report = report

    def export(self, path: Path):
        raise NotImplementedError(
            "Inheritors of ReportExporter must implement ``export``."
        )


class PDFExporter(ReportExporter):
    def export(self, path: Path):
        pdfkit.from_string(self.report.generate_html_report(), str(path))
        print(f"PDF report exported to {path}.")


class HTMLExporter(ReportExporter):
    def export(self, path: Path):
        with open(path, "w+") as f:
            f.write(self.report.generate_html_report())

        print(f"HTML report exported to {path}.")


class RSTExporter(ReportExporter):
    def export(self, path: Path):
        with open(path, "w+") as f:
            f.write(self.report.generate_textual_report("rst"))


class StdoutExporter(ReportExporter):
    def export(self, path: Path):
        tablefmt = self.config["tableformat"]
        text = self.report.generate_textual_report(tablefmt)
        print(text.replace("| ", "").replace("**", ""))


report_exporters = {
    "pdf": PDFExporter,
    "html": HTMLExporter,
    "rst": RSTExporter,
    "stdout": StdoutExporter,
}
