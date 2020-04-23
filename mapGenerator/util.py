from geopy.distance import distance
from math import floor


# TODO - dziala tylko dla 1szej cwiartki
def bbox_to_meters(bbox):

    top_right = (bbox["north"], bbox["west"])
    top_left = (bbox["north"], bbox["east"])

    bottom_right = (bbox["south"], bbox["west"])

    width = floor(distance(top_right, top_left).meters)
    height = floor(distance(bottom_right, top_right).meters)

    print("width: " + str(width))
    print("height:" + str(height))

    return width, height
