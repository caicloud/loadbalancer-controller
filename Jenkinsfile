def REGISTRY = "cargo.caicloudprivatetest.com"

podTemplate(
    cloud: 'dev-cluster',
    namespace: 'kube-system',
    name: 'loadbalancer-controller',
    label: 'loadbalancer-controller',
    containers: [
        containerTemplate(
            name: 'jnlp',
            image: "cargo.caicloud.io/circle/jnlp:2.62",
            alwaysPullImage: true,
            command: '',
            args: '${computer.jnlpmac} ${computer.name}',
        ),
        containerTemplate(
            name: 'dind', 
            image: "cargo.caicloud.io/caicloud/docker:17.03-dind", 
            alwaysPullImage: true,
            ttyEnabled: true,
            command: '', 
            args: '--host=unix:///home/jenkins/docker.sock',
            privileged: true,
        ),
        containerTemplate(
            name: 'golang',
            image: "cargo.caicloud.io/caicloud/golang-docker:1.8.1-17.05",
            alwaysPullImage: true,
            ttyEnabled: true,
            command: '',
            args: '',
            envVars: [
                containerEnvVar(key: 'DOCKER_HOST', value: 'unix:///home/jenkins/docker.sock'),
                containerEnvVar(key: 'DOCKER_API_VERSION', value: '1.26'),
                containerEnvVar(key: 'WORKDIR', value: '/go/src/github.com/caicloud/loadbalancer-controller'),
                containerEnvVar(key: 'VERSION', value: "${BUILD_NUMBER}")
            ],
        ),
    ]
) {
    node('loadbalancer-controller') {
        stage('Checkout') {
            checkout scm
        }
        container('golang') {
            ansiColor('xterm') {

                stage("Complie") {
                    sh('''
                        set -e 
                        mkdir -p $(dirname ${WORKDIR})
                        rm -rf ${WORKDIR}
                        ln -sf $(pwd) ${WORKDIR}
                        cd ${WORKDIR}
                        make build-linux
                    ''')
                }

                stage('Run e2e test') {
                    sh('''
                        echo "skip e2e test: it is not accomplished now"
                    ''')
                }
            }

            stage("Build image and publish") {
                docker.withRegistry("https://${REGISTRY}", "cargo-private-admin") {
                    sh('''
                        cd ${WORKDIR} && make push
                    ''') 
                }
            }
        }

        stage('Deploy') {
            echo "Coming soon..."
        }
    }
}
