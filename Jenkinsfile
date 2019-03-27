#!/usr/bin/env groovy
@Library('com.fxinnovation.fxinnovation-itoa.tools.library') _

pipelineItoa(
        stageTest: {
            sh './docker.sh make tools lint vet test-it-test-coverage'
            cobertura coberturaReportFile: 'target/coverage-test-it-test.xml'
        },
        stagePackage: {
            sh './docker.sh make tools build'
            archiveArtifacts 'alertmanager-webhook-rocketchat'
        }
)
