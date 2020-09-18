docker rm -f gatewaytest

docker pull garsonyang/gateway

docker run \
    -d \
    -e ADDR=':443' \
    -v /etc/letsencrypt:/etc/letsencrypt:ro \
    -e TLSCERT='/etc/letsencrypt/live/api.garson.me/fullchain.pem' \
    -e TLSKEY='/etc/letsencrypt/live/api.garson.me/privkey.pem' \
    -e SESSIONKEY='tiancy' \
    -e DSN='root:password@tcp(mysqltest:3306)/demo' \
    -e REDISADDR='redistest:6379' \
    -e MESSAGESADDR='http://messagingservice:4001' \
    -e SUMMARYADDR='http://summaryservice:5001' \
    -p 443:443 \
    --network myNet \
    --name gatewaytest \
    garsonyang/gateway

exit