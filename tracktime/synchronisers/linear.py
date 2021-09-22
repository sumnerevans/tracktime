"""Linear synchroniser module."""
from typing import Any, Dict, Optional, Tuple

from tracktime.time_entry import TimeEntry
from tracktime.synchronisers.base import AggregatedTime, ExternalSynchroniser


def get_path(obj: Dict[str, Any], *path: str) -> Any:
    v = obj
    for k in path:
        v = v.get(k, {})
    return v


class LinearSynchroniser(ExternalSynchroniser):
    types = ("linear",)

    def __init__(self, config):
        linear_config = config.get("linear", {})
        self.default_org = linear_config.get("default_org")

    def get_name(self):
        return "Linear"

    def sync(
        self,
        aggregated_time: AggregatedTime,
        synced_time: AggregatedTime,
        year_month: Tuple[int, int],
    ) -> AggregatedTime:
        return synced_time

    def get_formatted_task_id(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.project or not entry.taskid:
            return None
        return f"{entry.project}-{entry.taskid}"

    def get_task_link(self, entry: TimeEntry) -> Optional[str]:
        if entry.type not in self.types or not entry.project or not entry.taskid:
            return None

        issue_id = f"{entry.project}-{entry.taskid}"
        return f"https://linear.app/{self.default_org}/issue/{issue_id}"
