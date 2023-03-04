import datetime


def futuretime(seconds: int):
    return (datetime.datetime.now() + datetime.timedelta(seconds=seconds)).timestamp()
