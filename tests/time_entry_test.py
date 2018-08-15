from datetime import datetime

from tracktime import TimeEntry


def expect_exception(fn, exception_text):
    try:
        fn()
    except Exception as e:
        assert str(e) == exception_text
    else:
        raise Exception('No exception thrown.')


def test_stop():
    def stop_before_start():
        start = datetime(2018, 1, 1, 13, 11)
        t = TimeEntry(start, 'cool')

        t.stop = datetime(2018, 1, 1, 13, 8)

    expect_exception(stop_before_start,
                     'Cannot stop a time entry before it was started.')
