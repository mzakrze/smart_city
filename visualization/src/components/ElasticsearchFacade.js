
class ElasticsearchFacade {


    // TODO - wstawić w serviceWorker

    constructor() {
        this.searchIndex = "simulation-map";

    }

    async isSearchIndexReady() {
        const result = await this._isSearchIndexReady();
        console.log("result:" + result) 
        return result;
    }

    _isSearchIndexReady() {
        return new Promise(doResolve => {
            fetch("/_cat/indices?format=json&pretty")
                .then(res => res.json())
                .then(indices => {
                    let indexCount = null;
                    let searchIndexCount = null;

                    for(let i of indices) {
                        if (i.index == this.index) {
                            indexCount = Number(i["docs.count"]);
                        }
                        if (i.index == this.searchIndex) {
                            searchIndexCount = Number(i["docs.count"]);
                        }
                    }

                    // TODO - należy sprawdzać czy jest gotowy w bardziej dokładny sposób: zliczać ile jest wpisów (suma lenght array location)
                    // TODO - docelowo inicjowane z wybieraka symulacji
                    var isReady = searchIndexCount != null;
                    console.debug("Simulation index has " + indexCount + " documents. ");
                    console.debug("Does simulation search index exist? " + (isReady ? "yes" : "no"));
                    
                    doResolve(isReady); 
                })
        })
    }

    getResultsForSecondAndBBox(epochSecond, boundBox) {
        const sizeParam = 10000;
        /*
        Zakładam, że samochód o danym id pojawia się na początku danej sekundy (czyli w timestamp = "*000"), i nie pojawia się "w połowie" sekundu
        */
       return new Promise(doResolve => {
        fetch("/" + this.searchIndex + "/_search?size=" + sizeParam, {
            method: "POST",
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                query: {
                    bool: {
                        filter: [
                            { term: { second: epochSecond }},
                        ]
                    }
                }})})
        .then(res => res.json())
        .then(res => {
            let hits = res.hits.hits;
            let result = {};
            for(let r of hits) {
                result[r._source.vehicle_id] = r._source.location_array;
                // TODO - może lepiej robić to w ES - będzie wydajniej
                for (let index in result[r._source.vehicle_id]) {
                    result[r._source.vehicle_id][index].alpha = r._source.alpha_array[index]
                    result[r._source.vehicle_id][index].state = r._source.state_array[index]
                }
            }
            if (res.hits.total > sizeParam) {
                console.error("Not every vehicle is showed!!! Total number of records exceeds " + sizeParam + " (max for Elasticsearch)")
            }
            doResolve(result)
            
            /*
            result = {
                0: [{lat, lon}, ...],
                1: [{lat, lon}, ...],
                ...,
                <vehicle_id>: <location array>
            }
            */
        })
       })
        

    }


}

export default ElasticsearchFacade;