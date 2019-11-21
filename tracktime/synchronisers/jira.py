import requests
from typing import Optional

from tracktime.config import get_config
from tracktime.synchronisers.base import ExternalSynchroniser


class JiraSynchroniser(ExternalSynchroniser):
    types = ('jira', 'JIRA')
    def __init__(self):
        self.config = get_config()
        self.root = self.config.get('jira', {}).get('root')
        self.api_user = self.config.get('jira', {}).get('api_user')
        self.api_key = self.config.get('jira', {}).get('api_key')

    def get_name(self):
        return 'JIRA'

    def sync(self, aggregated_time, synced_time):
        return {}

    def get_formatted_task_id(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return

        return f'{entry.project}-{entry.taskid}'

    def get_task_link(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return
        return f'{self.root}/browse/{self.get_formatted_task_id(entry)}'

    def get_task_description(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return
        if not self.api_user or not self.api_key:
            return

        auth = requests.auth.HTTPBasicAuth(self.api_user, self.api_key)
        url = f"{self.root}/rest/agile/1.0/issue/{self.get_formatted_task_id(entry)}"
        try:
            response = requests.get(
                url,
                headers={'Accept': 'application/json'},
                auth=auth,
            )
            return response.json().get('title')
        except Exception as e:
            print(response.text)
            return