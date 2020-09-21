#! /usr/bin/env python3
import os
import argparse
import sys
from datetime import datetime

from tracktime import cli


def main():
    parser = argparse.ArgumentParser(description="Time tracker")

    parser.add_argument(
        "-v",
        "--version",
        help="show version and exit",
        action="store_true",
    )

    default_config_filename = (
        os.environ.get("XDG_CONFIG_HOME")
        or os.environ.get("APPDATA")
        or os.path.join(os.environ.get("HOME", os.path.expanduser("~")), ".config")
    )
    default_config_filename = os.path.join(
        default_config_filename,
        "tracktime/tracktimerc",
    )

    parser.add_argument(
        "--config",
        help="the configuration file to use. Defaults to "
        "~/.config/tracktime/tracktimerc.",
        default=default_config_filename,
    )

    subparsers = parser.add_subparsers(
        dest="action",
        help="specify an action to perform",
    )

    start_parser = subparsers.add_parser(
        "start",
        description="Start a time entry for today.",
    )
    start_parser.add_argument(
        "-s",
        "--start",
        default=datetime.now(),
        help="specify a start time for the time entry (defaults to now)",
    )
    start_parser.add_argument(
        "-t",
        "--type",
        help="specify the type of time entry to start",
    )
    start_parser.add_argument(
        "-p",
        "--project",
        help="specify a project for the time entry",
    )
    start_parser.add_argument(
        "-c",
        "--customer",
        help="specify a customer for the time entry",
    )
    start_parser.add_argument(
        "-i",
        "--taskid",
        help="specify the task being worked on",
    )
    start_parser.add_argument(
        "description",
        help="specify a description for the time entry",
        nargs="?",
    )

    stop_parser = subparsers.add_parser(
        "stop",
        description="Stop the current time entry.",
    )
    stop_parser.add_argument(
        "-s",
        "--stop",
        default=datetime.now(),
        help="specify a stop time for the time entry (defaults to now)",
    )

    resume_parser = subparsers.add_parser(
        "resume",
        description="Resume an entry from today.",
    )
    resume_parser.add_argument(
        "-s",
        "--start",
        default=datetime.now(),
        help="specify the start time for the resumed time entry " "(defaults to now)",
    )
    resume_parser.add_argument(
        "entry",
        type=int,
        nargs="?",
        help="the entry to resume (Python-style indexing, defaults to -1)",
    )

    list_parser = subparsers.add_parser(
        "list",
        description="List the time entries for a date.",
    )
    list_parser.add_argument(
        "-d",
        "--date",
        default=datetime.today().date(),
        help="the date to list time entries for (defaults to today)",
    )
    list_parser.add_argument(
        "-c",
        "--customer",
        help="list only time entries for the given customer",
    )

    edit_parser = subparsers.add_parser(
        "edit",
        description="Edit time entries for a date.",
    )
    edit_parser.add_argument(
        "-d",
        "--date",
        default=datetime.today().date(),
        help="the date to edit time entries for (defaults to today)",
    )

    sync_parser = subparsers.add_parser(
        "sync",
        description="Synchronize time entries for a month.",
    )
    sync_parser.add_argument(
        "month",
        default=datetime.now().month,
        nargs="?",
        help=" ".join(
            [
                "the month to synchronize time entries for (defaults to the",
                "current month, accepted formats: 01, 1, Jan, January, 2019-01)",
            ]
        ),
    )

    report_parser = subparsers.add_parser(
        "report",
        description="Output a report about time spent in a date range.",
    )
    report_parser.add_argument(
        "-m",
        "--month",
        help="shorthand for reporting over an entire month (accepted formats: "
        "01, 1, Jan, January, 2019-01)",
    )
    report_parser.add_argument(
        "-y",
        "--year",
        help="shorthand for reporting over an entire year",
    )
    report_parser.add_argument(
        "--lastweek",
        action="store_true",
        help="shorthand for reporting on last week (Sunday-Saturday)",
    )
    report_parser.add_argument(
        "--thisweek",
        action="store_true",
        help="shorthand for reporting on the current week (Sunday-today)",
    )
    report_parser.add_argument(
        "--thismonth",
        action="store_true",
        help="shorthand for reporting on the current month",
    )
    report_parser.add_argument(
        "--today",
        action="store_true",
        help="shorthand for reporting on today",
    )
    report_parser.add_argument(
        "--yesterday",
        action="store_true",
        help="shorthand for reporting on yesterday",
    )
    report_parser.add_argument(
        "--lastyear",
        action="store_true",
        help="shorthand for reporting on last year",
    )
    report_parser.add_argument(
        "--thisyear",
        action="store_true",
        help="shorthand for reporting on this year",
    )

    task_grain_parser = report_parser.add_mutually_exclusive_group(required=False)
    task_grain_parser.add_argument(
        "--taskgrain",
        dest="taskgrain",
        action="store_true",
        help="report on the task grain",
    )
    task_grain_parser.add_argument(
        "--no-taskgrain",
        dest="taskgrain",
        action="store_false",
        help="do not report on the task grain",
    )
    report_parser.set_defaults(taskgrain="not_specified")

    description_grain_parser = report_parser.add_mutually_exclusive_group(
        required=False
    )
    description_grain_parser.add_argument(
        "--descriptiongrain",
        dest="descriptiongrain",
        action="store_true",
        help="report on the description grain",
    )
    description_grain_parser.add_argument(
        "--no-descriptiongrain",
        dest="descriptiongrain",
        action="store_false",
        help="do not report on the description grain",
    )
    report_parser.set_defaults(descriptiongrain="not_specified")

    report_parser.add_argument(
        "-c",
        "--customer",
        help="customer ID to generate a report for",
    )
    report_parser.add_argument(
        "-p",
        "--project",
        help="project name to generate a report for",
    )
    report_parser.add_argument(
        "-s",
        "--sort",
        choices=["alphabetical", "alpha", "a", "time-spent", "time", "t"],
        help="the grain to sort the report by (defaults to alphabetical)",
    )

    sort_direction_parser = report_parser.add_mutually_exclusive_group(required=False)
    sort_direction_parser.add_argument(
        "-d",
        "--desc",
        action="store_true",
        help="sort descending",
    )
    sort_direction_parser.add_argument(
        "-a",
        "--asc",
        action="store_true",
        help="sort ascending",
    )
    report_parser.set_defaults(desc=False, asc=False)

    report_parser.add_argument(
        "-o",
        "--outfile",
        help="specify the filename to export the report to. "
        "If none specified, output to stdout",
    )
    report_parser.add_argument(
        "range_start",
        nargs="?",
        help="specify the start of the reporting range (defaults to the "
        "beginning of last month)",
    )
    report_parser.add_argument(
        "range_stop",
        nargs="?",
        help="specify the end of the reporting range (defaults to the end of "
        "last month)",
    )

    args = parser.parse_args()
    if not args.action:
        args = parser.parse_args([*sys.argv[1:], "list"])

    # Show version if requested.
    if args.version:
        cli.version()
        return

    {
        "start": cli.start,
        "stop": cli.stop,
        "resume": cli.resume,
        "list": cli.list_entries,
        "edit": cli.edit,
        "sync": cli.sync,
        "report": cli.report,
    }[args.action](args)
