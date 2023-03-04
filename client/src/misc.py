import datetime

def futuretime():
    return (datetime.datetime.now() +datetime.timedelta(seconds=10000)).timestamp()