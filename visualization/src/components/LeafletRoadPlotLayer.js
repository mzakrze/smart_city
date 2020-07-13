const L = window.L;

const IMG_SIZE_METERS = 50;

const LeafletRoadPlotLayer = L.CanvasLayer.extend({

    initRoadPlotLayer: function (theGraph) {

        this.theGraph = theGraph;

        var that = this;
        new Promise((resolve, reject) => {
            let img = new Image();
            img.src = "grass.jpg";
            img.addEventListener('load', e => resolve(img));
            img.addEventListener('error', () => {
                reject(new Error(`Failed to load image's`));
            });
        })
            .then(img => {
                that.grass = img;
                that.initialized = true
            })
            .catch(err => {
                console.error(err)
            })
    },

    render: function() {
        if (!this.initialized) {
            this.redraw();
            return;
        }

        var canvas = this.getCanvas();
        var ctx = canvas.getContext('2d', { alpha: false });

        const widthPixels = this._map.getSize().x;
        const heightPixels = this._map.getSize().y;

        const widthMeters = this._map.containerPointToLatLng([0, 0]).distanceTo(this._map.containerPointToLatLng([widthPixels, 0]));
        const heightMeters = this._map.containerPointToLatLng([0, 0]).distanceTo(this._map.containerPointToLatLng([0, heightPixels]));

        const METERS_TO_PIXELS_X = widthPixels / widthMeters;
        const METERS_TO_PIXELS_Y = heightPixels / heightMeters;

        let w = IMG_SIZE_METERS * METERS_TO_PIXELS_X;
        let l = IMG_SIZE_METERS * METERS_TO_PIXELS_Y;

        let {x, y} = this._map.latLngToContainerPoint(new L.LatLng(52.219111, 21.011711)); // arbitrary **fixed** point

        for (let i = -10; i < 10; i++) {
            for (let j = -10; j < 10; j++) {
                ctx.drawImage(this.grass, x+w*i, y+l*j, w, l);
            }
        }

        const LANE_WIDTH = 3;
        let laneWithPixels = LANE_WIDTH * METERS_TO_PIXELS_X + 2;
        let next = {};
        let entrypoints = [];
        for (let e of this.theGraph.edges) {
            let nodeFrom = this.theGraph.nodes[e.from]
            let nodeTo = this.theGraph.nodes[e.to]

            let f = this._map.latLngToContainerPoint(new L.LatLng(nodeFrom.lat, nodeFrom.lon))
            let t = this._map.latLngToContainerPoint(new L.LatLng(nodeTo.lat, nodeTo.lon));

            if (e.arc) {
                if ((e.from in next) == false) {
                    next[e.from] = []
                }
                next[e.from].push(e.to);
            } else {
                entrypoints.push(e.to)
            }

            ctx.beginPath();
            ctx.moveTo(f.x, f.y);
            ctx.lineTo(t.x, t.y);
            ctx.strokeStyle = "#343434";
            ctx.lineWidth = laneWithPixels;
            ctx.stroke()

        }

        let guard = 100;
        let paintIt = (nId) => {
            guard -= 1;
            if (guard < 0) {
                return
            }
            let nodeFrom = this.theGraph.nodes[nId]
            for (let i = 0; i < 5; i++) {
                nId = next[nId]
                if (nId == undefined) {
                    return
                }
                nId = nId[0]
            }
            let nodeTo = this.theGraph.nodes[nId]

            let f = this._map.latLngToContainerPoint(new L.LatLng(nodeFrom.lat, nodeFrom.lon))
            let t = this._map.latLngToContainerPoint(new L.LatLng(nodeTo.lat, nodeTo.lon));

            ctx.beginPath();
            ctx.moveTo(f.x, f.y);
            ctx.lineTo(t.x, t.y);
            ctx.strokeStyle = "#343434";
            ctx.lineWidth = laneWithPixels;
            ctx.stroke()

            paintIt(nId)
        };

        for (let edge of entrypoints) {
            if (undefined != next[edge]) {
                guard = 1000;
                paintIt(next[edge][0]);
                guard = 1000;
                paintIt(next[edge][1]);
            }
        }
    }
});

export default LeafletRoadPlotLayer;