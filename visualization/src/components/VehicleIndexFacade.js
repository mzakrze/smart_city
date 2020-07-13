class VehicleIndexFacade {

    constructor() {

    }

    async getVehicleIdTurnsMap() {
        return fetch("/simulation-vehicle/_search?size=10000", {
            method: 'GET',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
        })
            .then(res => res.json())
    }
}

export default VehicleIndexFacade;