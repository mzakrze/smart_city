import json
import math

NODES_FILE = "../map_visualization/public/nodes.ndjson"
EDGES_FILE = "../map_visualization/public/edges.ndjson"
META_FILE = "../map_visualization/public/graph.json"

GRAPH_FILE = "../map_visualization/public/intersection_graph.json"

class MapSerializer:

    def serialize(self, graph):

        min_lat = 52.219111
        min_lon = 21.011711

        earth = 6378.137  # radius of the earth in kilometer
        m = (1.0 / ((2.0 * math.pi / 360.0) * earth)) / 1000

        max_lat = min_lat + (graph["mapHeight"] * m)
        max_lon = min_lon + (graph["mapWidth"] * m) / math.cos(min_lat * (math.pi / 180.0))

        bbox = {"north": max_lat, "south": min_lat, "west": min_lon, "east": max_lon,
                "width": graph["mapWidth"], "height": graph["mapHeight"]}

        meta = {
            "graph": bbox,
            "im": graph["im"],
        }

        self.add_coords(bbox, graph)

        with open(META_FILE, "w") as f:
            json.dump(meta, f)

        with open(NODES_FILE, "w") as f:
            for n in graph["nodes"]:
                json.dump(n, f)
                f.write("\n")

        with open(EDGES_FILE, "w") as f:
            for e in graph["edges"]:
                json.dump(e, f)
                f.write("\n")

        with open(GRAPH_FILE, "w") as f:
            json.dump(graph, f, indent=4)

    def add_coords(self, bbox, graph):
        """
        Zakladamy, ze Ziemia jest plaska
        """

        # TODO - dziala tylko dla ćwiartki NE
        diff_lat = bbox["north"] - bbox["south"]; assert diff_lat > 0
        diff_lon = bbox["east"] - bbox["west"]; assert diff_lon > 0

        min_lat = bbox["south"]
        min_lon = bbox["west"]

        def meters_to_lat(meters):
            return min_lat + (meters / bbox["height"]) * diff_lat

        def meters_to_lon(meters):
            return min_lon + (meters / bbox["width"]) * diff_lon

        for n in graph["nodes"]:
            n["lat"] = meters_to_lat(n["y"])
            n["lon"] = meters_to_lon(n["x"])