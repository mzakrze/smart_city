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


parser = argparse.ArgumentParser(description='Generates city map as a graph.')

parser.add_argument('--type', required=True, type=Algorithm.from_string, choices=list(Algorithm), help='Algorithm used for generating city map')
parser.add_argument('--bbox_north', default="52.2254000", required=False, help='Bounding box for generated map')
parser.add_argument('--bbox_south', default="52.2221000", required=False, help='Bounding box for generated map')
parser.add_argument('--bbox_west', default="21.0220000", required=False, help='Bounding box for generated map')
parser.add_argument('--bbox_east', default="21.0353000", required=False, help='Bounding box for generated map')

args = parser.parse_args()

bbox = {
    "north": float(args.bbox_north),
    "south": float(args.bbox_south),
    "west": float(args.bbox_west),
    "east": float(args.bbox_east),
}

width, height = util.bbox_to_meters(bbox)
bbox["width"] = width
bbox["height"] = height

if args.type == Algorithm.manhattan:

    generated_map = ManhattanMapGenerator(width, height).generate()

    MapSerializer().serialize(bbox, generated_map)

else:
    print("ERROR: algorithm not implemented")
