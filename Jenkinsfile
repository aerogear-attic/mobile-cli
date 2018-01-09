node ("go") {
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
      
      stage ("Run") {
        // workaround because of the https://issues.jboss.org/browse/FH-4471 
        sh "mkdir -p /home/jenkins/.kube"
        sh "rm /home/jenkins/.kube/config || true"
        sh "oc config view > /home/jenkins/.kube/config"
        //end of workaround

        sh "./mobile"  
      }
    }
  }
}

