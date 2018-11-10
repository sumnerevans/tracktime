from subprocess import check_output


def test_cli():
    # TODO validate that the output makes sense
    check_output('tt'.split()).decode()

    # '\n'.join([
    #         'Entries for ...',
    #         '======================',
    #         '...',
    #         'Total: ...',
    #     ]))

    check_output('tt --help'.split()).decode()
    check_output('tt sync'.split()).decode()
    check_output('tt sync -y 2018'.split()).decode()
    # should error
