pipeline {
    agent { docker { image 'golang:1.14' } }
    environment {
      XDG_CACHE_HOME = "/tmp/cache"
    }
    stages {
        stage('build') {
            steps {
                sh 'go version'
                sh 'go get github.com/axw/gocov/...'
                sh 'go get github.com/AlekSi/gocov-xml'
                sh 'gocov test -race ./... | gocov-xml  > coverage.xml'
                cobertura coberturaReportFile: 'coverage.xml'
            }
        }
    }
}
