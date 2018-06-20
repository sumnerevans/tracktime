class EntryList:
    def __init__(self, date):
        self.date = date
        self.entries = []

    def load(self, date):
        entry_list = EntryList(date)
        # TODO: load entries from the file
        return entry_list

    def append(self, entry):
        self.entries.append(entry)

    def save(self):
        print('save')
        pass

    @staticmethod
    def list(date, **kwargs):
        """Gives you a list of ``TimeEntry``s for the given date."""
        return EntryList(date).entries

    @staticmethod
    def edit(date, **kwargs):
        print(date, kwargs)
