
class GraphParserAndgem {

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


    parseGraph() {
        const expect = (expected, actual) => { if (false == actual.startsWith(expected)) throw 'Err, expected: ' + expected + ', got: ' + actual}
        
        const graphResult = {}

        expect("# Road Graph File v.0.4", this.advance())
        expect("# number of nodes", this.advance())
        expect("# number of edges", this.advance())
        expect("# node_properties", this.advance())
        expect("# ...", this.advance())
        expect("# edge_properties", this.advance())
        expect("# ...", this.advance())

        const nodesNo = Number(this.advance())
        const edgesNo = Number(this.advance())

        let nodes = {}
        for (let i = 0; i < nodesNo; i++) {
            let n = this.parseNode()
            nodes[n.id] = {
                lat: n.lat,
                lon: n.lon
            }
        }

        let edges = []
        for (let i = 0; i < edgesNo; i++) {
            edges.push(this.parseEdge())
        }

        return {
            nodes,
            edges
        }
    }

    parseNode() {
        let s = this.advance().split(" ")

        return {
            id: Number(s[0]),
            lat: Number(s[1]),
            lon: Number(s[2]),
        }
    }

    parseEdge() {
        let s = this.advance().split(" ")

        return {
            source: Number(s[0]),
            dest: Number(s[1]),
            // s[2] - length, we dont care
            type: s[3],
            // s[4] - speed limit, we dont care 
            bidirectional: s[0] == '1'
        }
    }
}
























module.exports = GraphParserAndgem ;