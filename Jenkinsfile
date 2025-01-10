/* Prerequisites Setup Steps:

1. ARTIFACTORY SETUP:
   a. Get Artifactory Server ID:
      - Go to Jenkins > Manage Jenkins > Configure System
      - Find "JFrog" or "Artifactory" section
      - Click "Add JFrog Platform Instance"
      - Fill in:
        * Instance ID: "artifactory-server" (This is your artifactoryServerId)
        * URL: your-artifactory-url (e.g., https://your-company.jfrog.io)
        * Click "Test Connection"

   b. Create Artifactory Credentials:
      - Go to Jenkins > Manage Jenkins > Credentials > System > Global credentials
      - Click "Add Credentials"
      - Choose "Username with password"
      - Fill in:
        * Username: your-artifactory-username
        * Password: your-artifactory-password
        * ID: "artifactory-credentials" (This is your artifactoryCredentialsId)
        * Description: "Artifactory Login"

2. BITBUCKET SETUP:
   a. Create Bitbucket Credentials:
      - Go to Jenkins > Manage Jenkins > Credentials > System > Global credentials
      - Click "Add Credentials"
      - Choose "Username with password"
      - Fill in:
        * Username: your-bitbucket-username
        * Password: your-bitbucket-app-password
        * ID: "bitbucket-credentials"
        * Description: "Bitbucket Access"

3. REPOSITORY STRUCTURE:
   Your artifacts will be published to:
   - Releases: libs-release-local/your-group-id/your-artifact-id/version
   - Snapshots: libs-snapshot-local/your-group-id/your-artifact-id/version
*/

@Library('maven-pipeline-plugin') _

