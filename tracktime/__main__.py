#! /usr/bin/env python3
import argparse
import sys
from datetime import datetime

from tracktime import cli


def main():
    parser = argparse.ArgumentParser(description='Time tracker')

    parser.add_argument(
        '-v', '--version', help='show version and exit', action='store_true')

    subparsers = parser.add_subparsers(
        dest='action', help='specify an action to perform')

    start_parser = subparsers.add_parser(
        'start', description='Start a time entry for today.')
    start_parser.add_argument(
        '-s',
        '--start',
        default=datetime.now(),
        help='specify a start time for the time entry (defaults to now)')
    start_parser.add_argument(
        '-t',
        '--type',
        help='specify the type of time entry to start',
        choices=['gitlab', 'github', 'gl', 'gh'])
    start_parser.add_argument(
        '-p', '--project', help='specify a project for the time entry')
    start_parser.add_argument(
        '-c', '--customer', help='specify a customer for the time entry')
    start_parser.add_argument(
        '-i', '--taskid', help='specify the task being worked on')
    start_parser.add_argument(
        'description',
        help='specify a description for the time entry',
        nargs='?')

    stop_parser = subparsers.add_parser('stop')
    stop_parser.add_argument(
        '-s',
        '--stop',
        default=datetime.now(),
        help='specify a stop time for the time entry (defaults to now)')

    resume_parser = subparsers.add_parser('resume')
    resume_parser.add_argument(
        '-s',
        '--start',
        default=datetime.now(),
        help='specify the start time for the resumed time entry '
        '(defaults to now)')
    resume_parser.add_argument(
        'entry',
        type=int,
        nargs='?',
        help='the entry to resume (Python-style indexing, defaults to -1)')

    list_parser = subparsers.add_parser('list')
    list_parser.add_argument(
        '-d',
        '--date',
        default=datetime.today().date(),
        help='the date to list time entries for (defaults to today)')

    edit_parser = subparsers.add_parser('edit')
    edit_parser.add_argument(
        '-d',
        '--date',
        default=datetime.today().date(),
        help='the date to edit time entries for (defaults to today)')

    sync_parser = subparsers.add_parser('sync')
    sync_parser.add_argument(
        '-y',
        '--year',
        default=datetime.now().year,
        help='the year to synchronize time entries for '
        '(defaults to the current month)')
    sync_parser.add_argument(
        '-m',
        '--month',
        default=datetime.now().month,
        help='the month to synchronize time entries for '
        '(defaults to the current month)')

    report_parser = subparsers.add_parser('report')
    report_parser.add_argument(
        '-m',
        '--month',
        help='specify the month to report on (defaults to previous month)')
    report_parser.add_argument(
        '-y',
        '--year',
        help='specify the year to report on '
        '(defaults to the year of the previous month)')
    report_parser.add_argument(
        '-c', '--customer', help='customer ID to generate a report for')
    report_parser.add_argument(
        '-p', '--project', help='project name to generate a report for')
    report_parser.add_argument(
        'filename',
        nargs='?',
        help='specify the filename to export the report to. '
        'If none specified, output to stdout')

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
