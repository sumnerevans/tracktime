import os
import sys
import calendar

from datetime import date, datetime, timedelta
from pathlib import Path
from subprocess import call

from tabulate import tabulate

import tracktime
from tracktime.config import get_config
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
    entry = args.entry if args.entry is not None else -1
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

    # Determine the editor. Grab it from the config, then look to the EDITOR or
    # VISUAL enviornment variables.
    editor = get_config().get(
        'editor',
        os.environ.get(
            'EDITOR',
            os.environ.get('VISUAL'),
        ),
    )

    # Default the editor to something sensible (well, notepad isn't really
    # sensible as it is total garbage, but at least almost always exists on
    # Windows).
    if not editor:
        if sys.platform in ('win32', 'cygwin'):
            editor = 'notepad'
        else:
            editor = 'vim'

    # Open the editor to edit the entries
    filename = str(get_path(date, makedirs=True))
    call([editor, filename])

    # Reload and sync the time entries
    EntryList(date).sync()


def sync(args):
    Synchroniser(parse_month(args.month)).sync()


def report(args):
    now = datetime.today().date()

    if args.range_start or args.range_stop:
        if args.range_start is None or args.range_stop is None:
            raise Exception('Must specify range start and stop.')
        # TODO this should allow for more than just date specifications
        start_date = parse_date(args.range_start)
        end_date = parse_date(args.range_stop)
    elif args.year or args.lastyear:
        start_date = date((int(args.year) if args.year else now.year - 1), 1, 1)
        end_date = date(start_date.year, 12, 31)
    elif args.today:
        start_date = now
        end_date = start_date
    elif args.yesterday:
        start_date = now - timedelta(days=1)
        end_date = now
    else:  # monthly

        # Default to last month. Need to do this calculation to correctly get
        # the previous month across years.
        last_day_of_last_month = (date(now.year, now.month, 1) -
                                  timedelta(days=1))
        start_date = date(
            last_day_of_last_month.year,
            last_day_of_last_month.month,
            1,
        )
        if args.month:
            start_date = parse_month(args.month)
        elif args.thismonth:
            start_date = date(now.year, now.month, 1)

        end_date = date(
            start_date.year,
            start_date.month,
            calendar.monthrange(start_date.year, start_date.month)[1],
        )

    report = Report(start_date, end_date, args.customer, args.project)
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
