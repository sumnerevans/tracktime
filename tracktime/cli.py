import os
from datetime import date, datetime
from pathlib import Path
from subprocess import call

import tracktime
from tabulate import tabulate
from tracktime.entry_list import EntryList, get_path
from tracktime.report import Report
from tracktime.synchronisers import Synchroniser
from tracktime.time_parser import parse_date, parse_month, parse_time


def start(args):
    start = parse_time(args.start)
    EntryList(start.date()).start(start, args.description, args.type,
                                  args.project, args.taskid, args.customer)


def stop(args):
    stop = parse_time(args.stop)
    try:
        EntryList(stop.date()).stop(stop)
    except Exception as e:
        print(e)


def resume(args):
    start = parse_time(args.start)
    entry = args.entry or -1
    EntryList(start.date()).resume(start, entry)


def list_entries(args):
    date = parse_date(args.date)
    entry_list = EntryList(date)
    print(f'Entries for {date:%Y-%m-%d}')
    print('=' * 22)
    print()
    print(tabulate([dict(x) for x in entry_list], headers='keys'))

    print()
    hours, minutes = entry_list.total
    print(f'Total: {hours}:{minutes:02}')


def edit(args):
    """Open an editor to edit the time entries."""
    date = parse_date(args.date)

    # Ensure the header exists.
    EntryList(date).save()

    # Edit the entries
    editor = os.environ['EDITOR'] or os.environ['VISUAL']
    call([editor, get_path(date, makedirs=True)])

    # Reload and sync the time entries
    EntryList(date).sync()


def sync(args):
    if args.year and not args.month:
        print('You must specify a month when year is specified.')
        return

    Synchroniser(int(args.year), parse_month(args.month)).sync()


def report(args):
    if args.year:
        if not args.month:
            print('You must specify a month when year is specified.')
            return
        start = date(int(args.year), parse_month(args.month), 1)
    else:
        now = datetime.today().date()
        if not args.month:
            # Default to previous month
            if now.month == 1:  # It's January, default to last December
                start = date(now.year - 1, 12, 1)
            else:
                start = date(now.year, now.month - 1, 1)
        else:
            start = date(now.year, parse_month(args.month), 1)

    report = Report(start, args.customer, args.project)
    if args.filename:
        path = Path(args.filename)
        if path.suffix == '.pdf':
            report.export_to_pdf(path)
        elif path.suffix == '.html':
            report.export_to_html(path)
        elif path.suffix == '.rst':
            report.export_to_rst(path)
        else:
            raise Exception(f'Cannot export to "{path.suffix}" file format.')
    else:
        report.export_to_stdout()


def version():
    print(f'tracktime version {tracktime.__version__}')
