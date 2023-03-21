import datetime


def futuretime(seconds: int):
    """
    This function is used for the calculation of time to monitor for the MonitorUpdates RPC.
    It takes the user input in minutes, converts it to seconds.
    Then add the value to the current datetime and returns the value.
    Once, the current time hits that time, the client no longer receives updates about their reserved flights.
    """
    return (datetime.datetime.now() + datetime.timedelta(seconds=seconds)).timestamp()
