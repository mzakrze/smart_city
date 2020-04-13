const fs = require('fs')

const parseSection = (parsingFunction) => {
    var n = Number(advance());
    var result = []
    for (var i = 0; i < n; i++) {
        result.push(parsingFunction());
    }
    return result;
}

function parseGraph() {
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
    width = Number(advance()); // (maxX - minX) * scale
    height = Number(advance());
    offX = Number(advance()); // #0 - minX
    offY = Number(advance()); // #0 - minY
    scale = Number(advance()); 
    
    global.scale = scale
    global.height = height
    global.offX = offX
    global.offY = offY
    
    var roads = parseSection(parseRoad);
    var edges = parseSection(parseEdge);
    var vertices = parseSection(parseVertex);
    var name = advance();
    var artifacts = parseSection(parseArtifact);

    return {
        width, height, offX, offY, scale, roads, edges, vertices, name, artifacts
    }
}


function parseRoad() {
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

    var id = Number(advance());
    var dirId = Number(advance());
    var length = Number(advance());
    var name = advance();
    var road_type = advance();
    var osm_id = advance();
    var vert1 = Number(advance());
    var vert2 = Number(advance());

    var points = parseSection(parseCoordinate);
    var shops = parseSection(parseCoordinate);
    var houses = parseSection(parseCoordinate);

    return {
        id, dirId, length, name, road_type, osm_id, vert1, vert2, points, shops, houses
    }
} 

function parseCoordinate() {
    var x = Number(advance())
    var y = Number(advance())
    // def fix(pt: Coordinate) = Coordinate((pt.x + offX) * scale, height - ((pt.y + offY) * scale))
    var oldX = x / global.scale - global.offX;
    var oldY = (global.height -y) / global.scale - global.offY; 
    console.log(oldX)
    return {x: oldX, y: oldY}
}

function parseEdge() {
    var id = Number(advance())
    var roadId = Number(advance())
    var laneNum = Number(advance())
    
    // TODO - tam jeszcze coś jest z kolejnością line'ów, ale narazie zostawiam tak
    var lines = parseSection(parseLine)

    return {
        id, roadId, laneNum, lines
    }
}

function parseLine() {
    var a = parseCoordinate()
    var b = parseCoordinate()
    return {x1: a.x, y1: a.y, x2: b.x, y2: b.y}
}

function parseVertex() {
    var location = parseCoordinate()
    var id = Number(advance())
    var turns = parseSection(parseTurn)
    return {
        id, location, turns
    }
}


function parseTurn() {
    var id = Number(advance())
    var edgeFrom = Number(advance())
    var edgeTo = Number(advance())
    return { id, edgeFrom, edgeTo }
}

function parseArtifact() {
    var points = parseSection(parseCoordinate)
    return { points }
}

class Reader {
    constructor(lines) {
        this.lines = lines;
        this.pointer = 0;
    }
    advance() {
        var res = this.lines[this.pointer]
        this.pointer++;
        return res;
    }
}

const data = readData();

const reader = new Reader(data.split("\n"));
    
const advance = () => reader.advance();
const global = {
    // height, scale, offX, offY
}

function readData() {
    try {
        const data = fs.readFileSync('weiti.map', 'utf8')
        return data;
    } catch (err) {
        console.error(err)
        throw err
    }
}

var graph = parseGraph()

