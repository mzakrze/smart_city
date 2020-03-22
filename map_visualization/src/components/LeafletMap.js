import React from 'react';
import './LeafletMap.css';
import './LeafletCanvasLayer.js';

const L = window.L;

class LeafletMap extends React.Component{

    constructor(props) {
        super(props);
    }
    componentDidMount() {
        var map = L.map('leaflet-map-id').setView([52.218994864793, 21.011712029573467], 14);

        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 18,
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        }).addTo(map);

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

            render: function() {
                var canvas = this.getCanvas();
                var ctx = canvas.getContext('2d');

                // clear canvas
                ctx.clearRect(0, 0, canvas.width, canvas.height);

                // get center from the map (projected)
                var point = this._map.latLngToContainerPoint(new L.LatLng(52.218994864793, 21.011712029573467));

                // render
                this.renderCircle(ctx, point, (1.0 + Math.sin(Date.now()*0.001))*100);

                this.redraw();
            }
        });

        var layer = new BigPointLayer();
        layer.addTo(map);

    }

    render() {
        return <div style={{display:'inline-block', marginRight:'20px'}}>
            <div id="leaflet-map-id"> </div>
            </div>
    }
}

export default LeafletMap;
