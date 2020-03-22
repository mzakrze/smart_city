import React from 'react';
import './LeafletMap.css';
import './LeafletCanvasLayer.js';

const L = window.L;

// musi być takie same, jak w Go. FIXME - wynieść gdzieś albo czytać z Elastica
const SimulationStepInterval = 100;

class LeafletMap extends React.Component{

    constructor(props) {
        super(props);

        this.simulationResultCached = null;
    }

    componentDidMount() {
        var map = L.map('leaflet-map-id').setView([52.218994864793, 21.011712029573467], 14);

        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 18,
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        }).addTo(map);

        var that = this;

        var BigPointLayer = L.CanvasLayer.extend({
            renderCircle: function(ctx, point, radius) {
                ctx.fillStyle = 'rgba(255, 60, 60, 0.2)';
                ctx.strokeStyle = 'rgba(255, 60, 60, 0.9)';
                ctx.beginPath();
                ctx.arc(point.x, point.y, radius, 0, Math.PI * 2.0, true, true);
                ctx.closePath();
                ctx.fill();
                ctx.stroke();
            },

            renderVehicle: function(ctx, location) {
                ctx.fillStyle = 'rgba(0, 255, 60, 1)';
                ctx.fillRect(location.x, location.y, 10, 10);
            },

            render: function() {
                var canvas = this.getCanvas();
                var ctx = canvas.getContext('2d');

                // clear canvas
                ctx.clearRect(0, 0, canvas.width, canvas.height);

                // get center from the map (projected)
                var point = this._map.latLngToContainerPoint(new L.LatLng(52.218994864793, 21.011712029573467));

                var s = that.simulationResultCached;
                if (s != null) {
                    var v = s.vehicles[s.index];
                    var p2 = this._map.latLngToContainerPoint(new L.LatLng(v.location.lat, v.location.lon));
                    this.renderVehicle(ctx, p2)
                    s.index = (s.index + 1) % 10;

                }

                // render
                this.renderCircle(ctx, point, (1.0 + Math.sin(Date.now()*0.001))*100);

                setTimeout(() => this.redraw(), SimulationStepInterval);
            }
        });

        var layer = new BigPointLayer();
        layer.addTo(map);

    }

    UNSAFE_componentWillReceiveProps(nextProps) {
        if(nextProps.runSimulationRev != this.props.runSimulationRev) {
            this.fetchAndCache(nextProps.timestamp);
        }
    }

    fetchAndCache(timestamp) {
        fetch('/simulation/_search', {
            method: 'POST',
            body: JSON.stringify({
                "query": {
                    "bool": {
                        "filter": {
                            "range": {
                                "@timestamp": {"gte": 0, "lte": 9991584882223417 }
                            }
                        }
                    }
                },
                "sort": [
                    { "@timestamp": {order: "asc" }},
                    { "car_no": {order: "asc" }}
                ]    
            }),
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
        })
        .then(res => res.json())
        .then(res => {
            this.simulationResultCached = {
                ts_gte: null,
                vehicles: res.hits.hits.map(e => e._source),
                ts_lte: null,
                index: 0
            }
            console.log(this.simulationResultCached.vehicles);


        })
    }

    render() {
        return <div style={{display:'inline-block', marginRight:'20px'}}>
            <div id="leaflet-map-id"> </div>
        </div>;
    }
}

export default LeafletMap;
