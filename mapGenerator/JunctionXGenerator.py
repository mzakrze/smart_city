import math


class Config:
    INTERSECTION_APPROACH_LENGTH = 50  # [meters]
    LANE_WIDTH = 3  # [meters]

# Układ odniesienia:
#
# (0,map_height),   ....,    (map_width, map_height)
#  ....,            ....,     ...
# (0,0)             ....,    (map_width, 0)


class JunctionXGenerator:

    def __init__(self):
        size = Config.INTERSECTION_APPROACH_LENGTH * 2 + 2 * Config.LANE_WIDTH
        self.map_width = size
        self.map_height = size
        self.intersection_center = {"x": size / 2, "y": size / 2}

        self.node_seq = 0

    def generate(self):
        unique_ids = {}

        def generate_id(*ns):
            for node in ns:
                node["id"] = (node["x"], node["y"])
                if node["id"] in unique_ids:
                    raise Exception("Not unique id !")
                unique_ids[node["id"]] = True

        def add_edge_arc(node_from, node_to):
            edges.append({
                "from": node_from["id"],
                "to": node_to["id"],
                "arc": True,
            })

        def generate_arc(node_from, node_to, radius, turn_dir, quarter):
            """
            quarter:\n
            1|2\n
            4|3
            """
            if quarter == 1:
                if turn_dir == "right":
                    r_x = max(node_from["x"], node_to["x"])
                    r_y = min(node_from["y"], node_to["y"])
                    coef1 = -1
                    coef2 = 1
                    order = False
                if turn_dir == "left":
                    r_x = max(node_from["x"], node_to["x"])
                    r_y = min(node_from["y"], node_to["y"])
                    coef1 = -1
                    coef2 = 1
                    order = True
            elif quarter == 2:
                if turn_dir == "right":
                    r_x = min(node_from["x"], node_to["x"])
                    r_y = min(node_from["y"], node_to["y"])
                    coef1 = 1
                    coef2 = 1
                    order = True
                if turn_dir == "left":
                    r_x = min(node_from["x"], node_to["x"])
                    r_y = min(node_from["y"], node_to["y"])
                    coef1 = 1
                    coef2 = 1
                    order = False
            elif quarter == 3:
                if turn_dir == "right":
                    r_x = min(node_from["x"], node_to["x"])
                    r_y = max(node_from["y"], node_to["y"])
                    coef1 = 1
                    coef2 = -1
                    order = False
                if turn_dir == "left":
                    r_x = min(node_from["x"], node_to["x"])
                    r_y = max(node_from["y"], node_to["y"])
                    coef1 = 1
                    coef2 = -1
                    order = True
            elif quarter == 4:
                if turn_dir == "right":
                    r_x = max(node_from["x"], node_to["x"])
                    r_y = max(node_from["y"], node_to["y"])
                    coef1 = -1
                    coef2 = -1
                    order = True
                if turn_dir == "left":
                    r_x = max(node_from["x"], node_to["x"])
                    r_y = max(node_from["y"], node_to["y"])
                    coef1 = -1
                    coef2 = -1
                    order = False
            else:
                raise Exception("Illegal quarter argument")
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
                            "to": d["id"],
                            "arc": True,
                        })
                    else:
                        edges.append({
                            "from": d["id"],
                            "to": node_to["id"],
                            "arc": True,
                        })
                else:
                    if order:
                        edges.append({
                            "from": prev["id"],
                            "to": d["id"],
                            "arc": True,
                        })
                    else:
                        edges.append({
                            "from": d["id"],
                            "to": prev["id"],
                            "arc": True,
                        })
                prev = { "x": d["x"], "y": d["y"], "id": d["id"] }
            if order:
                edges.append({
                    "from": prev["id"],
                    "to": node_to["id"],
                    "arc": True,
                })
            else:
                edges.append({
                    "from": node_from["id"],
                    "to": prev["id"],
                    "arc": True,
                })

        nodes = []
        edges = []

        n1_a = {"x": 0, "y": self.intersection_center["y"] - Config.LANE_WIDTH / 2, "entrypoint": 1}
        n2_a = {"x": 0, "y": self.intersection_center["y"] + Config.LANE_WIDTH / 2, "exitpoint": 1}
        n3_a = {"x": self.map_width, "y": self.intersection_center["y"] - Config.LANE_WIDTH / 2, "exitpoint": 2}
        n4_a = {"x": self.map_width, "y": self.intersection_center["y"] + Config.LANE_WIDTH / 2, "entrypoint": 2}
        generate_id(n1_a, n2_a, n3_a, n4_a)
        nodes = nodes + [n1_a, n2_a, n3_a, n4_a]

        n1 = {"x": self.intersection_center["x"] - Config.LANE_WIDTH / 2, "y": 0, "exitpoint": 3}
        n2 = {"x": self.intersection_center["x"] + Config.LANE_WIDTH / 2, "y": 0, "entrypoint": 3}
        n3 = {"x": self.intersection_center["x"] - Config.LANE_WIDTH / 2, "y": self.map_height, "entrypoint": 4}
        n4 = {"x": self.intersection_center["x"] + Config.LANE_WIDTH / 2, "y": self.map_height, "exitpoint": 4}
        generate_id(n1, n2, n3, n4)
        nodes = nodes + [n1, n2, n3, n4]

        #      |   |
        #      n1  n2
        #
        # --n8         n3---
        #
        # --n7         n4---
        #
        #      n6  n5
        #       |   |
        n_1 = {"x": self.intersection_center["x"] - Config.LANE_WIDTH / 2, "y": self.intersection_center["y"] + Config.LANE_WIDTH    }
        n_2 = {"x": self.intersection_center["x"] + Config.LANE_WIDTH / 2, "y": self.intersection_center["y"] + Config.LANE_WIDTH    }
        n_3 = {"x": self.intersection_center["x"] + Config.LANE_WIDTH    , "y": self.intersection_center["y"] + Config.LANE_WIDTH / 2}
        n_4 = {"x": self.intersection_center["x"] + Config.LANE_WIDTH    , "y": self.intersection_center["y"] - Config.LANE_WIDTH / 2}
        n_5 = {"x": self.intersection_center["x"] + Config.LANE_WIDTH / 2, "y": self.intersection_center["y"] - Config.LANE_WIDTH    }
        n_6 = {"x": self.intersection_center["x"] - Config.LANE_WIDTH / 2, "y": self.intersection_center["y"] - Config.LANE_WIDTH    }
        n_7 = {"x": self.intersection_center["x"] - Config.LANE_WIDTH    , "y": self.intersection_center["y"] - Config.LANE_WIDTH / 2}
        n_8 = {"x": self.intersection_center["x"] - Config.LANE_WIDTH    , "y": self.intersection_center["y"] + Config.LANE_WIDTH / 2}
        generate_id(n_1, n_2, n_3, n_4, n_5, n_6, n_7, n_8)
        nodes = nodes + [n_1, n_2, n_3, n_4, n_5, n_6, n_7, n_8]

        generate_arc(n_3, n_2, Config.LANE_WIDTH / 2, "right", 4)
        generate_arc(n_3, n_6, Config.LANE_WIDTH * 1.5, "left", 1)
        add_edge_arc(n_3, n_8)
        generate_arc(n_5, n_4, Config.LANE_WIDTH / 2, "right", 1)
        generate_arc(n_5, n_8, Config.LANE_WIDTH * 1.5, "left", 2)
        add_edge_arc(n_5, n_2)
        generate_arc(n_7, n_6, Config.LANE_WIDTH / 2, "right", 2)
        generate_arc(n_7, n_2, Config.LANE_WIDTH * 1.5, "left", 3)
        add_edge_arc(n_7, n_4)
        generate_arc(n_1, n_8, Config.LANE_WIDTH / 2, "right", 3)
        generate_arc(n_1, n_4, Config.LANE_WIDTH * 1.5, "left", 4)
        add_edge_arc(n_1, n_6)

        edges.append({
            "from": (0, n_7["y"]),
            "to": n_7["id"],
        })
        edges.append({
            "from": n_8["id"],
            "to": (0, n_8["y"]),
        })
        edges.append({
            "from": (self.map_width, n_3["y"]),
            "to": n_3["id"],
        })
        edges.append({
            "from": n_4["id"],
            "to":  (self.map_width, n_4["y"]),
        })
        edges.append({
            "from":(n_1["x"], self.map_height),
            "to":  n_1["id"],
        })
        edges.append({
            "from": n_2["id"],
            "to":  (n_2["x"], self.map_height),
        })
        edges.append({
            "from": (n_5["x"], 0),
            "to": n_5["id"],
        })
        edges.append({
            "from": n_6["id"],
            "to": (n_6["x"], 0),
        })

        self.fix_nodes_ids(nodes, edges)

# TODO - ruszyć stąd: obszar jurysdykcji IM
        # potem - komunikaty v<->im: REQ, GRANT, DENY (narazie 100% komunikatów i 0 latency)
        # potem - mock ip_sequential: tylko DENY, i potem pojazdy słucha się IM i jak nie ma granta to nie wjeżdzą
        # potem - napisać PID'a aby samochód się zatrzymywał przed wjazdem na skrzyżowanie
        # potem - IM taki, że wpuszcza 1 na raz
        # potem - IM FCFS
        intersection_manager = {
            "bboxUp": self.intersection_center["y"] + Config.LANE_WIDTH,
            "bboxDown": self.intersection_center["y"] - Config.LANE_WIDTH,
            "bboxLeft": self.intersection_center["x"] - Config.LANE_WIDTH,
            "bboxRight": self.intersection_center["x"] + Config.LANE_WIDTH,
        }

        return {
            "nodes": nodes,
            "edges": edges,
            "im": intersection_manager,
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


