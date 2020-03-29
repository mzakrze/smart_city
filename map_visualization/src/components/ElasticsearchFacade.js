
class ElasticsearchFacade {

    /*
    TODO: 
    - narazie zamockowana nazwa 1 symulacji o nazwie "1"
    - docelowo: baza danych symulacji + mozliwosc wyboru
    */

    /*
    TODO:
    - narazie stworzone w ES 2 indexy
    - docelowo - stworzyć 2 index template
    */

    // TODO - wstawić w serviceWorker

    constructor() {
        const simulationNameMock = "";
        const INDEX_SEARCH_SUFFIX = "-search"
        this.index = "simulation-1-log";
        this.searchIndex = "simulation-1-map";

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
        // TODO - przetestować dla innej półkuli niż NE
        const topLeftLat = boundBox.north;
        const topLeftLon = boundBox.west;
        const bottomRightLat = boundBox.south;
        const bottomRightLon = boundBox.east;

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
                            { term: { startSecond: epochSecond }},
                            { range: { bbox_north : {gte: boundBox.south }}},
                            { range: { bbox_south : {lte: boundBox.north }}},
                            { range: { bbox_east : {gte: boundBox.west }}},
                            { range: { bbox_west : {lte: boundBox.east }}},
                        ]
                    }
                }})})
        .then(res => res.json())
        .then(res => {
            let hits = res.hits.hits;
            let result = {};
            for(let r of hits) {
                result[r._source.car_id] = r._source.location;
            } 
            if (res.hits.total > sizeParam) {
                console.error("Not every vehicle is showed!!! Total number of records exceeds " + sizeParam + " (max for Elasticsearch)")
            }
            doResolve(result);
            
            /*
            result = {
                0: [{lat, lon}, ...],
                1: [{lat, lon}, ...],
                ...,
                <car_id>: <location array>
            }
            */
        })
       })
        

    }

    getMinMaxTimestamp() {
        return new Promise(doResolve => {
            fetch("/" + this.index + "/_search", {
                method: 'POST',
                body: JSON.stringify({
                    "aggs" : {
                        "min_date": {"min": {"field": "@timestamp"}},
                        "max_date": {"max": {"field": "@timestamp"}}
                    }
                }),
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json'
                }})
            .then(res => res.json())
            .then(res => {
                doResolve({
                    minTimestamp: res.aggregations.min_date.value,
                    maxTimestamp: res.aggregations.max_date.value
                })
            })
        });
    }

}

export default ElasticsearchFacade;