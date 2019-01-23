#!/bin/bash

#set -e

# Defaults
test_files_dir='create_endpoint_mappings_test'
mappings_file='mappings.json'
mappings_reference="$test_files_dir/mappings.json.expected"
mappings_diff='mappings_diff.txt'
output_file='output.txt'
output_reference="$test_files_dir/output.txt.expected"
output_diff='output_diff.txt'

# Cleanup
rm -f $mappings_file
rm -f $mappings_diff
rm -f $output_file
rm -f $output_diff

# Run create_endpoint.sh and store output in a '$output_file'. 'mappings.json' will be create in the current directory
# during the script's execution.
./create_endpoint_mappings.sh $test_files_dir > $output_file

# Compare create mappings.json file with reference
mappings_diff=$(diff -w $mappings_file $mappings_reference)

# Compare output file with output reference file
output_diff=$(diff -w $output_file $output_reference)

# Print test results
if [ -z "$mappings_diff" ]; then
    echo "SUCCESS: generated mappings file is as expected."
else
    echo "FAILURE: generated mappings file is not as expected. Diff can be found in 'mappings_diff.txt'"
    echo $mappings_diff > 'mappings_diff.txt'
fi
if [ -z "$output_diff" ]; then
    echo "SUCCESS: script output is as expected."
else
    echo "FAILURE: script output is not as expected. Diff can be found in 'output_diff.txt'"
    echo $mappings_diff > 'output_diff.txt'
fi
