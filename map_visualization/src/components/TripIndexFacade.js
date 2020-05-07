class TripIndexFacade {

    constructor() {

    }

    async getVehicleIdToSizeMap() {

        let res = {}
        for (let i = 0; i < 100; i++) {
            res[i + ""] = {width: 2, length: 4}
        }

        return res

        let vehicleIdToSize = {}

        // TODO - wykrywać czy jest więcej niż limit
        await fetch("/simulation-trip-1/_search?size=10000", {
            method: 'GET',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
        })
            .then(res => res.json())
            .then(res => {
                for (let v of res.hits.hits) {
                    vehicleIdToSize[v._source.vehicle_id] = {
                        width: Number(v._source.vehicle_width),
                        length: Number(v._source.vehicle_length),
                    }
                }
            });

        return vehicleIdToSize;
    }
}

export default TripIndexFacade;