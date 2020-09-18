docker rm -f summaryclient

docker rm -f messageclient

docker pull garsonyang/messageclient

docker run -d \
    --name messageclient \
    -p 80:80 -p 443:443 \
    -v /etc/letsencrypt:/etc/letsencrypt:ro \
    garsonyang/messageclient