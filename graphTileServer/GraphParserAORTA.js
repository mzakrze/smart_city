
class GraphParserAorta {

    constructor(mapData) {
        var lines = mapData.split("\n")
        var line = 0;
        
        this.advance = () => {
            var res = lines[line]
            line++;
            return res
        }
    }

    getGraph() {
        return this.parseGraph();
    }

    parseSection(parsingFunction) {
        parsingFunction = parsingFunction.bind(this)
        var n = Number(this.advance());
        var result = []
        for (var i = 0; i < n; i++) {
            result.push(parsingFunction());
        }
        return result;
    }
    
    parseGraph() {
        /*
        ## Graph.serialize
        w.doubles(width, height, offX, offY, scale)
        w.int(roads.size)
        roads.foreach(r => r.serialize(w))
        w.int(edges.size)
        edges.foreach(e => e.serialize(w))
        w.int(vertices.size)
        vertices.foreach(v => v.serialize(w))
        w.string(name)
        w.int(artifacts.size)
        artifacts.foreach(a => a.serialize(w))
        */
        var width, height, offX, offY, scale;
        width = Number(this.advance()); // (maxX - minX) * scale
        height = Number(this.advance());
        offX = Number(this.advance()); // #0 - minX
        offY = Number(this.advance()); // #0 - minY
        scale = Number(this.advance()); 
        
        this.scale = scale
        this.height = height
        this.offX = offX
        this.offY = offY
        
        var roads = this.parseSection(this.parseRoad.bind(this));
        var edges = this.parseSection(this.parseEdge.bind(this));
        var vertices = this.parseSection(this.parseVertex.bind(this));
        var name = this.advance();
        var artifacts = this.parseSection(this.parseArtifact.bind(this));

        return {
            width, height, offX, offY, scale, roads, edges, vertices, name, artifacts
        }
    }


    parseRoad() {
        /*
        ## Road.serialize
        w.int(w.roads(id).int)
        w.int(dir.id)
        w.double(length)
        w.strings(name, road_type, osm_id)
        w.int(w.vertices(v1.id).int)
        w.int(w.vertices(v2.id).int)
        w.int(points.size)
        points.foreach(pt => pt.serialize(w))
        w.int(shops.size)
        shops.foreach(pt => pt.serialize(w))
        w.int(houses.size)
        houses.foreach(pt => pt.serialize(w))
        */

        var id = Number(this.advance());
        var dirId = Number(this.advance());
        var length = Number(this.advance());
        var name = this.advance();
        var road_type = this.advance();
        var osm_id = this.advance();
        var vert1 = Number(this.advance());
        var vert2 = Number(this.advance());

        var points = this.parseSection(this.parseCoordinate);
        var shops = this.parseSection(this.parseCoordinate);
        var houses = this.parseSection(this.parseCoordinate);

        return {
            id, dirId, length, name, road_type, osm_id, vert1, vert2, points, shops, houses
        }
    } 

    parseCoordinate() {
        var x = Number(this.advance())
        var y = Number(this.advance())
        // def fix(pt: Coordinate) = Coordinate((pt.x + offX) * scale, height - ((pt.y + offY) * scale))
        var oldX = x / this.scale - this.offX;
        var oldY = (this.height -y) / this.scale - this.offY; 
        return {x: oldX, y: oldY}
    }

    parseEdge() {
        var id = Number(this.advance())
        var roadId = Number(this.advance())
        var laneNum = Number(this.advance())
        
        // TODO - tam jeszcze coś jest z kolejnością line'ów, ale narazie zostawiam tak
        var lines = this.parseSection(this.parseLine)

        return {
            id, roadId, laneNum, lines
        }
    }

    parseLine() {
        var a = this.parseCoordinate()
        var b = this.parseCoordinate()
        return {x1: a.x, y1: a.y, x2: b.x, y2: b.y}
    }

    parseVertex() {
        var location = this.parseCoordinate()
        var id = Number(this.advance())
        var turns = this.parseSection(this.parseTurn)
        return {
            id, location, turns
        }
    }


    parseTurn() {
        var id = Number(this.advance())
        var edgeFrom = Number(this.advance())
        var edgeTo = Number(this.advance())
        return { id, edgeFrom, edgeTo }
    }

    parseArtifact() {
        var points = this.parseSection(this.parseCoordinate)
        return { points }
    }

}

module.exports = GraphParserAorta ;