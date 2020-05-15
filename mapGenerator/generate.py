from JunctionXGenerator import JunctionXGenerator
from MapSerializer import MapSerializer


def generate_map_json(type, lanes):
    if type != "junction_x":
        raise Exception("Map generation algorithm " + type + " is not implemented.")

    generated_map = JunctionXGenerator().generate(lanes)

    generated_map_json = MapSerializer().serialize(generated_map)

    return generated_map_json

