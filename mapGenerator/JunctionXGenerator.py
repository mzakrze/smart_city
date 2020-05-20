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

    def generate(self, lanes):
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

        def add_edge(node_from, node_to):
            edges.append({
                "from": node_from["id"],
                "to": node_to["id"],
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
            STEPS = 30
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

        mArr = []

        for n in range(-lanes, lanes):
            if n < lanes:
                mArr.append(- n * Config.LANE_WIDTH - Config.LANE_WIDTH / 2)
            elif n == lanes:
                mArr.append(Config.LANE_WIDTH / 2)
            else:
                mArr.append(n * Config.LANE_WIDTH + Config.LANE_WIDTH / 2)
        mArr.sort()
        pointsId = [x for x in range(lanes - 1, 0, -1)] + [0] + [x for x in range(lanes)]

        i_x = self.intersection_center["x"]
        i_y = self.intersection_center["y"]
        i_n = {}
        counter = -1
        for lane in range(lanes * 2):
            counter += 1
            entrypoint = mArr[lane] < 0
            exitpoint = not entrypoint

            # up
            x = i_x + mArr[lane]
            y = i_y + Config.LANE_WIDTH * lanes * 2
            n_u_i = {"x": x, "y": y}
            n_u_e = {"x": x, "y": self.map_height, "wayId": 1}
            generate_id(n_u_i, n_u_e)
            nodes = nodes + [n_u_i, n_u_e]
            if exitpoint:
                n_u_e["isExitPoint"] = True
                n_u_e["exitPointId"] = pointsId[counter]
                add_edge(n_u_i, n_u_e)
            if entrypoint:
                n_u_e["isEntryPoint"] = True
                n_u_e["entryPointId"] = pointsId[counter]
                add_edge(n_u_e, n_u_i)
            i_n["u" + str(counter)] = n_u_i

            # down
            x = i_x - mArr[lane]
            y = i_y - Config.LANE_WIDTH * lanes * 2
            n_d_i = {"x": x, "y": y}
            n_d_e = {"x": x, "y": 0, "wayId": 3}
            generate_id(n_d_i, n_d_e)
            nodes = nodes + [n_d_i, n_d_e]
            if exitpoint:
                n_d_e["isExitPoint"] = True
                n_d_e["exitPointId"] = pointsId[counter]
                add_edge(n_d_i, n_d_e)
            if entrypoint:
                n_d_e["isEntryPoint"] = True
                n_d_e["entryPointId"] = pointsId[counter]
                add_edge(n_d_e, n_d_i)
            i_n["d" + str(counter)] = n_d_i

            # right
            x = i_x + Config.LANE_WIDTH * lanes * 2
            y = i_y - mArr[lane]
            n_r_i = {"x": x, "y": y}
            n_r_e = {"x": self.map_width, "y": y, "wayId": 2}
            generate_id(n_r_i, n_r_e)
            nodes = nodes + [n_r_i, n_r_e]
            if exitpoint:
                n_r_e["isExitPoint"] = True
                n_r_e["exitPointId"] = pointsId[counter]
                add_edge(n_r_i, n_r_e)
            if entrypoint:
                n_r_e["isEntryPoint"] = True
                n_r_e["entryPointId"] = pointsId[counter]
                add_edge(n_r_e, n_r_i)
            i_n["r" + str(counter)] = n_r_i

            # left
            x = i_x - Config.LANE_WIDTH * lanes * 2
            y = i_y + mArr[lane]
            n_l_i = {"x": x, "y": y}
            n_l_e = {"x": 0, "y": y, "wayId": 4}
            generate_id(n_l_i, n_l_e)
            nodes = nodes + [n_l_i, n_l_e]
            if exitpoint:
                n_l_e["isExitPoint"] = True
                n_l_e["exitPointId"] = pointsId[counter]
                add_edge(n_l_i, n_l_e)
            if entrypoint:
                n_l_e["isEntryPoint"] = True
                n_l_e["entryPointId"] = pointsId[counter]
                add_edge(n_l_e, n_l_i)
            i_n["l" + str(counter)] = n_l_i


        for lane in range(lanes):
            # u->l
            f = i_n["u" + str(lane)]
            t = i_n["l" + str(lanes * 2 - 1 - lane)]
            generate_arc(f, t, math.fabs(f["x"] - t["x"]), "right", 3)
            # u->r
            t = i_n["r" + str(lanes * 2 - 1 - lane)]
            generate_arc(f, t, math.fabs(f["x"] - t["x"]), "left", 4)
            # u->d
            t = i_n["d" + str(lanes * 2 - 1 - lane)]
            add_edge(f, t)
            # r->u
            f = i_n["r" + str(lane)]
            t = i_n["u" + str(lanes * 2 - 1 - lane)]
            generate_arc(f, t, math.fabs(f["x"] - t["x"]), "right", 4)
            # r->d
            t = i_n["d" + str(lanes * 2 - 1 - lane)]
            generate_arc(f, t, math.fabs(f["x"] - t["x"]), "left", 1)
            # r->l
            t = i_n["l" + str(lanes * 2 - 1 - lane)]
            add_edge(f, t)
            # d->l
            f = i_n["d" + str(lane)]
            t = i_n["l" + str(lanes * 2 - 1 - lane)]
            generate_arc(f, t, math.fabs(f["x"] - t["x"]), "left", 2)
            # d->r
            t = i_n["r" + str(lanes * 2 - 1 - lane)]
            generate_arc(f, t, math.fabs(f["x"] - t["x"]), "right", 1)
            # d->u
            t = i_n["u" + str(lanes * 2 - 1 - lane)]
            add_edge(f, t)
            # l->u
            f = i_n["l" + str(lane)]
            t = i_n["u" + str(lanes * 2 - 1 - lane)]
            generate_arc(f, t, math.fabs(f["x"] - t["x"]), "left", 3)
            # l->d
            t = i_n["d" + str(lanes * 2 - 1 - lane)]
            generate_arc(f, t, math.fabs(f["x"] - t["x"]), "right", 2)
            # l->r
            t = i_n["r" + str(lanes * 2 - 1 - lane)]
            add_edge(f, t)

        self.fix_nodes_ids(nodes, edges)

        conflict_zone = {
            "maxY": self.intersection_center["y"] + Config.LANE_WIDTH * lanes * 2,
            "minY": self.intersection_center["y"] - Config.LANE_WIDTH * lanes * 2,
            "minX": self.intersection_center["x"] - Config.LANE_WIDTH * lanes * 2,
            "maxX": self.intersection_center["x"] + Config.LANE_WIDTH * lanes * 2,
        }

        return {
            "nodes": nodes,
            "edges": edges,
            "mapWidth": self.map_width,
            "mapHeight": self.map_height,
            "conflictZone": conflict_zone,
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


