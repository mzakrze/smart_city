

var http = require('http');
var fs = require('fs');
const { createCanvas } = require('canvas')
const { parseString } = require('xml2js');
var Decimal = require('decimal.js'); // TODO - to nie jest potrzebne, precyzja i tak była taka sama
var fastXmlParser = require('fast-xml-parser');
var GraphParserAorta = require('./GraphParserAORTA.js')
var GraphParserAndgem = require('./GraphParserAndgem.js')
var GraphParserRawOsm = require('./GraphParserRawOsm.js')
var GraphEnhancer = require('./GraphEnhancer.js')
const MAP_HEIGHT = 400
const MAP_WIDTH = 400

/*
FIXME - create package.json with fixed dependencies versions

*/

const draw_myRawOsmParse = (ctx, lowerLat, upperLat, lowerLon, upperLon) => {

    const determineWaysColor = (way) => {


        const is = (type) => way.tag && (
            Array.isArray(way.tag) ? 
                way.tag.filter(e => e.attr['@_k'] == "highway" && e.attr['@_v'] == type)[0] != null
                :
                way.tag.attr['@_k'] == "highway" && way.tag.attr['@_v'] == type);

        const typesToShow = [
            "residential",
            "service",
            "living_street",
            "secondary",
            "secondary_link",
            "tertiary",
            "tertiary_link",
        ];

        const typesNotToShow = [
            "footway",
            "corridor",
            "cycleway",
            "pedestrian",
            "playground",
            "parking",
            "fence",
            "apartments",
            "riverbank",
        ]

        const typesIdontKnowYetIfShow = [
            "platform",
            "track",
            "steps",
            "path",
            "elevator",
        ]

        for (let type of typesNotToShow) {
            if (is(type)) {
                return undefined;
            }
        }

        for (let type of typesToShow) {
            if (is(type)) {
                return "#ff0000";
            }
        }

        for (let type of typesIdontKnowYetIfShow) {
            if (is(type)) {
                return "#00ff00";
            }
        }

        if (Array.isArray(way.tag)) {
            let r = way.tag.filter(e => e.attr['@_k'] == "highway").map(e => e.attr['@_v']);
            if (r.length > 0){
                console.log(r)
            }
        } else {
            if (way.tag && way.tag.attr && way.tag.attr['@_k'] == "highway") {
                console.log(way.tag.attr['@_v'])
            }
        }
        return undefined

    }

    let newLatDiff = upperLat - lowerLat;
    let newLonDiff = upperLon - lowerLon;

    let lonToX = (lng) => (lng - lowerLon) * MAP_WIDTH / newLonDiff;
    let latToY = (lat) => MAP_HEIGHT - ((lat - lowerLat) * MAP_HEIGHT / newLatDiff);

    let nodes = {};

    for(let n of loadedOsmFileAsObject.node) {
        let x = lonToX(parseFloat(n.attr['@_lon']));
        let y = latToY(parseFloat(n.attr['@_lat']));

        nodes[n.attr['@_id']] = {x, y};
        ctx.fillRect(x, y, 2, 2);

        if(0 < x && x < MAP_WIDTH && 0 < y && y < MAP_HEIGHT) {
            // TODO: warning
        }
    }

    for(let w of loadedOsmFileAsObject.way) {
        let color = determineWaysColor(w);
        if(color == null) {
            // null <=> don't draw this object
            continue;
        }
        let prevX, prevY;
        let first = true;
        
        for(let n of w.nd) {
            let nodeTo = nodes[n.attr['@_ref']];
            if(nodeTo == undefined) {
                console.error('err, id: ', n)
                continue;
            }
            let x = nodeTo.x;
            let y = nodeTo.y;

            if(first == false) {
                ctx.beginPath();
                ctx.moveTo(prevX, prevY);
                ctx.lineTo(x, y);
                ctx.strokeStyle = color;
                ctx.stroke();
            }
            prevX = x;
            prevY = y;
            first = false;
        }
    }

}


