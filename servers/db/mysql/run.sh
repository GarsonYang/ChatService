docker rm -f mysqltest

docker pull garsonyang/mysqltest

docker run -d \
-p 3306:3306 \
--name mysqltest \
-e MYSQL_ROOT_PASSWORD='password' \
-e MYSQL_DATABASE=demo \
--network myNet \
garsonyang/mysqltest 