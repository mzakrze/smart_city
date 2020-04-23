import json

NODES_FILE = "../map_visualization/public/nodes.ndjson"
EDGES_FILE = "../map_visualization/public/edges.ndjson"
META_FILE = "../map_visualization/public/graph.json"

class MapSerializer:

    def serialize(self, bbox, graph):

        self.add_coords(bbox, graph)

        with open(META_FILE, "w") as f:
            json.dump(bbox, f)

        with open(NODES_FILE, "w") as f:
            for n in graph["nodes"]:
                json.dump(n, f)
                f.write("\n")

        with open(EDGES_FILE, "w") as f:
            for e in graph["edges"]:
                json.dump(e, f)
                f.write("\n")

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