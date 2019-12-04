"""
Requires you to install the Edge WebDriver by running:

DISM.exe /Online /Add-Capability /CapabilityName:Microsoft.WebDriver~~~~0.0.1.0
"""
import json
import os
import re

from selenium import webdriver
from typing import Dict, Optional

from tracktime.synchronisers.base import ExternalSynchroniser


class JiraSynchroniser(ExternalSynchroniser):
    """
    Uses Selenium with Microsoft Edge so that I don't have to do login. No login
    logic is required, since there is some sort of vudu magic with MS Active
    Directory happens such that it auto-logs me on in Edge.
    """
    types = ('jira', 'JIRA')

    def __init__(self, config):
        self.config = config
        self.root = self.config.get('jira', {}).get('root')
        if self.root and self.root[-1] == '/':
            self.root = self.root[:-1]
        self.driver = None

    def init_driver(self):
        self.driver = webdriver.Edge()
        self.driver.implicitly_wait(1)

    def __del__(self):
        if self.driver:
            self.driver.close()

    def get_name(self):
        return 'JIRA'

    def sync(self, aggregated_time, synced_time):
        return {}

    def get_formatted_task_id(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None

        return f'{entry.project}-{entry.taskid}'

    def get_task_link(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None
        return f'{self.root}/browse/{self.get_formatted_task_id(entry)}'

    def get_task_description(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None

        # This operation is expenive. Allow users to bypass.
        if os.environ.get('JIRA_DISABLE_TASK_DESCRIPTION_SCRAPE') == '1':
            return None

        cache_path = os.path.expanduser('~/.cache/tracktime/')
        os.makedirs(cache_path, exist_ok=True)
        cache_path += '/jira_selenium_ms_sso.json'

        # It's kinda inefficient to do this for every single task description,
        # but it's not as slow as Selenium, anyway.
        description_cache: Dict[str, str] = {}
        if os.path.exists(cache_path):
            with open(cache_path) as f:
                description_cache = json.load(f)

        formatted_task_id = self.get_formatted_task_id(entry)
        if not formatted_task_id:
            return None

        if not description_cache.get(formatted_task_id):
            if not self.driver:
                self.init_driver()

            self.driver.get(f'{self.root}/browse/{formatted_task_id}')
            description = self.driver.find_element_by_id(
                'summary-val').get_attribute('innerHTML')

            if not description:
                return None

            # The description resides in the first part of the #summary-val
            # component's HTML. The <span> is the edit button as far as I can tell.
            description_match = re.match(
                '(.*)<span class=".*"></span>',
                description,
            )
            if not description_match:
                return None

            description_cache[formatted_task_id] = description_match.group(1)

            with open(cache_path, 'w+') as f:
                json.dump(description_cache, f)

        return description_cache.get(formatted_task_id)
