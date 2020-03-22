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
        this.cacheRefilling = false;
        this.simulationTimeOffset_ms = null;
        // TODO - moze bedzie trezba parametryzować pingPongDuration

        this.elasticsearchFacade = new ElasticsearchFacade();

        this.startSimulationTS = null;

        this.runningSimulationRev = null;
        this.startSimulationRev = null;
    }

    componentDidMount() {
        var map = L.map('leaflet-map-id').setView([52.218994864793, 21.011712029573467], 14);

        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 18,
            attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        }).addTo(map);

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
                ctx.fillRect(location.x, location.y, 10, 10);
            },

            render: function() {
                var canvas = this.getCanvas();
                var ctx = canvas.getContext('2d');

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
                    
                    that.cacheNextPingPongIfNecessary()

                    let ms = new Date().getTime();

                    let step = ms - that.simulationTimeOffset_ms - that.simulationCurrentSecond * 1000
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
                        console.log(that.simulationPingPongState);
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

                    let index = Math.floor(step / Settings.SamplingPeriod_ms);

                    let p = res.location[index];
                    // TODO - new L.LatLng można robić przed włożeniem do cache
                    let point = this._map.latLngToContainerPoint(new L.LatLng(p.lat, p.lon));

                    this.renderVehicle(ctx, point);
                }

                this.redraw();
            }
        });

        new SimulationVisualizationLayer()
            .addTo(map);

        this.elasticsearchFacade.initSearchIndexExtractTransormLoad();
    }

    cacheNextPingPongIfNecessary() {
        if(this.cacheRefilling) {
            return;
        }

        let setValueCallback = null;
        if (this.simulationPingPongState == "ping" && this.simulationResultCachedPong == null) {
            setValueCallback = (value) => this.simulationResultCachedPong = value; 
        }
        if (this.simulationPingPongState == "pong" && this.simulationResultCachedPing == null) {
            setValueCallback = (value) => this.simulationResultCachedPing = value;
        }

        // TODO - dodać bbox
        // TODO - w zależności od rozmiaru bbox przerzucać sie na inny (mniej / bardziej dokładny algorytm)
        if (setValueCallback != null) {
            this.elasticsearchFacade.getResultsForSecondAndBBox(this.simulationCurrentSecond + 1)
                .then(value => {
                    setValueCallback(value);
                    this.cacheRefilling = false;
                })
        }
    }

    // TODO - rewrite to newer version of React
    UNSAFE_componentWillReceiveProps(nextProps) {
        if(nextProps.runSimulationRev != this.props.runSimulationRev) {
            // TODO - docelowo tutaj też przepychana nazwa symulacji

            this.simulationCurrentSecond = Math.floor(nextProps.timestamp / 1000);
            this.simulationStopSecond = Math.floor(nextProps.simulationStop_ms / 1000);

            this.elasticsearchFacade.getResultsForSecondAndBBox(this.simulationCurrentSecond)
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

    render() {
        return <div style={{display:'inline-block', marginRight:'20px'}}>
            <div id="leaflet-map-id"> </div>
        </div>;
    }
}

export default LeafletMap;