const drawMap_aorta = (ctx, lowerLat, upperLat, lowerLon, upperLon) => {
    const graph = new GraphParserAorta(MAP_DATA).getGraph();

    let newLatDiff = upperLat - lowerLat;
    let newLonDiff = upperLon - lowerLon;

    let lonToX = (lng) => (lng - lowerLon) * MAP_WIDTH / newLonDiff;
    let latToY = (lat) => MAP_HEIGHT - ((lat - lowerLat) * MAP_HEIGHT / newLatDiff);

    var edgesMap = {}
    for (let v of graph.vertices) {
        let x = lonToX(v.location.x)
        let y = latToY(v.location.y)

        // console.log(x)

        ctx.fillRect(x, y, 2, 2);
    }
    ctx.strokeStyle = "#ff0000";
    for (let e of graph.edges) {

        edgesMap[e.id] = {}
        for (let el of e.lines) {
            ctx.beginPath();
            ctx.moveTo(lonToX(el.x1), latToY(el.y1));
            ctx.lineTo(lonToX(el.x2), latToY(el.y2));
            ctx.stroke();
        }
    }
    ctx.strokeStyle = "#00ff00";
    // for (let r of graph.roads) {
    //     for (let pIndex = 1; pIndex < r.points.length; pIndex++) {
    //         var p1 = r.points[pIndex - 1];
    //         var p2 = r.points[pIndex];
    //         ctx.beginPath();
    //         ctx.moveTo(lonToX(p1.x), latToY(p1.y));
    //         ctx.lineTo(lonToX(p1.x), latToY(p2.y));
    //         ctx.stroke();
    //     }
    // }    

    // FIXME

}


const drawMap_andgem = (ctx, lowerLat, upperLat, lowerLon, upperLon) => {

    const graphParsed = new GraphParserAndgem(MAP_DATA).getGraph();

    const edgeColor = {
        "secondary": "#070707",
        "secondary_link": "#ff0000",
    }
    const otherEdgeColor = "#000000"

    const graph = new GraphEnhancer(graphParsed).getEnhancedGraph()

    let newLatDiff = upperLat - lowerLat;
    let newLonDiff = upperLon - lowerLon;

    let lonToX = (lng) => (lng - lowerLon) * MAP_WIDTH / newLonDiff;
    let latToY = (lat) => MAP_HEIGHT - ((lat - lowerLat) * MAP_HEIGHT / newLatDiff);

    for (let vId in graph.nodes) {
        let v = graph.nodes[vId]
        let x = lonToX(v.lon)
        let y = latToY(v.lat)
        ctx.fillRect(x-2, y-2, 5, 5);
    }

    for (let e of graph.edges) {
        let s = graph.nodes[e.source];
        let d = graph.nodes[e.dest];
        let c = edgeColor[e.type]
        if (c == undefined) {
            c = otherEdgeColor;
        }

        ctx.beginPath();
        ctx.moveTo(lonToX(s.lon), latToY(s.lat));
        ctx.lineTo(lonToX(d.lon), latToY(d.lat));
        ctx.strokeStyle = c;
        ctx.stroke();

    }

}

const drawMap_debug = (ctx, lowerLat, upperLat, lowerLon, upperLon) => {

    var r = Math.floor((Math.random() * 256));
    var g = Math.floor((Math.random() * 256));
    var b = Math.floor((Math.random() * 256));
    var color = "rgb("+r+","+g+","+b+")";

    // draw box
    ctx.beginPath();
    ctx.moveTo(0, 00);
    ctx.lineTo(0, MAP_HEIGHT);
    ctx.lineTo(MAP_WIDTH, MAP_HEIGHT);
    ctx.lineTo(MAP_WIDTH, 0);
    ctx.closePath();
    ctx.lineWidth = 5;
    ctx.fillStyle = color;
    ctx.fill();

}


const drawMap_osmParser = (ctx, lowerLat, upperLat, lowerLon, upperLon) => {

    let newLatDiff = upperLat - lowerLat;
    let newLonDiff = upperLon - lowerLon;

    let lonToX = (lng) => (lng - lowerLon) * MAP_WIDTH / newLonDiff;
    let latToY = (lat) => MAP_HEIGHT - ((lat - lowerLat) * MAP_HEIGHT / newLatDiff);

    nodes = {}
    for(let nId in theGraph.nodes) {
        let n = theGraph.nodes[nId]
        let x = lonToX(n.lon);
        let y = latToY(n.lat);

        nodes[nId] = {x, y}

        if (theGraph.nodes[nId].isIntersection) {
            ctx.fillRect(x-5, y-5, 10, 10);
        }

        if(0 < x && x < MAP_WIDTH && 0 < y && y < MAP_HEIGHT) {
            // TODO: warning
        }
    }

    for(let eId in theGraph.edges) {
        let e = theGraph.edges[eId]
        let color = "#000000"
        if(color == null) {
            // null <=> don't draw this object
            continue;
        }
        let prevX, prevY;
        let first = true;
        
        for(let n of e.vertices) {
            let nodeTo = nodes[n];
            if(nodeTo == undefined) {
                console.error('err, id: ', n, e.vertices)
                continue;
            }
            let x = nodeTo.x;
            let y = nodeTo.y;

            if(first == false) {
                ctx.beginPath();
                ctx.moveTo(prevX, prevY);
                ctx.lineTo(x, y);
                ctx.strokeStyle = color;
                ctx.stroke();
            }
            prevX = x;
            prevY = y;
            first = false;
        }
    }

}



