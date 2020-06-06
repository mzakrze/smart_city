#!/bin/bash

ip_range=(trafficLights)
vrp_range=($(seq 10 10 200))

echo "vrp;policy;delay;throughput" > result.csv
echo "Date: "`date` > result_verbose.txt
cat simulation_templ.yml >> result_verbose.txt
echo "" >> result_verbose.txt
echo "-------" >> result_verbose.txt

for ip in ${ip_range[@]}; do
    for vrp in ${vrp_range[@]}; do

        echo "Running policy ${ip} on ${vrp} vehicles per minute ..."

        cp simulation_templ.yml tmp.yml
        sed "s/PLACEHOLDER_vehicles_per_minute/${vrp}/" -i tmp.yml
        sed "s/PLACEHOLDER_intersection_policy/${ip}/" -i tmp.yml

        out=$(./app -quiet -logsOff tmp.yml)

        delay=$(echo "${out}" | grep "Vehicle delay:")
        delay="${delay:15:100}"

        throughput=$(echo "${out}" | grep "Intersection throughput:")
        throughput="${throughput:25:100}"

        echo "$vrp;$ip;$delay;$throughput" >> result.csv

        echo "vrp=${vrp}, ip=${ip}" >> result_verbose.txt
        echo "${out}" >> result_verbose.txt
        echo "-----------" >> result_verbose.txt

    done
done

echo "Done."

