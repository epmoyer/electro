#!/bin/bash -x
# go run . --project ../test/data/raw/test_cases/static/electro.json
# Really this should be part of a test suite which copies the raw to the processed dir first.
cp -r ../test/data/raw/test_cases/static ../test/data/processed/test_cases/static/incoming
go run . --project ../test/data/processed/test_cases/static/incoming/electro.json