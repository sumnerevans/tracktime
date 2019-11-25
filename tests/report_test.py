import os
import tempfile
from datetime import date, timedelta
import subprocess

import pytest


@pytest.fixture()
def dummy_tracktime_dir():
    # Create a dummy tracktime directory.
    tmp_tracktime_dir = tempfile.TemporaryDirectory()
    return tmp_tracktime_dir


@pytest.fixture()
def dummy_config(dummy_tracktime_dir):
    # Create a dummy configuration file
    config_file = tempfile.NamedTemporaryFile('w+', delete=False)
    config_file.writelines(f'{x}\n' for x in [
        'fullname: Test User',
        'sync_time: true',
        f'directory: {dummy_tracktime_dir.name}',
        'customer_aliases:',
        '  SC: Some Company',
        'customer_addresses:',
        '  SC: |',
        '    1234 Some Street',
        '    Some City, CA 12345',
        '    USA',
    ])
    config_file.close()
    return config_file


def test_report_date_display(dummy_config):
    today = date.today()
    weekstart = today - timedelta(days=today.weekday())
    weekend = weekstart + timedelta(days=6)
    lastweekstart = weekstart - timedelta(days=7)
    lastweekend = weekend - timedelta(days=7)
    test_cases = [
        (('--thisyear', ), f'Time Report: {today.year}'),
        (('--lastyear', ), f'Time Report: {today.year - 1}'),
        (('--thismonth', ), f'Time Report: {today:%B %Y}'),
        (('--thisweek', ), f'Time Report: {weekstart} - {weekend}'),
        (('--lastweek', ), f'Time Report: {lastweekstart} - {lastweekend}'),
        (('-y', '2019', '-m', '01'), 'Time Report: January 2019'),
        (('-m', '11'), f'Time Report: November {today.year}'),
        (
            ('2019-01-02', '2019-02-09'),
            f'Time Report: 2019-01-02 - 2019-02-09',
        ),
        (
            ('--today', ),
            f'Time Report: {today.year}-{today.month}-{today.day}',
        ),
    ]
    for date_args, expected in test_cases:
        output_lines = subprocess.check_output(
            ['tt', '--config', dummy_config.name, 'report',
             *date_args]).decode().split(os.linesep)

        assert output_lines[0] == expected


def test_invalid_date_specifications(dummy_config):
    test_cases = [
        (('2019-01-01', ), 'Must specify range start and stop.'),
        (
            ('-y', '2019', '-m', '2018-02'),
            'When specifying a year with --year/-y, the month must be in the '
            'same year.',
        ),
    ]
    for date_args, expected in test_cases:
        result = subprocess.run(
            ['tt', '--config', dummy_config.name, 'report', *date_args],
            capture_output=True,
        )
        assert result.returncode != 0
        assert result.stderr.decode().endswith(
            f'Exception: {expected}{os.linesep}')


def test_report(dummy_config):
    output_lines = subprocess.check_output([
        'tt', '--config', dummy_config.name, 'report', '-y', '2019', '-m', '01'
    ]).decode().split(os.linesep)

    lines = [
        'Time Report: January 2019',
        '=========================',
        '',
        'User: Test User',
        '',
        'Grand Total: $0.00',
        '',
        'Detailed Time Report:',
        '',
        '                                                 Hours    Rate ($/h)     Total ($)',
        '----------------------------------------  ------------  ------------  ------------',
        'TOTAL                                             0.00                           0',
        '',
        '',
    ]

    for i, (expected, actual) in enumerate(zip(lines, output_lines)):
        print(i)
        assert expected == actual

    assert len(lines) == len(output_lines)


def test_report_with_customer(dummy_config):
    output_lines = subprocess.check_output([
        'tt', '--config', dummy_config.name, 'report', '-c', 'SC', '-y', '2019', '-m', '01'
    ]).decode().split(os.linesep)

    lines = [
        'Time Report: January 2019',
        '=========================',
        '',
        'User: Test User',
        '',
        'Customer:',
        '',
        '    Some Company',
        '    1234 Some Street',
        '    Some City, CA 12345',
        '    USA',
        '',
        'Grand Total: $0.00',
        '',
        'Detailed Time Report:',
        '',
        '                                                 Hours    Rate ($/h)     Total ($)',
        '----------------------------------------  ------------  ------------  ------------',
        'TOTAL                                             0.00                           0',
        '',
        '',
    ]

    for i, (expected, actual) in enumerate(zip(lines, output_lines)):
        print(i)
        assert expected == actual

    assert len(lines) == len(output_lines)
