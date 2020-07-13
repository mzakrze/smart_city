#!/bin/bash

vpm_range=($(seq 10 10 200))
platoon_range=(off)

echo "vpm;platoon;delay;throughput" > result3.csv
echo "Date: "`date` > result3_verbose.txt
cat simulation_templ3.yml >> result3_verbose.txt
echo "" >> result3_verbose.txt
echo "-------" >> result3_verbose.txt

for platoon in ${platoon_range[@]}; do
    for vpm in ${vpm_range[@]}; do

        echo "Running platoon=${platoon} on ${vpm} vehicles per minute ..."

        cp simulation_templ3.yml tmp.yml
        sed "s/PLACEHOLDER_vehicles_per_minute/${vpm}/" -i tmp.yml
        sed "s/PLACEHOLDER_platooning/${platoon}/" -i tmp.yml

        out=$(./app -quiet -logsOff tmp.yml)

        delay=$(echo "${out}" | grep "Vehicle delay:")
        delay="${delay:15:100}"

        throughput=$(echo "${out}" | grep "Intersection throughput:")
        throughput="${throughput:25:100}"

        echo "$vpm;$platoon;$delay;$throughput" >> result3.csv

        echo "vpm=${vpm}, ip=${ip}" >> result3_verbose.txt
        echo "${out}" >> result3_verbose.txt
        echo "-----------" >> result3_verbose.txt

    done
done

rm tmp.yml

echo "Done."

