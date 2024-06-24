#!/bin/bash

# Define the path to your log file
LOG_FILE="./provisioning/ESresp.json"

# Define the Elasticsearch index URL
ES_URL="http://localhost:9200/log/_doc"

# Function to post a log entry to Elasticsearch with a unique ID
post_log_to_elasticsearch() {
  local log_entry="$1"
  local id="$2"
  response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$ES_URL/$id?pretty" -H 'Content-Type: application/json' -d "$log_entry")
  if [ "$response" -eq 201 ]; then
    echo "Successfully posted log entry with ID: $id"
  else
    echo "Failed to post log entry with ID: $id - HTTP $response"
  fi
}

# Initialize counter for unique IDs
counter=1

# Read the log file and extract log entries
jq -c '.hits.hits[]._source' "$LOG_FILE" | while read -r log_entry; do
  post_log_to_elasticsearch "$log_entry" "$counter"
  counter=$((counter + 1))
done
