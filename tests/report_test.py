from subprocess import check_output


def test_report():
    # TODO validate that the output makes sense
    check_output('tt report'.split()).decode()