// rozne algorytmy oraz datasource'y
// const MAP_FILENAME = "weiti.pycgr"
const DRAW_MAP_FUNCTION = drawMap_osmParser

/*
Dostępne: 
- andgem: https://github.com/AndGem/OsmToRoadGraph
- aorta: 
*/

/********** \/ framework \/ ******** */

// inspired by: https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
Decimal.set({ precision: 20, rounding: 1 })

const tile2long = (x,z) => {

    var before = (x/Math.pow(2,z)*360-180);

    x = new Decimal(x)
    z = new Decimal(z)
    
    var after = Number(x.div(  Decimal.pow(2, z)  ).times('360').sub('180'))

    if (before - after != 0) {
        console.log('lon, diff = ' + (before - after))
    }

    return after;
}
const tile2lat = (y,z) => {
    var n=Math.PI-2*Math.PI*y/Math.pow(2,z);
    var before = (180.0/Math.PI*Math.atan(0.5*(Math.exp(n)-Math.exp(-n))));


    const pi = Decimal.acos(-1)
    var nn = pi.minus(
        pi.times(2).times(y).div(Decimal.pow(2, z))
    )

    var after = new Decimal('180').div(pi).times(Decimal.atan(
        Decimal.exp(nn).minus(Decimal.exp( nn.times(-1) )).times(0.5)
    ))

    // if (before - after != 0) {
    //     console.log('lat, diff = ' + (before - after))
    // }

    return after;
}

const drawAndGetCanvas = (lowerLat, upperLat, lowerLon, upperLon) => {

    var canvas = createCanvas(MAP_WIDTH, MAP_HEIGHT); // najpierw width - potwierdzone info
    var context = canvas.getContext('2d');
    context.clearRect(0, 0, MAP_WIDTH, MAP_HEIGHT);

    DRAW_MAP_FUNCTION(context, lowerLat, upperLat, lowerLon, upperLon)

    // var r = Math.floor((Math.random() * 256));
    // var g = Math.floor((Math.random() * 256));
    // var b = Math.floor((Math.random() * 256));
    // var color = "rgb("+r+","+g+","+b+")";

    // // draw box
    // context.beginPath();
    // context.moveTo(0, 00);
    // context.lineTo(0, 800);
    // context.lineTo(800, 800);
    // context.lineTo(800, 0);
    // context.closePath();
    // context.lineWidth = 5;
    // context.fillStyle = color;
    // context.fill();

    return canvas;
}

const handleSimpleCanvas = (req, res) => {

    const { url } = req;

    const s = url.split("/")
    var zoom = Number(s[1])
    var y = Number(s[2])
    var x = Number(s[3])

    // TODO - nie bedzie dzialac dla innych cwiartek
    const lowerLat = tile2lat(x + 1, zoom),
        upperLat = tile2lat(x, zoom),
        lowerLon = tile2long(y, zoom),
        upperLon = tile2long(y + 1, zoom);

    // console.log("lowerLat: " + lowerLat + ", upperLat = " + upperLat + ", is ok:" + (lowerLat < upperLat))
    // console.log("lowerLon: " + lowerLon + ", upperLon = " + upperLon + ", is ok:" + (lowerLon < upperLon))

    const canvas = drawAndGetCanvas(lowerLat, upperLat, lowerLon, upperLon)

    var out = fs.createWriteStream(__dirname + '/generated_maps/' + 'text.png')

    var stream = canvas.pngStream();

    stream.on('data', chunk => out.write(chunk) );

    stream.on('end', () => {
        // console.log('saved png');

        var s = fs.createReadStream(__dirname + '/generated_maps/' + 'text.png');

        s.on('open', function () {
            res.setHeader('Content-Type', "image/jpeg");
            s.pipe(res);
        });

    }); 
}

const handleRequest = (req, res) => {

    handleSimpleCanvas(req, res)

}

http.createServer(handleRequest).listen(8080);

// const MAP_DATA = (() => {
//     try {
//         return fs.readFileSync('map_src/' + MAP_FILENAME, 'utf8')
//     } catch (err) { throw(err); }
// })()

// var res = fastXmlParser.parse(MAP_DATA, {attrNodeName: 'attr', parseAttributeValue : true, ignoreAttributes : false,});
// const loadedOsmFileAsObject = res.osm;

const theGraph = new GraphParserRawOsm().getGraph()

//console.log(loadedOsmFileAsObject.node)
