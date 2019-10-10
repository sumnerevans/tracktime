#! /usr/bin/env python3
import argparse
import sys
from datetime import datetime

from tracktime import cli


def main():
    parser = argparse.ArgumentParser(description='Time tracker')

    parser.add_argument(
        '-v',
        '--version',
        help='show version and exit',
        action='store_true',
    )

    subparsers = parser.add_subparsers(
        dest='action',
        help='specify an action to perform',
    )

    start_parser = subparsers.add_parser(
        'start',
        description='Start a time entry for today.',
    )
    start_parser.add_argument(
        '-s',
        '--start',
        default=datetime.now(),
        help='specify a start time for the time entry (defaults to now)',
    )
    start_parser.add_argument(
        '-t',
        '--type',
        help='specify the type of time entry to start',
    )
    start_parser.add_argument(
        '-p',
        '--project',
        help='specify a project for the time entry',
    )
    start_parser.add_argument(
        '-c',
        '--customer',
        help='specify a customer for the time entry',
    )
    start_parser.add_argument(
        '-i',
        '--taskid',
        help='specify the task being worked on',
    )
    start_parser.add_argument(
        'description',
        help='specify a description for the time entry',
        nargs='?',
    )

    stop_parser = subparsers.add_parser(
        'stop',
        description='Stop the current time entry.',
    )
    stop_parser.add_argument(
        '-s',
        '--stop',
        default=datetime.now(),
        help='specify a stop time for the time entry (defaults to now)',
    )

    resume_parser = subparsers.add_parser(
        'resume',
        description='Resume an entry from today.',
    )
    resume_parser.add_argument(
        '-s',
        '--start',
        default=datetime.now(),
        help='specify the start time for the resumed time entry '
        '(defaults to now)',
    )
    resume_parser.add_argument(
        'entry',
        type=int,
        nargs='?',
        help='the entry to resume (Python-style indexing, defaults to -1)',
    )

    list_parser = subparsers.add_parser(
        'list',
        description='List the time entries for a date.',
    )
    list_parser.add_argument(
        '-d',
        '--date',
        default=datetime.today().date(),
        help='the date to list time entries for (defaults to today)',
    )

    edit_parser = subparsers.add_parser(
        'edit',
        description='Edit time entries for a date.',
    )
    edit_parser.add_argument(
        '-d',
        '--date',
        default=datetime.today().date(),
        help='the date to edit time entries for (defaults to today)',
    )

    sync_parser = subparsers.add_parser(
        'sync',
        description='Synchronize time entries for a month.',
    )
    sync_parser.add_argument(
        'month',
        default=datetime.now().month,
        nargs='?',
        help=' '.join([
            'the month to synchronize time entries for (defaults to the',
            'current month, accepted formats: 01, 1, Jan, January, 2019-01)',
        ])
    )

    report_parser = subparsers.add_parser(
        'report',
        description='Output a report about time spent in a date range.',
    )
    report_parser.add_argument(
        '-m',
        '--month',
        help='shorthand for reporting over an entire month (accepted formats: '
        '01, 1, Jan, January, 2019-01)',
    )
    report_parser.add_argument(
        '-y',
        '--year',
        help='shorthand for reporting over an entire year',
    )
    report_parser.add_argument(
        '--lastweek',
        action='store_true',
        help='shorthand for reporting on last week (Sunday-Saturday)',
    )
    report_parser.add_argument(
        '--thisweek',
        action='store_true',
        help='shorthand for reporting on the current week (Sunday-today)',
    )
    report_parser.add_argument(
        '--thismonth',
        action='store_true',
        help='shorthand for reporting on the current month',
    )
    report_parser.add_argument(
        '--today',
        action='store_true',
        help='shorthand for reporting on today',
    )
    report_parser.add_argument(
        '--yesterday',
        action='store_true',
        help='shorthand for reporting on yesterday',
    )
    report_parser.add_argument(
        '--lastyear',
        action='store_true',
        help='shorthand for reporting on last year',
    )
    report_parser.add_argument(
        '--thisyear',
        action='store_true',
        help='shorthand for reporting on this year',
    )
    report_parser.add_argument(
        '-c',
        '--customer',
        help='customer ID to generate a report for',
    )
    report_parser.add_argument(
        '-p',
        '--project',
        help='project name to generate a report for',
    )
    report_parser.add_argument(
        'range_start',
        nargs='?',
        help='specify the start of the reporting range (defaults to the '
        'beginning of last month)',
    )
    report_parser.add_argument(
        'range_stop',
        nargs='?',
        help='specify the end of the reporting range (defaults to the end of '
        'last month)',
    )
    report_parser.add_argument(
        'filename',
        nargs='?',
        help='specify the filename to export the report to. '
        'If none specified, output to stdout',
    )

    if len(sys.argv) > 1:
        args = parser.parse_args()
    else:
        args = parser.parse_args(['list'])

    # Show version if requested.
    if args.version:
        cli.version()
        return

    {
        'start': cli.start,
        'stop': cli.stop,
        'resume': cli.resume,
        'list': cli.list_entries,
        'edit': cli.edit,
        'sync': cli.sync,
        'report': cli.report,
    }[args.action](args)
