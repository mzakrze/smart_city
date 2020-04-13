#/bin/bash




docker run --rm -it -v $PWD:/download openmaptiles/openmaptiles-tools \
       download-osm planet -- -d /download