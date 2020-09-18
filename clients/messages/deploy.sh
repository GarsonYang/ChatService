sh build.sh

docker push garsonyang/messageclient

ssh -i "~/garson.pem" ec2-user@ec2-34-217-66-96.us-west-2.compute.amazonaws.com < run.sh