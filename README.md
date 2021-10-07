## Api gateway 
### Description
This repository contains the public api gateway project it is a go web backend that handle communication and real-time between clusters in private network and the public browser web application, also contains managment apis to handle add new cluster , users ,permissions etc. 
### Technologies
- Go lang without frameworks just http handler 
- etcd distributed database https://etcd.io/

### Set up
```
#Uncomment the following variables after adding the right value
#export NODE1=YOUR_LOCAL_HOST_IP
#export DATA_DIR=YOUR_LOCAL_HOST_DATA_PATH
#export REGISTRY=quay.io/coreos/etcd
#export KARBO_DEFAULT_USER_NAME=YOUR_ADMIN_USER_NAME
#export KARBO_DEFAULT_PASSWORDYOUR_ADMIN_PASSWORD
#export SERVER_URL=YOUR_HOST_NAME_FOR_API_INCLUDE_SECHEMA

#Run etcd db 

docker run -d   -p 2379:2379   -p 2380:2380   --volume=${DATA_DIR}:/etcd-data  \
--name etcd ${REGISTRY}:latest   /usr/local/bin/etcd   --data-dir=/etcd-data --name node1  \
--initial-advertise-peer-urls http://${NODE1}:2380 --listen-peer-urls http://0.0.0.0:2380   \
--advertise-client-urls http://${NODE1}:2379 --listen-client-urls http://0.0.0.0:2379  \
--initial-cluster node1=http://${NODE1}:2380

#Build app and run 
docker build -t carbo-api-server:latest .
docker run -d --restart unless-stopped -p 8099:8099 -e ACCESS_SECRET=your-secret-key -e ETCD_NODES=${NODE1}:2379 -e SERVER_URL=${SERVER_URL} -e DEFAULT_USER_NAME=${KARBO_DEFAULT_USER_NAME} -e DEFAULT_PASSWORD=${KARBO_DEFAULT_PASSWORD}  carbo-api-server:latest
```
