import time
from datetime import datetime


def parse_time_today(time_representation):
    if isinstance(time_representation, str):
        timestamp = time.strptime(time_representation, '%H:%M')
        now = datetime.now()
        dt = datetime(now.year, now.month, now.day, timestamp.tm_hour,
                      timestamp.tm_min)
        return dt

    return time_representation