mavenPipeline {
    // Basic Configuration
    mavenVersion = 'Maven-3.8.6'
    jdkVersion = 'JDK11'
    
    // Source Control Configuration
    gitCredentialsId = 'bitbucket-credentials'
    gitUrl = "https://bitbucket.org/${BITBUCKET_REPO}"
    gitBranch = env.BRANCH_NAME ?: 'main'
    
    // Artifactory Configuration
    artifactoryServerId = 'artifactory-server'
    artifactoryCredentialsId = 'artifactory-credentials'
    
    // Environment Variables
    environment {
        PROJECT_NAME = 'your-project-name'
        JIRA_PROJECT_KEY = 'YOUR-JIRA-KEY'
        BITBUCKET_REPO = 'your-repo-name'
        SPK = 'YOUR-SPK'
        ARTIFACT_VERSION = readMavenPom().getVersion()
        IS_SNAPSHOT = ARTIFACT_VERSION.endsWith('-SNAPSHOT')
    }
    
    // Build Parameters
    parameters {
        choice(name: 'ENVIRONMENT', choices: ['dev', 'qa', 'staging', 'prod'], description: 'Deployment Environment')
        string(name: 'JIRA_TICKET', defaultValue: '', description: 'JIRA ticket number')
    }
    
    // Artifactory Repository Configuration
    artifactoryRepo = [
        // Release Repository Configuration
        releaseRepo: [
            local: 'libs-release-local',
            remote: 'libs-release',
            virtual: 'libs-release-virtual'
        ],
        
        // Snapshot Repository Configuration
        snapshotRepo: [
            local: 'libs-snapshot-local',
            remote: 'libs-snapshot',
            virtual: 'libs-snapshot-virtual'
        ],
        
        // Deployment settings
        deployReleases: true,
        deploySnapshots: true,
        
        // Properties to be added to artifacts
        properties: [
            'project.name': '${PROJECT_NAME}',
            'project.environment': '${ENVIRONMENT}',
            'git.branch': '${GIT_BRANCH}',
            'build.number': '${BUILD_NUMBER}',
            'jira.ticket': '${JIRA_TICKET}',
            'spk': '${SPK}'
        ]
    ]
    
    // Maven goals and options
    goals = 'clean install deploy'
    mavenOpts = '''
        -Dartifactory.publish.artifacts=true
        -Dartifactory.publish.buildInfo=true
        -DskipTests=false
    '''
    
    // Custom Stages
    customStages = [
        validate_setup: {
            script {
                // Validate credentials and connections
                validateSetup()
            }
        },
        
        prepare_build: {
            script {
                // Configure repository based on version type
                configurePaths()
            }
        },
        
        build_and_test: {
            parallel {
                stage('Compile and Package') {
                    steps {
                        sh 'mvn clean package'
                    }
                }
                stage('Unit Tests') {
                    steps {
                        sh 'mvn test'
                    }
                    post {
                        always {
                            junit '**/target/surefire-reports/*.xml'
                        }
                    }
                }
            }
        },
        
        publish_artifact: {
            steps {
                // Deploy to Artifactory
                rtMavenDeployer(
                    id: "MAVEN_DEPLOYER",
                    serverId: artifactoryServerId,
                    releaseRepo: artifactoryRepo.local,
                    snapshotRepo: artifactoryRepo.local
                )
                
                rtMavenRun(
                    tool: 'Maven-3.8.6',
                    pom: 'pom.xml',
                    goals: 'deploy',
                    deployerId: "MAVEN_DEPLOYER"
                )
                
                // Publish build info
                rtPublishBuildInfo(
                    serverId: artifactoryServerId
                )
            }
        },
        
        update_status: {
            script {
                updateBuildStatus()
            }
        }
    ]
    
    // Helper Methods
    void validateSetup() {
        // Validate Artifactory connection
        def artifactoryConnection = rtServer(
            id: "ARTIFACTORY_SERVER",
            serverId: artifactoryServerId,
            credentialsId: artifactoryCredentialsId
        )
        
        if (!artifactoryConnection) {
            error "Failed to connect to Artifactory"
        }
        
        // Validate Bitbucket connection
        withCredentials([usernamePassword(credentialsId: 'bitbucket-credentials', 
                                       usernameVariable: 'BITBUCKET_USER', 
                                       passwordVariable: 'BITBUCKET_PASS')]) {
            def connection = sh(script: """
                curl -s -u ${BITBUCKET_USER}:${BITBUCKET_PASS} \
                https://api.bitbucket.org/2.0/repositories/${BITBUCKET_REPO}
            """, returnStatus: true)
            
            if (connection != 0) {
                error "Failed to connect to Bitbucket"
            }
        }
    }
    
    void configurePaths() {
        if (IS_SNAPSHOT) {
            artifactoryRepo.local = artifactoryRepo.snapshotRepo.local
            artifactoryRepo.remote = artifactoryRepo.snapshotRepo.remote
            artifactoryRepo.virtual = artifactoryRepo.snapshotRepo.virtual
        } else {
            artifactoryRepo.local = artifactoryRepo.releaseRepo.local
            artifactoryRepo.remote = artifactoryRepo.releaseRepo.remote
            artifactoryRepo.virtual = artifactoryRepo.releaseRepo.virtual
        }
    }
    
    void updateBuildStatus() {
        if (currentBuild.resultIsBetterOrEqualTo('SUCCESS')) {
            // Update Bitbucket status
            bitbucketStatusNotify(
                buildState: 'SUCCESSFUL',
                repoSlug: BITBUCKET_REPO,
                commitId: GIT_COMMIT
            )
            
            // Update JIRA if ticket provided
            if (params.JIRA_TICKET) {
                def artifactPath = "/${artifactoryRepo.local}/${PROJECT_NAME}/${ARTIFACT_VERSION}"
                jiraAddComment(
                    idOrKey: params.JIRA_TICKET,
                    site: 'YOUR-JIRA-SITE',
                    body: "Build successful. Artifact published to: ${artifactPath}"
                )
            }
        }
    }
}

/* Required pom.xml Configuration:

<project>
    ...
    <distributionManagement>
        <repository>
            <id>central</id>
            <name>libs-release</name>
            <url>${env.ARTIFACTORY_URL}/libs-release-local</url>
        </repository>
        <snapshotRepository>
            <id>snapshots</id>
            <name>libs-snapshot</name>
            <url>${env.ARTIFACTORY_URL}/libs-snapshot-local</url>
        </snapshotRepository>
    </distributionManagement>
    ...
</project>

Required settings.xml Configuration:

<settings>
    <servers>
        <server>
            <id>central</id>
            <username>${env.ARTIFACTORY_USERNAME}</username>
            <password>${env.ARTIFACTORY_PASSWORD}</password>
        </server>
        <server>
            <id>snapshots</id>
            <username>${env.ARTIFACTORY_USERNAME}</username>
            <password>${env.ARTIFACTORY_PASSWORD}</password>
        </server>
    </servers>
</settings>
*/
