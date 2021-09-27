export NODE1=YOUR_LOCAL_HOST_IP
export DATA_DIR=YOUR_LOCAL_HOST_PATH
export REGISTRY=quay.io/coreos/etcd
export KARBO_DEFAULT_USER_NAME
export KARBO_DEFAULT_PASSWORD
docker run -d   -p 2379:2379   -p 2380:2380   --volume=${DATA_DIR}:/etcd-data  \
--name etcd ${REGISTRY}:latest   /usr/local/bin/etcd   --data-dir=/etcd-data --name node1  \
--initial-advertise-peer-urls http://${NODE1}:2380 --listen-peer-urls http://0.0.0.0:2380   \
--advertise-client-urls http://${NODE1}:2379 --listen-client-urls http://0.0.0.0:2379  \
--initial-cluster node1=http://${NODE1}:2380

docker build -t carbo-api-server:latest .

docker run -d --restart unless-stopped \
-p 8099:8099 \
-e ACCESS_SECRET=your-secret-key -e ETCD_NODES=${NODE1}:2379 -e DEFAULT_USER_NAME=${KARBO_DEFAULT_USER_NAME} -e DEFAULT_PASSWORD=${KARBO_DEFAULT_PASSWORD} \
 carbo-api-server:latest