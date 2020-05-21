class GraphProvider {

    async getGraph() {
        var graphResult = {}

       await fetch('/simulation-info/_search')
            .then(res => res.json())
            .then(res => {
                let graphRaw = res.hits.hits[0]._source.graph_raw
                graphResult = JSON.parse(graphRaw)
            });

        let nodes = {}
        for (let n of graphResult.nodes) {
            nodes[n.id] = n
        }
        graphResult.nodes = nodes
        return graphResult

    }

}


export default GraphProvider;