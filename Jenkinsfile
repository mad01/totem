#!groovy

pipeline {
  agent {
    label "kubernetes-tools-medium"
  }
  stages {
    stage("Build") {
      steps {
        sh "make container"
      }
    }
    stage("Publish") {
      when {
        branch 'master'
      }
      steps {
          sh "make container-push"
      }
    }
  }
}
