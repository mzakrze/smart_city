
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
        const simulationNameMock = ""; //"-1"
        const INDEX_SEARCH_SUFFIX = "-search"
        this.index = "simulation" + simulationNameMock;
        this.searchIndex = "simulation" + simulationNameMock + INDEX_SEARCH_SUFFIX;

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

    /**
     * returns Promise which is resolved after ETL process is completed
     */
    async initSearchIndexExtractTransormLoad() {
        // TODO - docelowo to wołane z "wybieraka" symulacji
        const isReady = await this._isSearchIndexReady();
        if (isReady) {
            return new Promise(resolve => {
                resolve({result: "success"});
            })
        }

        var {minTimestamp, maxTimestamp} = await this.getMinMaxTimestamp();

        var fetchAllPromises = []

        const interval_ms = 1000;
        for(let ts = minTimestamp; ts <= maxTimestamp; ts += interval_ms) {
            
            // FIXME - poniższe będzie działać tylko do 50 pojazdów i sampling = 50ms
            // https://discuss.elastic.co/t/how-to-elasticsearch-return-all-hits/133373
            const sizeParam = 10000;
            // FIXME - zakladamy narazie 1 samochod
            var promise = fetch("/" + this.index + "/_search?size=" + sizeParam, {
                method: "POST",
                body: JSON.stringify({
                    query: {
                        bool: {
                            filter: {
                                range: {
                                    "@timestamp": {gte: ts, lte: ts + interval_ms - 1}
                                }
                            }
                        }
                    },
                    sort: [
                        { "@timestamp": {order: "asc"}}
                    ]
                }),
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json'
                }
                })
            .then(res => res.json())
            .then(res => {
                if (res.hits.total > sizeParam) {
                    console.error("ERROR: Total hits number is greater than fetched");
                }

                // TODO - może nie działać dla półkuli innej niż NE
                let locationArray = [];
                let bbox_north = null; 
                let bbox_south = null; 
                let bbox_east = null; 
                let bbox_west = null; 
                for(let hit of res.hits.hits) {
                    let s = hit._source;
                    locationArray.push(s.location);
                    bbox_north = bbox_north == null || s.location.lon > bbox_north ? s.location.lon : bbox_north;
                    bbox_south = bbox_south == null || s.location.lon < bbox_south ? s.location.lon : bbox_south;
                    bbox_east = bbox_east == null || s.location.lon > bbox_east ? s.location.lat : bbox_east;
                    bbox_west = bbox_west == null || s.location.lon > bbox_west ? s.location.lat : bbox_west;
                }

                let document = {
                    car_id: res.hits.hits[0]._source.car_id, // FIXME - dostosować do więcej niż 1 pojazdu
                    startSecond: Math.floor(ts / 1000), // conversion: miliseconds -> seconds,
                    location: locationArray,
                    bbox_north,
                    bbox_south,
                    bbox_east,
                    bbox_west,
                };

                fetch("/" + this.searchIndex + "/_doc", {
                    method: "POST",
                    body: JSON.stringify(document),
                    headers: {
                        'Accept': 'application/json',
                        'Content-Type': 'application/json'
                    }})
                .then(res => res.json())
                .then(res => {
                    console.log(res);
                });
            })

            fetchAllPromises.push(fetchAllPromises);
        }

        return Promise.all(fetchAllPromises);
    }

    getResultsForSecondAndBBox(epochSecond, boundBox) {
        // TODO - przetestować dla innej półkuli niż NE
        // const topLeftLat = boundBox.bbox_north;
        // const topLeftLon = boundBox.bbox_west;
        // const bottomRightLat = boundBox.bbox_south;
        // const bottomRightLon = boundBox.bbox_east;

        /*
        Zakładam, że samochód o danym id pojawia się na początku danej sekundy (czyli w timestamp = "*000"), i nie pojawia się "w połowie" sekundu
        */
       return new Promise(doResolve => {
        fetch("/" + this.searchIndex + "/_search?", {
            method: "POST",
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                "query": {
                    term : {
                        startSecond: epochSecond
                    },
                    // FIXME - bounding box
                        // "filter" : {
                        //     "range" : {
                        //         "TODO" : {
                        //             "top_left" : {
                        //                 "lat" : topLeftLat,
                        //                 "lon" : topLeftLon
                        //             },
                        //             "bottom_right" : {
                        //                 "lat" : bottomRightLat,
                        //                 "lon" : bottomRightLon
                        //             }
                        //         }
                        //     }
                        // },







                }})})
        .then(res => res.json())
        .then(res => {
            // TODO - dodać obsługę wielu pojazdów
            doResolve(res.hits.hits[0]._source)
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