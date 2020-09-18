docker rm -f messagingservice

docker pull garsonyang/messagingservice

docker run -d \
-p 4001:4001 \
-e PORT='4001' \
--name messagingservice \
--network myNet \
garsonyang/messagingservice 