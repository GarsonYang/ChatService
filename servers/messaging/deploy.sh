sh build.sh

docker push garsonyang/messagingservice

ssh -i "~/garson.pem" ec2-user@ec2-18-236-87-183.us-west-2.compute.amazonaws.com < run.sh