docker rm -f redistest

docker pull garsonyang/redistest

docker run \
    -d \
    -p 6379:6379 \
    --name redistest \
    --network myNet \
garsonyang/redistest