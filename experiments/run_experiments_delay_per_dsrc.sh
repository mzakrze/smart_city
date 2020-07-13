#!/bin/bash

dsrc_delay_range=(60 70 80)

echo "dsrc_delay;vehicle_delay;throughput" > result.csv
echo "Date: "`date` > result_verbose.txt
cat simulation_templ_per_dsrc.yml >> result_verbose.txt
echo "" >> result_verbose.txt
echo "-------" >> result_verbose.txt

for dsrc_delay in ${dsrc_delay_range[@]}; do

    echo "Running dsrc delay ${dsrc_delay} ..."

    cp simulation_templ_per_dsrc.yml tmp.yml
    sed "s/PLACEHOLDER_avg_latency/${dsrc_delay}/" -i tmp.yml

    out=$(./app -quiet -logsOff tmp.yml)

    delay=$(echo "${out}" | grep "Vehicle delay:")
    delay="${delay:15:100}"

    throughput=$(echo "${out}" | grep "Intersection throughput:")
    throughput="${throughput:25:100}"

    echo "$dsrc_delay;$delay;$throughput" >> result.csv
    echo "${out}" >> result_verbose.txt
    echo "-----------" >> result_verbose.txt

done

echo "Done."

