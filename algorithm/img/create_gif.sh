#!/bin/bash

convert -delay 10 -loop 0 `ls | grep png | sort` the_gif.gif

rm *.png

