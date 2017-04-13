podTemplate(
    // 之前配置的 Kubernetes Cloud Provider
    cloud: "dev-cluster",
    // 这个 pipeline 执行环境名称
    name: "cyclone",
    // 运行在带有 always-golang 标签的 Jenkins Slave 上 
    label: 'cyclone',
    containers: [
        // Kubernetes Pod 的配置, 这个 Pod 包含两个容器
        containerTemplate(
            name: 'jnlp',
            alwaysPullImage: true,
            // Jenkins Slave ， 与 Master 通信进程
            image: 'cargo.caicloud.io/circle/jnlp:2.62',
            command: "",
            args: '${computer.jnlpmac} ${computer.name}',
            resourceRequestCpu: '300m',
            resourceLimitCpu: '5000m',
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
            command: "",
            args: "",
            envVars: [
                containerEnvVar(key: "KAFKA_ADVERTISED_HOST_NAME", value: "0.0.0.0"),
                containerEnvVar(key: "KAFKA_ADVERTISED_PORT", value: "9092"),
                containerEnvVar(key: "KAFKA_ZOOKEEPER_CONNECT", value: "localhost:2181"),
            ],
            resourceRequestCpu: '300m',
            resourceLimitCpu: '500m',
            resourceRequestMemory: '300Mi',
            resourceLimitMemory: '500Mi',
        ),     
        containerTemplate(
            name: 'golang',
            // Jenkins Slave 作业执行环境， 此处为一个 Docker in Docker 环境，用于跑作业
            image: 'cargo.caicloud.io/caicloud/golang:1.7',
            ttyEnabled: true,
            command: "",
            args: "",
            resourceRequestCpu: '500m',
            resourceLimitCpu: '1000m',
            resourceRequestMemory: '2000Mi',
            resourceLimitMemory: '2000Mi',
        ),
        containerTemplate(
            name: 'mongo',
            image: 'cargo.caicloud.io/caicloud/mongo:3.0.5',
            ttyEnabled: true,
            command: "mongod",
            args: "--smallfiles",
            resourceRequestCpu: '300m',
            resourceLimitCpu: '500m',
            resourceRequestMemory: '300Mi',
            resourceLimitMemory: '500Mi',
        ),
        containerTemplate(
            name: 'etcd',
            image: 'cargo.caicloud.io/caicloud/etcd:v3.1.3',
            ttyEnabled: true,
            command: "etcd",
            args: "-name=etcd0 -advertise-client-urls http://0.0.0.0:2379 -listen-client-urls http://0.0.0.0:2379 -initial-advertise-peer-urls http://127.0.0.1:2380 -listen-peer-urls http://0.0.0.0:2380 -initial-cluster-token etcd-cluster-1 -initial-cluster etcd0=http://127.0.0.1:2380 -initial-cluster-state new",
            resourceRequestCpu: '300m',
            resourceLimitCpu: '500m',
            resourceRequestMemory: '300Mi',
            resourceLimitMemory: '500Mi',
        )
    ]
) {
    node("cyclone") {
        stage("Checkout") {
            checkout scm
        }

        stage("Run e2e test") {
            sh("echo e2e")
        }
    }
}
