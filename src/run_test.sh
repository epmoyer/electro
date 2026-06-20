#!/bin/bash -x
# go run . --project ../test/data/raw/test_cases/static/electro.json

# Really this should be part of a test suite which copies the raw to the processed dir first.

# rsync -a --delete ../test/data/raw/test_cases/static/ ../test/data/processed/test_cases/static/incoming/
# go run . --project ../test/data/processed/test_cases/static/incoming/electro.json

rsync -a --delete ../test/data/raw/test_cases/single_file/ ../test/data/processed/test_cases/single_file/incoming/
go run . --noembed --project ../test/data/raw/test_cases/single_file/electro.json
