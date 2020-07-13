import math


class Config:
    INTERSECTION_APPROACH_LENGTH = 50  # [meters]
    LANE_WIDTH = 3  # [meters]

# Układ odniesienia:
#
# (0,map_height),   ....,    (map_width, map_height)
#  ....,            ....,     ...
# (0,0)             ....,    (map_width, 0)


class RoundaboutGenerator:
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

        def add_edge(node_from, node_to):
            edges.append({
                "from": node_from["id"],
                "to": node_to["id"],
            })

        def add_edge_arc(node_from, node_to):
            edges.append({
                "from": node_from["id"],
                "to": node_to["id"],
                "arc": True,
            })

        def generate_arc(node_from, node_to, dir):
            print(node_from)
            print(node_to)


            x_diff = (node_from["x"] - node_to["x"]) / 2
            y_diff = (node_from["y"] - node_to["y"]) / 2

            alpha = math.atan(y_diff / x_diff)
            middle = {"x": (node_from["x"] + node_to["x"]) / 2, "y": (node_from["y"] + node_to["y"]) / 2}
            coef = 0.5
            if dir == "up":
                middle["x"] += coef * math.sin(alpha)
                middle["y"] -= coef * math.cos(alpha)
            elif dir == "down":
                middle["x"] -= coef * math.sin(alpha)
                middle["y"] -= coef * math.cos(alpha)
            elif dir == "right":
                middle["x"] += coef * math.sin(alpha)
                middle["y"] -= coef * math.cos(alpha)
            elif dir == "left":
                middle["x"] -= coef * math.sin(alpha)
                middle["y"] += coef * math.cos(alpha)
            print("middle:", middle)
            print("alpha:", alpha)

            # middle["x"] += 1.2 * math.sin(alpha)
            # middle["y"] += 1 * math.cos(alpha)

            generate_id(middle)
            nodes.append(middle)
            add_edge_arc(node_from, middle)
            add_edge_arc(middle, node_to)

            # STEPS = 10
            # prev = node_from
            # for i in range(1, 5):
            #     x = middle["x"] + i / STEPS * 2 * math.sin(alpha)
            #     if i < STEPS / 2:
            #         y = middle["y"] - i / STEPS
            #     else:
            #         y = middle["y"] + i / STEPS
            #     n = {"x": x, "y": y}
            #     generate_id(n)
            #     nodes.append(n)
            #     add_edge(prev, n)
            #     prev = n
            # add_edge(n, node_to)




        nodes = []
        edges = []

        conflict_zone = {
            "maxY": self.intersection_center["y"] + Config.LANE_WIDTH * 3,
            "minY": self.intersection_center["y"] - Config.LANE_WIDTH * 3,
            "minX": self.intersection_center["x"] - Config.LANE_WIDTH * 3,
            "maxX": self.intersection_center["x"] + Config.LANE_WIDTH * 3,
        }

        mArr = []

        for n in range(-1, 1):
            if n < 1:
                mArr.append(- n * Config.LANE_WIDTH - Config.LANE_WIDTH / 2)
            elif n == 1:
                mArr.append(Config.LANE_WIDTH / 2)
            else:
                mArr.append(n * Config.LANE_WIDTH + Config.LANE_WIDTH / 2)
        mArr.sort()
        pointsId = [i for i in range(1 - 1, 0, -1)] + [0] + [i for i in range(1)]

        i_x = self.intersection_center["x"]
        i_y = self.intersection_center["y"]
        i_n = {}

        counter = -1
        exitpoints = []
        entrypoints = []
        for lane in range(2):
            counter += 1
            entrypoint = mArr[lane] < 0
            exitpoint = not entrypoint

            # up
            x = i_x + mArr[lane]
            y = i_y + Config.LANE_WIDTH * 2.8
            n_u_i = {"x": x, "y": y}
            n_u_e = {"x": x, "y": self.map_height, "wayId": 1}
            generate_id(n_u_i, n_u_e)
            nodes = nodes + [n_u_i, n_u_e]
            if exitpoint:
                n_u_e["isExitPoint"] = True
                n_u_e["exitPointId"] = pointsId[counter]
                add_edge(n_u_i, n_u_e)
                exitpoints.append(n_u_i)
            if entrypoint:
                n_u_e["isEntryPoint"] = True
                n_u_e["entryPointId"] = pointsId[counter]
                add_edge(n_u_e, n_u_i)
                entrypoints.append(n_u_i)

            i_n["u" + str(counter)] = n_u_i

            # down
            x = i_x - mArr[lane]
            y = i_y - Config.LANE_WIDTH * 1 * 2.8
            n_d_i = {"x": x, "y": y}
            n_d_e = {"x": x, "y": 0, "wayId": 3}
            generate_id(n_d_i, n_d_e)
            nodes = nodes + [n_d_i, n_d_e]
            if exitpoint:
                n_d_e["isExitPoint"] = True
                n_d_e["exitPointId"] = pointsId[counter]
                add_edge(n_d_i, n_d_e)
                exitpoints.append(n_d_i)
            if entrypoint:
                n_d_e["isEntryPoint"] = True
                n_d_e["entryPointId"] = pointsId[counter]
                add_edge(n_d_e, n_d_i)
                entrypoints.append(n_d_i)
            i_n["d" + str(counter)] = n_d_i

            # right
            x = i_x + Config.LANE_WIDTH * 1 * 2.8
            y = i_y - mArr[lane]
            n_r_i = {"x": x, "y": y}
            n_r_e = {"x": self.map_width, "y": y, "wayId": 2}
            generate_id(n_r_i, n_r_e)
            nodes = nodes + [n_r_i, n_r_e]
            if exitpoint:
                n_r_e["isExitPoint"] = True
                n_r_e["exitPointId"] = pointsId[counter]
                add_edge(n_r_i, n_r_e)
                exitpoints.append(n_r_i)
            if entrypoint:
                n_r_e["isEntryPoint"] = True
                n_r_e["entryPointId"] = pointsId[counter]
                add_edge(n_r_e, n_r_i)
                entrypoints.append(n_r_i)
            i_n["r" + str(counter)] = n_r_i

            # left
            x = i_x - Config.LANE_WIDTH * 1 * 2.8
            y = i_y + mArr[lane]
            n_l_i = {"x": x, "y": y}
            n_l_e = {"x": 0, "y": y, "wayId": 4}
            generate_id(n_l_i, n_l_e)
            nodes = nodes + [n_l_i, n_l_e]
            if exitpoint:
                n_l_e["isExitPoint"] = True
                n_l_e["exitPointId"] = pointsId[counter]
                add_edge(n_l_i, n_l_e)
                exitpoints.append(n_l_i)
            if entrypoint:
                n_l_e["isEntryPoint"] = True
                n_l_e["entryPointId"] = pointsId[counter]
                add_edge(n_l_e, n_l_i)
                entrypoints.append(n_l_i)
            i_n["l" + str(counter)] = n_l_i


        middle = {"x": i_x, "y": i_y}
        generate_id(middle)
        nodes = nodes + [middle]

        STEPS = 100
        prev = None
        first = None
        r_entries = [0, 25, 50, 75]
        for n in range(STEPS):
            alpha = n / STEPS * 3.1452 * 2
            x = middle["x"] + 1 * math.sin(alpha) * Config.LANE_WIDTH * 2
            y = middle["y"] + 1 * math.cos(alpha) * Config.LANE_WIDTH * 2
            d = { "x": x, "y": y }
            generate_id(d)
            nodes.append(d)

            if prev is not None:
                add_edge_arc(d, prev)
            else:
                first = d
            prev = d
            if n == STEPS - 1:
                add_edge_arc(first, d)

            if n == int(STEPS*11/12):
                generate_arc(entrypoints[0], d, "up")
            if n == int(STEPS*5/12):
                generate_arc(entrypoints[1], d, "down")
            if n == int(STEPS*2/12):
                generate_arc(entrypoints[2], d, "right")
            if n == int(STEPS*8/12):
                generate_arc(entrypoints[3], d, "left")

            if n == int(STEPS*1/12):
                generate_arc(d, exitpoints[0], "up")
            if n == int(STEPS*7/12):
                generate_arc(d, exitpoints[1], "down")
            if n == int(STEPS*4/12):
                generate_arc(d, exitpoints[2], "left")
            if n == int(STEPS*10/12):
                generate_arc(d, exitpoints[3], "right")


        self.fix_nodes_ids(nodes, edges)

        return {
            "nodes": nodes,
            "edges": edges,
            "mapWidth": self.map_width,
            "mapHeight": self.map_height,
            "conflictZone": conflict_zone,
            "type": "roundabout",
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