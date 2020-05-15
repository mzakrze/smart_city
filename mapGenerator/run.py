import argparse
from enum import Enum
import util

from JunctionXGenerator2 import JunctionXGenerator2
from MapSerializer import MapSerializer

class Algorithm(Enum):
    manhattan = 'manhattan'
    not_implemented = 'not_implemented'

    def __str__(self):
        return self.value

    @staticmethod
    def from_string(s):
        try:
            return Algorithm[s]
        except KeyError:
            return ValueError()


parser = argparse.ArgumentParser(description='Generates intersection as a graph.')

generated_map = JunctionXGenerator2().generate(3)

MapSerializer().serialize(generated_map)


