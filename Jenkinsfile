node ("go") {
  sh "mkdir -p src/github.com/aerogear/mobile-cli"
  withEnv(["GOPATH=${env.WORKSPACE}/","PATH=${env.PATH}:${env.WORKSPACE}/bin"]) {
    dir ("src/github.com/aerogear/mobile-cli") {

      stage("Checkout") {
        checkout scm
      }
      
      stage ("Build") {
        sh "make build"
      }
      
    }
  }
}

