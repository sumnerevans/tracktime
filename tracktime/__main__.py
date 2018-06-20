#! /usr/bin/env python3
import argparse
import os
from datetime import datetime

import tracktime
from tracktime.entry_list import EntryList
from tracktime.report import Report
from tracktime.time_entry import TimeEntry


def main():
    parser = argparse.ArgumentParser(description='Time tracker')
    parser.add_argument(
        '-d',
        '--directory',
        default=os.path.expanduser('~/.tracktime'),
        help='specify the directory to use for tracktime (defaults to ~/.tracktime)')

    subparsers = parser.add_subparsers(
        dest='action', help='specify an action to perform')
    subparsers.required = True

    start_parser = subparsers.add_parser('start')
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
        '-d',
        '--description',
        help='specify a description for the time entry')
    start_parser.add_argument(
        'task', help='specify the task being worked on', nargs='?')

    stop_parser = subparsers.add_parser('stop')
    stop_parser.add_argument(
        '-s',
        '--stop',
        default=datetime.now(),
        help='specify a stop time for the time entry (defaults to now)')

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

    report_parser = subparsers.add_parser('report')
    report_parser.add_argument(
        '-m',
        '--month',
        help='specify the month to report on (defaults to previous month)')
    report_parser.add_argument(
        '-y',
        '--year',
        help='specify the year to report on (defaults to the year of the previous month)')
    report_parser.add_argument(
        'filename',
        nargs='?',
        help='specify the filename to export the report to')

    args = parser.parse_args()

    tracktime.root_directory = args.directory
    {
        'start': TimeEntry.start,
        'stop': TimeEntry.stop,
        'list': EntryList.list,
        'edit': EntryList.edit,
        'report': Report.create,
    }[args.action](**vars(args))


if __name__ == '__main__':
    main()
