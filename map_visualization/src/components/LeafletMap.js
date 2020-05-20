import React from 'react';
import './LeafletMap.css';
import './LeafletCanvasLayer.js';
import ElasticsearchFacade from './ElasticsearchFacade.js';
import { Settings } from './../App.js';
import TripIndexFacade from "./TripIndexFacade";
import LeafletRoadPlotLayer from "./LeafletRoadPlotLayer";

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

        this.elasticsearchFacade = new ElasticsearchFacade();

        this.runningSimulationRev = null;
        this.startSimulationRev = null;

        this.map = null;
        this.simulationVisualizationLayer = null;
        new TripIndexFacade().getVehicleIdToSizeMap()
            .then((value => {
                this.vehicleIdToSizeMap = value
            }))
    }

    componentDidMount() {
        this.map = L.map('leaflet-map-id').setView([52.219111, 21.011711], 19);

        // TODO - experiment with tile servers: https://wiki.openstreetmap.org/wiki/Tile_servers
        var mapTileLayer = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 21,
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
            opacity: 0.0 // FIXME - taki hack - dzięki temu tilelayer działa nam przesuwanie i zoomowanie, dajemy opacity 0 bo nie potrzebujemy tej mapy
        });
        mapTileLayer.addTo(this.map);

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

            renderVehicle: function(ctx, location, size, metersToPixelsX, metersToPixelsY, alpha, vId, state) {
                ctx.fillStyle = 'rgba(255, 0, 60, 1)';
                let w = size.width * metersToPixelsX;
                let l = size.length * metersToPixelsY;
                w = Math.max(w, 6);
                l = Math.max(l, 6);
                let locationCorner = {x: location.x - l/2, y: location.y - w/2};

                // https://developer.mozilla.org/en-US/docs/Web/API/Canvas_API/Tutorial/Transformations
                const stateToColor = {
                    1: 'rgba(255, 0, 0, 1)',
                    2: 'rgba(0, 0, 255, 1)',
                    3: 'rgba(0, 0, 255, 1)',
                    4: 'rgba(255, 0, 0, 1)',
                }
                ctx.fillStyle = stateToColor[state]
                ctx.save();
                ctx.translate(locationCorner.x + l/2, locationCorner.y + w/2);

                ctx.rotate(alpha);

                ctx.translate(-locationCorner.x - l/2, -locationCorner.y - w/2);
                ctx.fillRect(locationCorner.x, locationCorner.y, l, w);
                ctx.restore();
            },

            render: function() {
                var canvas = this.getCanvas();
                var ctx = canvas.getContext('2d', { alpha: false });

                // clear canvas
                ctx.clearRect(0, 0, canvas.width, canvas.height);

                if (that.runningSimulationRev != that.startSimulationRev) {
                    // TODO - narazie restart nie działa (tylko start)
                    that.runningSimulationRev = that.startSimulationRev;
                    that.startSimulationTS = new Date().getTime(); 
                }

                // TODO - cache it - once per zoom
                const widthPixels = this._map.getSize().x;
                const heightPixels = this._map.getSize().y;

                const widthMeters = this._map.containerPointToLatLng([0, 0]).distanceTo(this._map.containerPointToLatLng([widthPixels, 0]));
                const heightMeters = this._map.containerPointToLatLng([0, 0]).distanceTo(this._map.containerPointToLatLng([0, heightPixels]));

                const METERS_TO_PIXELS_X = widthPixels / widthMeters;
                const METERS_TO_PIXELS_Y = heightPixels / heightMeters;

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
                        let size = that.vehicleIdToSizeMap[vehicle_id];
                        this.renderVehicle(ctx, point, size, METERS_TO_PIXELS_X, METERS_TO_PIXELS_Y, p.alpha, vehicle_id, p.state);
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

        this.graphPlotLayer = new GraphPlotLayer();
        this.graphPlotLayer
            .addTo(this.map);

        // this.roadPlotLayer = new LeafletRoadPlotLayer();
        // this.roadPlotLayer
        //     .addTo(this.map);

        var scale = L.control.scale({
            metric: true,
            imperial: false
        }).addTo(this.map);


        // FIXME - do poprawy przed przeniesieniem na inne srodowisko
        let nodesPromise = fetch('http://localhost:3000/nodes.ndjson')
            .then(res => res.text())

        let edgesPromise = fetch('http://localhost:3000/edges.ndjson')
            .then(res => res.text())

        Promise.all([nodesPromise, edgesPromise])
            .then(values => {
                let nodes = {} 
                let edges = [] 
                for (let line of values[0].split("\n")) {
                    if (line == "") {
                        continue;
                    }
                    let n = JSON.parse(line)
                    nodes[n.id] = n
                }
                for (let line of values[1].split("\n")) {
                    if (line == "") {
                        continue;
                    }
                    let e = JSON.parse(line)
                    edges.push(e)
                }

                const graph = { nodes, edges }
                that.graphPlotLayer.initGraphPlotLayer(graph)
                // that.roadPlotLayer.initRoadPlotLayer(graph)
            })

        const overlayers = {
            "vehicles": this.simulationVisualizationLayer,
            "Graph": this.graphPlotLayer,
            // "Road": this.roadPlotLayer,
        };
        const baseLayers = { };
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



const GraphPlotLayer = L.CanvasLayer.extend({

    initGraphPlotLayer: function (theGraph) {
        this.theGraph = theGraph;
        this.initialized = true
    },

    render: function() {
        if (!this.initialized) {
            this.redraw();
            return;
        }

        var canvas = this.getCanvas();
        var ctx = canvas.getContext('2d', { alpha: false });

        // clear canvas
        ctx.clearRect(0, 0, canvas.width, canvas.height);

        let nodes = {}
        for(let nId in this.theGraph.nodes) {
            let n = this.theGraph.nodes[nId]
            let {x, y} = this._map.latLngToContainerPoint(new L.LatLng(n.lat, n.lon));
            nodes[nId] = {x, y}
        }

        const canvas_arrow = (context, fromx, fromy, tox, toy) => {
            var dx = tox - fromx;
            var dy = toy - fromy;
            context.moveTo(fromx, fromy);
            context.lineTo(tox, toy);
            // drawing arrow for debug:
            // var angle = Math.atan2(dy, dx);
            // var headlen = 25; // length of head in pixels
            // context.lineTo(tox - headlen * Math.cos(angle - Math.PI / 6), toy - headlen * Math.sin(angle - Math.PI / 6));
            // context.moveTo(tox, toy);
            // context.lineTo(tox - headlen * Math.cos(angle + Math.PI / 6), toy - headlen * Math.sin(angle + Math.PI / 6));
        };

        for(let e of this.theGraph.edges) {
            let color = "#000000";
            let nodeFrom = nodes[e.from];
            let nodeTo = nodes[e.to];

            ctx.beginPath();
            ctx.strokeStyle = color;
            canvas_arrow(ctx, nodeFrom.x, nodeFrom.y, nodeTo.x, nodeTo.y)
            ctx.stroke()

        }

    }
});



export default LeafletMap;
