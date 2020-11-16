"""
Uses the Firefox Webdriver to get task descriptions from JIRA.
"""
import os
import pickle
import re
import time
from pathlib import Path
from typing import Dict, Optional, Tuple

from selenium import webdriver
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support.ui import WebDriverWait

from tracktime.synchronisers.base import AggregatedTime, ExternalSynchroniser


class JiraSynchroniser(ExternalSynchroniser):
    """
    Uses Selenium with Firefox.
    """

    types = ("jira", "JIRA")

    def __init__(self, config):
        self.config = config
        self.root = self.config.get("jira", {}).get("root")
        self.username = self.config.get("jira", {}).get("sso_email")
        self.password = self.config.get("jira", {}).get("sso_password")
        if self.root and self.root[-1] == "/":
            self.root = self.root[:-1]
        self.driver = None

    def init_driver(self):
        # Create the driver.
        self.driver = webdriver.Firefox()
        wait = WebDriverWait(self.driver, 10)

        # Open the JIRA root and clock on the Login button.
        self.driver.get(f"{self.root}")
        self.driver.find_element_by_css_selector("a.login-link").click()

        # Wait for the login field.
        def find_login_field():
            return self.driver.find_element_by_css_selector(
                "input[type=email][name=loginfmt]"
            )

        def find_password_field():
            return self.driver.find_element_by_css_selector(
                "input[type=password][name=passwd]"
            )

        wait.until(lambda d: find_login_field().is_displayed())
        time.sleep(1)
        login_field = find_login_field()
        login_field.send_keys(self.username)
        login_field.send_keys(Keys.ENTER)

        wait.until(lambda d: find_password_field().is_displayed())
        time.sleep(1)
        password_field = find_password_field()
        password_field.send_keys(self.password)
        password_field.send_keys(Keys.ENTER)

        input("Press enter when you have followed all of the login prompts.")

    def __del__(self):
        if self.driver:
            self.driver.close()

    def get_name(self):
        return "JIRA"

    def sync(
        self,
        aggregated_time: AggregatedTime,
        synced_time: AggregatedTime,
        year_month: Tuple[int, int],
    ) -> AggregatedTime:
        return synced_time

    def get_formatted_task_id(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None

        return f"{entry.project}-{entry.taskid}"

    def get_task_link(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None
        return f"{self.root}/browse/{self.get_formatted_task_id(entry)}"

    def get_task_description(self, entry) -> Optional[str]:
        if entry.type not in self.types or not entry.taskid:
            return None

        # This operation is expenive. Allow users to bypass.
        if os.environ.get("JIRA_DISABLE_TASK_DESCRIPTION_SCRAPE") == "1":
            return None

        formatted_task_id = self.get_formatted_task_id(entry)
        if not formatted_task_id:
            return None

        cache_path = Path("~/.cache/tracktime").expanduser()
        cache_path.mkdir(parents=True, exist_ok=True)
        cache_file = cache_path.joinpath("jira_selenium_firefox.pickle")

        # It's kinda inefficient to do this for every single task description,
        # but it's not as slow as Selenium, anyway.
        description_cache: Dict[str, str] = {}
        if cache_file.exists():
            with open(cache_file, "rb") as f:
                try:
                    description_cache = pickle.load(f)
                except Exception:
                    pass

        if not description_cache.get(formatted_task_id):
            if not self.driver:
                self.init_driver()

            self.driver.get(f"{self.root}/browse/{formatted_task_id}")

            def find_summary_val():
                return self.driver.find_element_by_id("summary-val")

            try:
                wait = WebDriverWait(self.driver, 10)
                wait.until(lambda d: find_summary_val().is_displayed())
                time.sleep(0.5)
                description = find_summary_val().get_attribute("innerHTML")
            except Exception:
                return None

            if not description:
                return None

            # The description resides in the first part of the #summary-val
            # component's HTML. The <span> is the edit button as far as I can tell.
            description_match = re.match('(.*)<span class=".*"></span>', description)
            if not description_match:
                return None

            description_cache[formatted_task_id] = description_match.group(1)

            with open(cache_file, "wb+") as f:
                pickle.dump(description_cache, f)

        return description_cache.get(formatted_task_id)
