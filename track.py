#! /usr/bin/env python3
import argparse
from datetime import datetime


def main():
    parser = argparse.ArgumentParser(description='Time tracker')
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
        default=datetime.today(),
        help='the date to list time entries for (defaults to today)')

    edit_parser = subparsers.add_parser('edit')
    edit_parser.add_argument(
        '-d',
        '--date',
        default=datetime.today(),
        help='the date to edit time entries for (defaults to today)')

    export_parser = subparsers.add_parser('export')
    export_parser.add_argument(
        '-m',
        '--month',
        help='specify the month to export (defaults to last month)')
    export_parser.add_argument(
        'filename', help='specify the filename to export the report to')

    args = parser.parse_args()
    print(args)
    {
        'start': None,
        'stop': None,
        'list': None,
        'edit': None,
        'export': None,
    }[args.action]


if __name__ == '__main__':
    main()
