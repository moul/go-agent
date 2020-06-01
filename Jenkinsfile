pipeline {
    agent {
    	docker {
    	    image 'golang:1.14'
	}
    }
    environment {
      GOLANGCI_VERSION = "v1.27.0"
      XDG_CACHE_HOME = "/tmp/cache"
    }
    stages {
        stage('downloads') {
	    steps {
	    	sh "ls -al /go"
        	sh 'go version'

        	sh 'go get golang.org/x/lint/golint'

                sh 'curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s ${GOLANGCI_VERSION}'
                sh './bin/golangci-lint --version'

                sh 'go get github.com/axw/gocov/...'
                sh 'go get github.com/AlekSi/gocov-xml'
            }
        }
        stage('static analysis') {
	    steps {
	    	// sh 'cloc --by-file --xml --out sloccount.xml .'
	    	// sloccountPublish encoding: 'UTF-8', pattern: '**/sloccount.xml'

	    	sh 'echo > golint.xml && golint -min_confidence 0.3 ./... >> golint.xml'
	    	sh './bin/golangci-lint run --out-format checkstyle ./... > golangci.xml'
	    	sh 'ls -al'
	    	recordIssues(
	    	    enabledForFailure: true,
	    	    aggregatingResults: true,
	    	    tools: [
	    	        checkStyle(reportEncoding: 'UTF-8', pattern: '**/golangci.xml'),
	    	        goLint(pattern: '**/golint.xml')
	    	    ]
	    	)
	    }
        }
        stage('tests') {
            steps {
                sh 'gocov test -race ./... | gocov-xml  > coverage.xml'
                cobertura coberturaReportFile: 'coverage.xml'
            }
        }
    }
}
