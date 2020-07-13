from JunctionXGenerator import JunctionXGenerator
from RoundaboutGenerator import RoundaboutGenerator
from MapSerializer import MapSerializer


def generate_map_json(type, lanes):
    if type == "junction_x":
        generated_map = JunctionXGenerator().generate(lanes)
    elif type == "roundabout":
        if lanes > 1:
            print("Error: more than 1 lines for intersection type = roundabout is not supported")
            sys.exit(1)
        generated_map = RoundaboutGenerator().generate()
    else:
        raise Exception("Map generation algorithm " + type + " is not implemented.")

    generated_map_json = MapSerializer().serialize(generated_map)
    return generated_map_json
