podTemplate(
    // 之前配置的 Kubernetes Cloud Provider
    cloud: 'dev-cluster',
    // 这个 pipeline 执行环境名称
    name: 'cyclone',
    // 运行在带有 always-golang 标签的 Jenkins Slave 上 
    label: 'cyclone',
    idleMinutes: 60,
    containers: [
        // Kubernetes Pod 的配置, 这个 Pod 包含两个容器
        containerTemplate(
            name: 'jnlp',
            alwaysPullImage: true,
            // Jenkins Slave ， 与 Master 通信进程
            image: 'cargo.caicloud.io/circle/jnlp:2.62',
            command: '',
            args: '${computer.jnlpmac} ${computer.name}',
            resourceRequestCpu: '300m',
            resourceLimitCpu: '500m',
            resourceRequestMemory: '300Mi',
            resourceLimitMemory: '500Mi',
        ),
         containerTemplate(
            name: 'zk',
            image: 'cargo.caicloud.io/caicloud/zookeeper:3.4.6',
            ttyEnabled: true,
            command: "",
            args: "",
            resourceRequestCpu: '300m',
            resourceLimitCpu: '500m',
            resourceRequestMemory: '300Mi',
            resourceLimitMemory: '500Mi',
         ),
         containerTemplate(
            name: 'kafka',
            image: 'cargo.caicloud.io/caicloud/kafka:0.10.1.0',
            ttyEnabled: true,
            command: '',
            args: '',
            envVars: [
                containerEnvVar(key: 'KAFKA_ADVERTISED_HOST_NAME', value: '0.0.0.0'),
                containerEnvVar(key: 'KAFKA_ADVERTISED_PORT', value: '9092'),
                containerEnvVar(key: 'KAFKA_ZOOKEEPER_CONNECT', value: 'localhost:2181'),
                containerEnvVar(key: 'KAFKA_ZOOKEEPER_CONNECTION_TIMEOUT_MS', value: '60000'),
            ],
            resourceRequestCpu: '300m',
            resourceLimitCpu: '500m',
            resourceRequestMemory: '300Mi',
            resourceLimitMemory: '500Mi',
        ),     
        containerTemplate(
            name: 'golang',
            image: 'cargo.caicloud.io/caicloud/golang-docker:1.8-17.03',
            ttyEnabled: true,
            command: '',
            args: '',
            envVars: [
                containerEnvVar(key: 'DEBUG', value: 'true'),
                containerEnvVar(key: 'CLAIR_DISABLE', value: 'true'),
                containerEnvVar(key: 'MONGODB_HOST', value: '127.0.0.1:27017'),
                containerEnvVar(key: 'KAFKA_HOST', value: '127.0.0.1:9092'),
                containerEnvVar(key: 'ETCD_HOST', value: 'http://127.0.0.1:2379'),
                containerEnvVar(key: 'REGISTRY_LOCATION', value: 'cargo.caicloud.io'),
                containerEnvVar(key: 'REGISTRY_USERNAME', value: 'caicloudadmin'),
                containerEnvVar(key: 'REGISTRY_PASSWORD', value: 'caicloudadmin'),
                containerEnvVar(key: 'WORKER_IMAGE', value: 'cargo.caicloud.io/caicloud/cyclone-worker:latest'),
                containerEnvVar(key: 'DOCKER_HOST', value: 'unix:///var/run/docker.sock'),
                containerEnvVar(key: 'DOCKER_API_VERSION', value: '1.23'),
            ],
            resourceRequestCpu: '1000m',
            resourceLimitCpu: '2000m',
            resourceRequestMemory: '1000Mi',
            resourceLimitMemory: '2000Mi',
        ),
        containerTemplate(
            name: 'mongo',
            image: 'cargo.caicloud.io/caicloud/mongo:3.0.5',
            ttyEnabled: true,
            command: 'mongod',
            args: '--smallfiles',
            resourceRequestCpu: '300m',
            resourceLimitCpu: '500m',
            resourceRequestMemory: '300Mi',
            resourceLimitMemory: '500Mi',
        ),
        containerTemplate(
            name: 'etcd',
            image: 'cargo.caicloud.io/caicloud/etcd:v3.1.3',
            ttyEnabled: true,
            command: 'etcd',
            args: '-name=etcd0 -advertise-client-urls http://0.0.0.0:2379 -listen-client-urls http://0.0.0.0:2379 -initial-advertise-peer-urls http://127.0.0.1:2380 -listen-peer-urls http://0.0.0.0:2380 -initial-cluster-token etcd-cluster-1 -initial-cluster etcd0=http://127.0.0.1:2380 -initial-cluster-state new',
            resourceRequestCpu: '300m',
            resourceLimitCpu: '500m',
            resourceRequestMemory: '300Mi',
            resourceLimitMemory: '500Mi',
        ),
        containerTemplate(
            name: 'dind', 
            // Jenkins Slave 作业执行环境， 此处为一个 Docker in Docker 环境，用于跑作业
            image: 'cargo.caicloud.io/caicloud/docker:1.11-dind', 
            ttyEnabled: true, 
            command: '', 
            args: '',
            privileged: true,
            resourceRequestCpu: '500m',
            resourceLimitCpu: '1000m',
            resourceRequestMemory: '500Mi',
            resourceLimitMemory: '1000Mi',
        ),
    ],
    volumes: [
        emptyDirVolume(mountPath: '/var/run', memory: true),
    ]
) {
    node('cyclone') {
        stage('Checkout') {
            checkout scm
        }
        container('golang') {
            stage('Run e2e test') {
                ansiColor('xterm') {
                    sh('''
                        set -e
                        cyclone_pid=$(ps -ef | grep cyclone-server | grep -v "grep" | awk '{print $1}')
                        if [[ -n "${cyclone_pid}" ]]; then
                            kill -9 ${cyclone_pid}
                        fi

                        # get host ip
                        HOST_IP=$(ifconfig eth0 | grep 'inet addr:'| grep -v '127.0.0.1' | cut -d: -f2 | awk '{ print $1}')
                        export CYCLONE_SERVER=http://${HOST_IP}:7099
                        export LOG_SERVER=ws://${HOST_IP}:8000/ws
                        
                        mkdir -p /go/src/github.com/caicloud
                        ln -sf $(pwd) /go/src/github.com/caicloud/cyclone
                        cd /go/src/github.com/caicloud/cyclone
                        echo "buiding server"
                        go build -i -v -o cyclone-server github.com/caicloud/cyclone/cmd/server

                        echo "buiding worker"
                        go build -i -v -o cyclone-worker github.com/caicloud/cyclone/cmd/worker 
                        docker build -t ${WORKER_IMAGE} -f Dockerfile.worker .

                        echo "start server"
                        ./cyclone-server --cloud-auto-discovery=false --log-force-color=true &

                        echo "testing ..."
                        # go test compile
                        go test -i ./tests/...
                        # go test
                        go test -v ./tests/service 
                        go test -v ./tests/version 
                        go test -v ./tests/yaml
                    ''')
                }
            }

            stage("Build image and push") {
                docker.build("caicloud/cyclone-server:$env.BUILD_NUMBER", "-f Dockerfile.server .")
                docker.build("caicloud/cyclone-worker:$env.BUILD_NUMBER", "-f Dockerfile.worker .")

                docker.withRegistry("https://cargo.caicloudprivatetest.com", "cargo-private-admin") {
                    docker.image("caicloud/cyclone-server:$env.BUILD_NUMBER").push()
                    docker.image("caicloud/cyclone-worker:$env.BUILD_NUMBER").push()
                }
            }
        }
    }
}
