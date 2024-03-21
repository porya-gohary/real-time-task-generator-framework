#!/bin/bash

output_dir="../output"
cores=4
fault=1
shopt -s globstar
## read all generated yaml file one by one and test them in output directory and its subdirectories
yaml_files=$(ls $output_dir/**/**.prec.yaml)

for yaml_file in $yaml_files
do
    echo "Testing $yaml_file"
    result=$(../test/nptest $yaml_file -m $cores -w -f $fault)
    IFS=',' read -ra result_array <<< "$result"
    # remove additional spaces
    result_array[1]=$(echo "${result_array[1]}" | tr -d '[:space:]')
    echo "SAG result: ${result_array[1]}"

    # if the result is 1, then the test succeeded
    if [ "${result_array[1]}" == "1" ]; then
        echo "Test succeeded"
    else
        echo "Test failed"
        # remove the failed yaml file
        rm $yaml_file
        # remove its corresponding .dot file
        dot_file=$(echo $yaml_file | sed 's/prec.yaml/dot/')
        rm $dot_file

        # remove the corresponding .yaml file
        yaml_file=$(echo $yaml_file | sed 's/prec.yaml/yaml/')
        rm $yaml_file

    fi
done