import React from 'react';
import './LeafletMap.css';
import './LeafletCanvasLayer.js';
import ElasticsearchFacade from './ElasticsearchFacade.js';
import { Settings } from './../App.js';

const L = window.L;
class LeafletMap extends React.Component{

    constructor(props) {
        super(props);

        this.simulationResultCachedPing = null;
        this.simulationResultCachedPong = null;
        this.simulationCurrentSecond = null;
        this.simulationStopSecond = null;
        this.simulationPingPongState = null;
        this.timeSpentWaiting_ms = null;
        this.cacheRefillingPing = false;
        this.cacheRefillingPong = false;
        this.simulationTimeOffset_ms = null;
        // TODO - moze bedzie trezba parametryzować pingPongDuration

        this.elasticsearchFacade = new ElasticsearchFacade();

        this.startSimulationTS = null;

        this.runningSimulationRev = null;
        this.startSimulationRev = null;

        this.map = null;
        this.simulationVisualizationLayer = null;
    }

    componentDidMount() {
        this.map = L.map('leaflet-map-id').setView([52.218994864793, 21.011712029573467], 14);

        // TODO - experiment with tile servers: https://wiki.openstreetmap.org/wiki/Tile_servers
        var mapTileLayer = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 18,
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        });
        mapTileLayer.addTo(this.map);

        var graphTileLayer = L.tileLayer("http://localhost:8080/{z}/{x}/{y}");
        graphTileLayer.addTo(this.map);

        var that = this;

        var SimulationVisualizationLayer = L.CanvasLayer.extend({
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
                ctx.fillStyle = 'rgba(255, 0, 60, 1)';
                ctx.fillRect(location.x, location.y, 5, 5);
            },

            render: function() {
                const redrawThrottle_ms = 50; 

                var canvas = this.getCanvas();
                var ctx = canvas.getContext('2d', { alpha: false });

                // clear canvas
                ctx.clearRect(0, 0, canvas.width, canvas.height);

                {
                    // debug drawing
                    var point = this._map.latLngToContainerPoint(new L.LatLng(52.218994864793, 21.011712029573467));
                    this.renderCircle(ctx, point, (1.0 + Math.sin(Date.now()*0.001))*100);


                    
                }

                if (that.runningSimulationRev != that.startSimulationRev) {
                    // TODO - narazie restart nie działa (tylko start)
                    that.runningSimulationRev = that.startSimulationRev;
                    that.startSimulationTS = new Date().getTime(); 
                }

                if (that.runningSimulationRev != null) {
                 
                    that.cacheNextPingPongIfNecessary() // TODO - to tak naprawde powinno byc inicjowane od razu przy przechodzeniu z ping na pong i odwrotnie, nie sprawdzane

                    var ms = new Date().getTime();

                    if (that.throttlingRedrawTs != null) {
                        // resuming after cache miss
                        that.timeSpentWaiting_ms += ms - that.throttlingRedrawTs;
                        that.throttlingRedrawTs = null;
                    }

                    let step = ms - that.timeSpentWaiting_ms - that.simulationTimeOffset_ms - that.simulationCurrentSecond * 1000;
                    if (step >= 1000) {
                        switch (that.simulationPingPongState) {
                        case "ping":
                            that.simulationPingPongState = "pong";
                            that.simulationResultCachedPing = null;
                            break;
                        case "pong":
                            that.simulationPingPongState = "ping";
                            that.simulationResultCachedPong = null;
                            break;
                        default:
                            throw new Error("Illegal state of ping pong, value:" + that.simulationPingPongState);
                        }

                        that.simulationCurrentSecond += 1;
                        step -= 1000;
                    }
                    
                    if (that.simulationStopSecond <= that.simulationCurrentSecond) {
                        that.runningSimulationRev = null;
                        that.startSimulationRev = null; // nie wiem czy ta flaga dobrze działa
                        alert("simulation finished");
                        this.redraw();
                        return;
                    }

                    let res = that.simulationPingPongState == "ping" ? that.simulationResultCachedPing : that.simulationResultCachedPong;

                    if (res == null) {
                        // renderowanie symbolu oczekiwania
                        // TODO - tymczasowy bieda-symbol: 
                        ctx.fillStyle = 'rgba(60, 60, 60, 0.5)';
                        ctx.fillRect(100, 100, 200, 200);

                        that.throttlingRedrawTs = new Date().getTime();;

                        // tutaj celowo nie jest wołany redraw, zostanie on zawołany jak będzie zwrotka z ES
                        return;
                    } 
                    
                    let index = Math.floor(step / Settings.SamplingPeriod_ms);

                    for(const [vehicle_id, location_array] of Object.entries(res)) {
                        let p = location_array[index];
                        // TODO - new L.LatLng można robić przed włożeniem do cache
                        let point = this._map.latLngToContainerPoint(new L.LatLng(p.lat, p.lon));

                        this.renderVehicle(ctx, point);
                    }

                }

                // TODO - w razie problemów wydajnosciowych - jesli sie da - w serviceworkerze pisac do canvasów i cacheować canvasy
                // TODO - zapisywać co ile czasu faktycznie bylo rysowanie
                // TODO - może lepiej zamiast renderowac co ktoras klatke - renderowac w spowolnionym tempie
                this.redraw();
            }
        });

        this.simulationVisualizationLayer = new SimulationVisualizationLayer();
        this.simulationVisualizationLayer
            .addTo(this.map);

        const overlayers = {
            "vehicles": this.simulationVisualizationLayer
        };
        const baseLayers = {
            "Maps": mapTileLayer,
            "Graph": graphTileLayer,
        };
        L.control.layers(baseLayers, overlayers).addTo(this.map);
    }

    cacheNextPingPongIfNecessary() {
        let callback = null;

        if (this.simulationPingPongState == "ping" && this.simulationResultCachedPong == null && this.cacheRefillingPong == false) {
            this.cacheRefillingPong = true;

            callback = (value) => {
                this.simulationResultCachedPong = value;
                this.cacheRefillingPong = false;

                if (this.throttlingRedrawTs) {
                    this.simulationVisualizationLayer.redraw();
                }
            }
        }
        if (this.simulationPingPongState == "pong" && this.simulationResultCachedPing == null && this.cacheRefillingPing == false) {
            this.cacheRefillingPing = true;

            callback = (value) => {
                this.simulationResultCachedPing = value;
                this.cacheRefillingPing = false;
        
                if (this.throttlingRedrawTs) {
                    this.simulationVisualizationLayer.redraw();
                }
            }
        }

        // TODO - w zależności od rozmiaru bbox przerzucać sie na inny (mniej / bardziej dokładny algorytm)
        if (callback != null) {
            this.elasticsearchFacade.getResultsForSecondAndBBox(this.simulationCurrentSecond + 1, this.getCurrentBoundingBox())
                .then(callback);
        }
    }

    // TODO - rewrite to newer version of React
    UNSAFE_componentWillReceiveProps(nextProps) {
        if(nextProps.runSimulationRev != this.props.runSimulationRev) {
            // TODO - docelowo tutaj też przepychana nazwa symulacji

            this.simulationCurrentSecond = Math.floor(nextProps.timestamp / 1000);
            this.simulationStopSecond = Math.floor(nextProps.simulationStop_ms / 1000);

            this.elasticsearchFacade.getResultsForSecondAndBBox(this.simulationCurrentSecond, this.getCurrentBoundingBox())
                .then(value => {
                    this.simulationPingPongState = "ping";
                    this.simulationTimeOffset_ms = new Date().getTime() - nextProps.timestamp;
                    this.timeSpentWaiting_ms = 0;
                    this.simulationResultCachedPing = value;
        
                    this.startSimulationRev = this.startSimulationRev == null ? 1 : this.startSimulationRev + 1;
                });
        }

        // TODO - jeśli stop
    }

    getCurrentBoundingBox() {
        return {
            north: this.map.getBounds()._northEast.lat,
            south: this.map.getBounds()._southWest.lat,
            east: this.map.getBounds()._northEast.lng,
            west: this.map.getBounds()._southWest.lng
        };
    }

    render() {
        return <div style={{display:'inline-block', marginRight:'20px'}}>
            <div id="leaflet-map-id"> </div>
        </div>;
    }
}

export default LeafletMap;
