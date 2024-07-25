#!groovy

ansiColor('xterm') {
  node('executor') {

  checkout scm

  def authorName  = sh(returnStdout: true, script: 'git --no-pager show --format="%an" --no-patch')
  def serviceName = env.JOB_NAME.tokenize("/")[1]

  try {
    stage("Build Container") {
          sh "docker build ."
    }

    stage("Run Tests") {
        sh "go test ./..."
    }

  } catch (e) {
    slackSend(color: '#b20000', message: "FAILED: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL}) by ${authorName}")
    throw e
  }

  slackSend(color: '#006600', message: "SUCCESSFUL: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL}) by ${authorName}")
  }
}