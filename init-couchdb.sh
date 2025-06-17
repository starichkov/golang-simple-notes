#!/bin/bash

# Wait for CouchDB to be ready
echo "Waiting for CouchDB to be ready..."
until curl -s http://admin:password@couchdb:5984 > /dev/null; do
  sleep 1
done

# Create system databases
echo "Creating system databases..."
curl -X PUT http://admin:password@couchdb:5984/_users
curl -X PUT http://admin:password@couchdb:5984/_replicator
curl -X PUT http://admin:password@couchdb:5984/_global_changes

# Create the notes database
echo "Creating notes database..."
curl -X PUT http://admin:password@couchdb:5984/notes

echo "CouchDB initialization completed."