import argparse
from enum import Enum
import util

from ManhattanMapGenerator import ManhattanMapGenerator
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

generated_map = ManhattanMapGenerator().generate()

MapSerializer().serialize(generated_map)


