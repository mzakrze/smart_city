import math


class Config:
    LANE_WIDTH = 3 # [meters]
    GRID_SIZE = 50 # co 100 metrów skrzyżowanie

# Układ odniesienia:
#
# (0,map_height),   ....,    (map_width, map_height)
#  ....,            ....,     ...
# (0,0)             ....,    (map_width, 0)


class ManhattanMapGenerator:

    def __init__(self, map_width, map_height):
        self.map_width = map_width
        self.map_height = map_height

        self.node_seq = 0

    def generate(self):
        unique_ids = {}
        def generate_id(*ns):
            for node in ns:
                node["id"] = (node["x"], node["y"])
                if node["id"] in unique_ids:
                    raise Exception("Not unique id !")
                unique_ids[node["id"]] = True

        def add_edge(node_from, node_to):
            edges.append({
                "from": node_from["id"],
                "to": node_to["id"],
            })

        def connect_up(node):
            is_uppermost_node = node["y"] == TOP_LANE_Y

            if is_uppermost_node:
                neighbour = (node["x"], self.map_height)
            else:
                neighbour = (node["x"], node["y"] + Config.GRID_SIZE - Config.LANE_WIDTH * 2)
            edges.append({
                "from": node["id"],
                "to": neighbour
            })

        def connect_down(node):
            is_downmost_interesction_node = node["y"] == BOTTOM_LANE_Y
            if is_downmost_interesction_node:
                neighbour = (node["x"], 0)
            else:
                neighbour = (node["x"], node["y"] - Config.GRID_SIZE + Config.LANE_WIDTH * 2)

            edges.append({
                "from": node["id"],
                "to": neighbour
            })

        def connect_left(node):
            is_leftmost_node = node["x"] == LEFT_LANE_X
            if is_leftmost_node:
                neighbour = (0, node["y"])
            else:
                neighbour = (node["x"] - Config.GRID_SIZE + Config.LANE_WIDTH * 2, node["y"])
            edges.append({
                "from": node["id"],
                "to": neighbour
            })

        def connect_right(node):
            is_rightmost_node = node["x"] == RIGHT_LANE_X
            if is_rightmost_node:
                neighbour = (self.map_width, node["y"])

            else:
                neighbour = (node["x"] + Config.GRID_SIZE - Config.LANE_WIDTH * 2, node["y"])
            edges.append({
                "from": node["id"],
                "to": neighbour
            })

        def generate_arc(fn1, fn2, node_from, node_to, radius, dir_x, dir_y, order):
            edges.append({
                "from": node_from["id"],
                "to": node_to["id"],
                "transitive": True
            })
            coef1 = { "up": -1, "down": 1}[dir_x]
            coef2 = { "left": -1, "right": 1}[dir_y]
            r_x = fn1(node_from["x"], node_to["x"])
            r_y = fn2(node_from["y"], node_to["y"])
            STEPS = 10
            prev = None
            for n in range(1, STEPS):
                alpha = n / STEPS * 3.1452 / 2
                x = r_x + coef1 * math.sin(alpha) * radius
                y = r_y + coef2 * math.cos(alpha) * radius
                d = { "x": x, "y": y }
                generate_id(d)
                nodes.append(d)
                if prev is None:
                    if order:
                        edges.append({
                            "from": node_from["id"],
                            "to": d["id"]
                        })
                    else:
                        edges.append({
                            "from": d["id"],
                            "to": node_to["id"]
                        })
                else:
                    if order:
                        edges.append({
                            "from": prev["id"],
                            "to": d["id"]
                        })
                    else:
                        edges.append({
                            "from": d["id"],
                            "to": prev["id"]
                        })
                prev = { "x": d["x"], "y": d["y"], "id": d["id"] }
            if order:
                edges.append({
                    "from": prev["id"],
                    "to": node_to["id"]
                })
            else:
                edges.append({
                    "from": node_from["id"],
                    "to": prev["id"]
                })

        nodes = []
        edges = []

        grids_x = self.map_width // Config.GRID_SIZE # floor division
        grids_y = self.map_height // Config.GRID_SIZE # floor division

        offset_x = (self.map_width - (grids_x * Config.GRID_SIZE)) / 2
        offset_y = (self.map_height - (grids_y * Config.GRID_SIZE)) / 2

        intersections = []
        for x in range(0, grids_x + 1):
            for y in range(0, grids_y + 1):
                intersections.append({
                    "x": offset_x + x * Config.GRID_SIZE,
                    "y": offset_y + y * Config.GRID_SIZE,
                })

        for y in range(0, self.map_height, Config.GRID_SIZE):
            n1 = {"x": 0, "y": y + offset_y - Config.LANE_WIDTH / 2}
            n2 = {"x": 0, "y": y + offset_y + Config.LANE_WIDTH / 2}
            n3 = {"x": self.map_width, "y": y + offset_y - Config.LANE_WIDTH / 2}
            n4 = {"x": self.map_width, "y": y + offset_y + Config.LANE_WIDTH / 2}
            generate_id(n1, n2, n3, n4)
            nodes = nodes + [n1, n2, n3, n4]

        for x in range(0, self.map_width, Config.GRID_SIZE):
            n1 = {"x": x + offset_x - Config.LANE_WIDTH / 2, "y": 0}
            n2 = {"x": x + offset_x + Config.LANE_WIDTH / 2, "y": 0}
            n3 = {"x": x + offset_x - Config.LANE_WIDTH / 2, "y": self.map_height}
            n4 = {"x": x + offset_x + Config.LANE_WIDTH / 2, "y": self.map_height}
            generate_id(n1, n2, n3, n4)
            nodes = nodes + [n1, n2, n3, n4]

        LEFT_LANE_X = offset_x - Config.LANE_WIDTH
        RIGHT_LANE_X = self.map_width - offset_x + Config.LANE_WIDTH
        TOP_LANE_Y = self.map_height - offset_y + Config.LANE_WIDTH
        BOTTOM_LANE_Y = offset_y - Config.LANE_WIDTH

        #      |   |
        #      n1  n2
        #
        # --n8         n3---
        #
        # --n7         n4---
        #
        #      n6  n5
        #       |   |
        for i in intersections:
            n_1 = {"x": i["x"] - Config.LANE_WIDTH / 2, "y": i["y"] + Config.LANE_WIDTH    }
            n_2 = {"x": i["x"] + Config.LANE_WIDTH / 2, "y": i["y"] + Config.LANE_WIDTH    }
            n_3 = {"x": i["x"] + Config.LANE_WIDTH    , "y": i["y"] + Config.LANE_WIDTH / 2}
            n_4 = {"x": i["x"] + Config.LANE_WIDTH    , "y": i["y"] - Config.LANE_WIDTH / 2}
            n_5 = {"x": i["x"] + Config.LANE_WIDTH / 2, "y": i["y"] - Config.LANE_WIDTH    }
            n_6 = {"x": i["x"] - Config.LANE_WIDTH / 2, "y": i["y"] - Config.LANE_WIDTH    }
            n_7 = {"x": i["x"] - Config.LANE_WIDTH    , "y": i["y"] - Config.LANE_WIDTH / 2}
            n_8 = {"x": i["x"] - Config.LANE_WIDTH    , "y": i["y"] + Config.LANE_WIDTH / 2}
            generate_id(n_1, n_2, n_3, n_4, n_5, n_6, n_7, n_8)
            nodes = nodes + [n_1, n_2, n_3, n_4, n_5, n_6, n_7, n_8]

            connect_up(n_2)
            connect_down(n_6)
            connect_left(n_8)
            connect_right(n_4)

            generate_arc(max, max, n_3, n_2, Config.LANE_WIDTH / 2,   "up", "left", True)
            generate_arc(max, min, n_3, n_6, Config.LANE_WIDTH * 1.5, "up",  "right", True) # TODO powinno byc down i left (ale dziala tak, ergo: funkcja jest źle zdefiniowana), poprawic tez nizej
            add_edge(n_3, n_8)
            generate_arc(max, min, n_5, n_4, Config.LANE_WIDTH / 2,   "up",  "right", False)
            generate_arc(min, min, n_5, n_8, Config.LANE_WIDTH * 1.5,  "down",  "right", False) # powinny byc up i left
            add_edge(n_5, n_2)
            generate_arc(min, min, n_7, n_6, Config.LANE_WIDTH / 2,   "down",  "right", True)
            generate_arc(min, max, n_7, n_2, Config.LANE_WIDTH * 1.5,  "down",  "left", True) # powinno byc up i right
            add_edge(n_7, n_4)
            generate_arc(min, max, n_1, n_8, Config.LANE_WIDTH / 2,   "down",  "left", False)
            generate_arc(max, max, n_1, n_4, Config.LANE_WIDTH * 1.5,  "up",  "left", False) # powinno byc down i right
            add_edge(n_1, n_6)

            if n_7["x"] == LEFT_LANE_X:
                edges.append({
                    "from": (0, n_7["y"]),
                    "to": n_7["id"]
                })
            if n_3["x"] == RIGHT_LANE_X:
                edges.append({
                    "from": (self.map_width, n_3["y"]),
                    "to": n_3["id"]
                })
            if n_1["y"] == TOP_LANE_Y:
                edges.append({
                    "from": (n_1["x"], self.map_height),
                    "to": n_1["id"]
                })
            if n_5["y"] == BOTTOM_LANE_Y:
                edges.append({
                    "from": (n_5["x"], 0),
                    "to": n_5["id"]
                })

        self.fix_nodes_ids(nodes, edges)

        return {
            "nodes": nodes,
            "edges": edges
        }

    def fix_nodes_ids(self, nodes, edges):
        seq = 0

        def next():
            nonlocal seq
            seq += 1
            return seq
        old_to_new = {}

        for n in nodes:
            id = next()
            old_to_new[n["id"]] = id
            n["id"] = id

        for e in edges:
            e["from"] = old_to_new[e["from"]]
            e["to"] = old_to_new[e["to"]]



