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
          sh "go get github.com/golang/dep/cmd/dep"
          sh "dep ensure"
          sh "go get golang.org/x/tools/cmd/cover"
          sh "go get github.com/mattn/goveralls"
        }

        withCredentials([string(credentialsId: "coveralls_io", variable: 'COVERALLS_TOKEN')]) {
          stage ("Build") {
            sh "make coveralls_build COVERALLS_TOKEN=${COVERALLS_TOKEN}"
            sh "go test -timeout 30m -c ./integration"
          }
        }
    
        stage ("Archive") {
          sh "mkdir out"
          sh "cp mobile out/"
          sh "cp integration.test out/"
          sh "cp -R integration out/integration"
          archiveArtifacts artifacts: 'out/**'
        }

      }
    }
  }
}

node("ocp-slave") {
    def project = sanitizeObjectName("mobile-cli-${env.CHANGE_AUTHOR}-${env.BUILD_TAG}")
    
    stage ("Integration test") {
        step([$class: 'WsCleanup'])
        sh "wget ${env.BUILD_URL}/artifact/*zip*/archive.zip"
        sh "unzip archive.zip"
        dir("archive/out"){
            sh "oc whoami"
            sh "oc new-project ${project}"
            sh "chmod +x *"
            sh "./integration.test -test.short -test.v -goldenFiles=`pwd`/integration -prefix=test-${sanitizeObjectName(env.BRANCH_NAME)}-build-$BUILD_NUMBER -namespace=`oc project -q` -executable=`pwd`/mobile"
        }
    }            
    stage ("Clear Project") {
        sh "oc delete project ${project}"
    }
}
