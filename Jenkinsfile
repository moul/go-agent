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
        stage('Downloads') {
	    steps {
        	sh 'go version'

		sh 'go mod download'
		sh 'go mod tidy'

        	sh 'go get golang.org/x/lint/golint'

                sh 'curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s ${GOLANGCI_VERSION}'
                sh './bin/golangci-lint --version'
                sh './bin/golangci-lint run -h | grep concurrency'

                sh 'go get github.com/axw/gocov/... github.com/AlekSi/gocov-xml'
            }
        }
        stage('Checks') {
	    parallel {
		stage('Static analysis') {
		    steps {
			// sh 'cloc --by-file --xml --out sloccount.xml .'
			// sloccountPublish encoding: 'UTF-8', pattern: '**/sloccount.xml'

			sh 'echo > golint.xml && golint -min_confidence 0.3 ./... | tee -a golint.xml'
			sh './bin/golangci-lint run --out-format checkstyle ./... | tee golangci.xml'
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
		stage('Tests') {
		    steps {
			sh 'gocov test -race ./... | gocov-xml  > coverage.xml'
			cobertura coberturaReportFile: 'coverage.xml'
		    }
		}
	    }
	}
    }
}
