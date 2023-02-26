class Flight:
    def __init__(self):
        self.name = ""


def main():
    f = Flight()

    if type(f) == object:
        print("this is a class")


main()
