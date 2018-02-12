#!groovy

// https://github.com/feedhenry/fh-pipeline-library
@Library('fh-pipeline-library') _

stage('Trust') {
    enforceTrustedApproval('aerogear')
}

@NonCPS
String sanitizeObjectName(String s) {
    s.replace('_', '-')
        .replace('.', '-')
        .toLowerCase()
        .reverse()
        .take(23)
        .replaceAll("^-+", "")
        .reverse()
        .replaceAll("^-+", "")
}

// The jnlp container will be the one that the configured node will run on
// you can define more containers and run them along-side
def goSlaveContainer = containerTemplate(
  name: 'jnlp',
  image: 'docker.io/fhwendy/jenkins-slave-go-centos7:201801081225',
  args: '${computer.jnlpmac} ${computer.name}',
  ttyEnabled: false)

podTemplate(label: 'mobile-cli-go', cloud: "openshift", containers: [goSlaveContainer]) {
  node ("mobile-cli-go") {
    sh "mkdir -p src/github.com/aerogear/mobile-cli"
    withEnv(["GOPATH=${env.WORKSPACE}/","PATH=${env.PATH}:${env.WORKSPACE}/bin"]) {
      dir ("src/github.com/aerogear/mobile-cli") {
        stage("Checkout") {
          checkout scm
        }

        stage ("Setup") {
          sh "glide install"
        }

        stage ("Build") {
          sh "make build"
        }
        def project = sanitizeObjectName("mobile-cli-${env.CHANGE_AUTHOR}-${env.BUILD_TAG}")
        stage ("Run") {
          // workaround because of the https://issues.jboss.org/browse/FH-4471
          sh "mkdir -p /home/jenkins/.kube"
          sh "rm /home/jenkins/.kube/config || true"
          sh "oc config view > /home/jenkins/.kube/config"
          sh "oc new-project ${project}"
          //end of workaround

          sh "./mobile"
        }

        stage ("Integration") {
          sh "oc project ${project}"
          sh "go test -timeout 30m -c ./integration"
          sh "./integration.test -test.short -test.v -prefix=test-${sanitizeObjectName(env.BRANCH_NAME)}-build-$BUILD_NUMBER -namespace=`oc project -q` -executable=`pwd`/mobile"
        }

        stage ("Archive") {
          sh "mkdir out"
          sh "cp mobile out/"
          sh "cp integration.test out/"
          sh "cp -R integration out/integration"
          archiveArtifacts artifacts: 'out/**'
        }

        stage ("Clear Project") {
            sh "oc delete project ${project}"
        }
      }
    }
  }
}
