docker rm -f mongodb
docker rm -f rabbitmq

docker run -d \
    -p 27017:27017 \
    --name mongodb \
    --network myNet \
    mongo

docker run -d \
    -p 5672:5672 \
    -p 15672:15672 \
    --name rabbitmq \
    --network myNet \
    rabbitmq:3-management