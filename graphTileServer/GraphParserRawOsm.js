const fastXmlParser = require('fast-xml-parser');
const fs = require('fs');
const MAP_FILENAME = "weiti.osm"
const readline = require('readline'); // moduł core'owy, nie trzeba doinstalowywać

var lineReader = require('line-reader');



const acceptedHigwayTypes = [
    "motorway", 
    "trunk", 
    "primary", 
    "secondary", 
    "tertiary", 
    "unclassified", 
    "residential",
    "service", 
    "motorway_link", 
    "trunk_link", 
    "primary_link", 
    "secondary_link", 
    "tertiary_link", 
    "living_street"
];


class GraphParserRawOsm {

    getGraph() {

        var readFromCache = false;
        let nodes = {}
        let edges = {}
        const nodesData = fs.readFileSync('map_src/nodes.ndjson', 
            {encoding:'utf8', flag:'r'}); 

        const nodesDataLines = nodesData.split("\n")
        for (let line of nodesDataLines) {
            if (line == "") {
                continue;
            }
            let n = JSON.parse(line)
            nodes[n.id] = n
            readFromCache = true
        }

        const edgesData = fs.readFileSync('map_src/edges.ndjson', 
            {encoding:'utf8', flag:'r'}); 

        const edgesDataLines = edgesData.split("\n")
        for (let line of edgesDataLines) {
            if (line == "") {
                continue;
            }
            readFromCache = true
            let e = JSON.parse(line)
            edges[e.id] = e
        }

        if (readFromCache) {
            console.log("Done reading from cache, edges: ")
            return {
                nodes, edges
            }
        }

        console.log("Cannot read from cache, calculating ...")

        let res = this.doGetGraph()

        console.log("Calculation done. Writing to cache ...")

        var nodesStream = fs.createWriteStream('map_src/nodes.ndjson', {flags: 'w' })
        for(let nId in res.nodes) {
            nodesStream.write(JSON.stringify(res.nodes[nId]) + "\r\n")
        }
        nodesStream.end()

        var edgesStream = fs.createWriteStream('map_src/edges.ndjson', {flags: 'w' })
        for(let eId in res.edges) {
            edgesStream.write(JSON.stringify(res.edges[eId]) + "\r\n")
        }
        edgesStream.end()

        console.log("Done writing to cache")

        return res;
    }

    doGetGraph() {
        const originalGraph = this.readOsmToJsonBulk() // TODO - czytać strumien

        const bounds = {
            minlat: originalGraph.bounds.attr['@_minlat'],
            maxlat: originalGraph.bounds.attr['@_maxlat'],
            minlon: originalGraph.bounds.attr['@_minlon'],
            maxlon: originalGraph.bounds.attr['@_maxlon']
        }

        let nodes = {}
        for(let n of originalGraph.node) {
            let id = Number(n.attr['@_id'])
            let lat = Number(n.attr['@_lat'])
            let lon = Number(n.attr['@_lon'])

            let isIntersection = false; 
            let usedByCounter = 0;

            nodes[id] = { id, lat, lon, isIntersection, usedByCounter }
        }

        let edges = []
        for(let way of originalGraph.way) {
            let highwayType = ''
            let isHighway = false
            if (Array.isArray(way.tag)) {
                let r = way.tag.filter(e => e.attr['@_k'] == "highway")
                if (r.length > 0) {
                    isHighway = true
                    highwayType = r[0].attr['@_v'];
                }
            } else {
                if (way && way.tag && way.tag.attr && way.tag.attr['@_k'] == "highway") {
                    isHighway = true
                    highwayType = way.tag.attr['@_v'] 
                }
            }

            if (false == isHighway) {
                continue;
            } 

            if (acceptedHigwayTypes.indexOf(highwayType) == -1) {
                continue
            }

            let vertices = []
            for(let nd of way.nd) {
                let v = Number(nd.attr['@_ref'])
                nodes[v].usedByCounter += 1;
                vertices.push(v)
            }

            let bidirectional = true;
            let lanes = 1;

            if (Array.isArray(way.tag)) {
                let r = []
                r = way.tag.filter(e => e.attr['@_k'] == "lanes")
                if (r.length > 0) {
                    lanes = Number(r[0].attr['@_v'])
                }

                r = way.tag.filter(e => e.attr['@_k'] == "oneway")
                if (r.length > 0) {
                    bidirectional = r[0].attr['@_v'] != "yes"
                }
            } else {
                switch (way.tag.attr['@_k']) {
                    case "lanes":
                        lanes = Number(way.tag.attr['@_v'])
                        break;
                    case "oneway":
                        bidirectional = way.tag.attr['@_v'] != "yes"
                        break;
                }
            }

            let id = Number(way.attr['@_id'])
            edges[id] = {
                id,
                vertices,
                lanes,
                bidirectional
            }
        }

        let toDelete = []
        for(let nId in nodes) {
            if (nodes[nId].usedByCounter == 0) {
                toDelete.push(nId)
            } else if (nodes[nId].usedByCounter >= 3) {
                nodes[nId].isIntersection = true;
            }
        }
        for(let nId of toDelete) {
            delete nodes[toDelete];
        }


        for(let r of originalGraph.relation) {
            if (r.tag && r.tag.attr) {

            } else {
                continue;
            }

            if (r.tag.attr['@_v'].startsWith("public_transport:") || r.tag.attr['@_v'].startsWith("building") || r.tag.attr['@_v'].startsWith("associatedStreet")) {
                // ok
            } else {
                console.error("Unknown relation type: ", r)
            }
        }

        console.log("done.")
        return {
            nodes, 
            edges
        }
    }

    readOsmToJsonBulk() {
        var raw = fs.readFileSync('map_src/' + MAP_FILENAME, 'utf8')
        var res = fastXmlParser.parse(raw, {attrNodeName: 'attr', parseAttributeValue : true, ignoreAttributes : false,});
        return res.osm;
    }

}

module.exports = GraphParserRawOsm ;