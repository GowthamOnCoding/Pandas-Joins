-- DDL Statements
CREATE TABLE TC_MASTER (
    TC_ID VARCHAR(50) PRIMARY KEY,
    TC_NAME VARCHAR(200) NOT NULL,
    DESCRIPTION TEXT,
    FLAG VARCHAR(20) NOT NULL,
    CREATED_BY VARCHAR(50),
    CREATED_DATE TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    MODIFIED_BY VARCHAR(50),
    MODIFIED_DATE TIMESTAMP
);

CREATE TABLE PARAMETER_SCHEMA (
    SCHEMA_ID VARCHAR(100) PRIMARY KEY,
    SCHEMA_VERSION INT NOT NULL,
    SCHEMA_DEFINITION TEXT NOT NULL,
    IS_ACTIVE BOOLEAN DEFAULT TRUE,
    DESCRIPTION TEXT,
    CREATED_DATE TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    MODIFIED_DATE TIMESTAMP,
    UNIQUE (SCHEMA_ID, SCHEMA_VERSION)
);

CREATE TABLE STEP_CONFIG (
    STEP_NAME VARCHAR(100) PRIMARY KEY,
    DESCRIPTION TEXT,
    TIMEOUT_SECONDS INT NOT NULL DEFAULT 300,
    MAX_RETRIES INT NOT NULL DEFAULT 3,
    CREATED_DATE TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    MODIFIED_DATE TIMESTAMP
);

CREATE TABLE STEP_SCHEMA_MAPPING (
    STEP_NAME VARCHAR(100),
    SCHEMA_ID VARCHAR(100),
    IS_REQUIRED BOOLEAN DEFAULT TRUE,
    SEQUENCE_NO INT NOT NULL,
    CREATED_DATE TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (STEP_NAME, SCHEMA_ID),
    FOREIGN KEY (STEP_NAME) REFERENCES STEP_CONFIG(STEP_NAME),
    FOREIGN KEY (SCHEMA_ID) REFERENCES PARAMETER_SCHEMA(SCHEMA_ID)
);

CREATE TABLE TC_STEPS (
    STEP_ID INT,
    TC_ID VARCHAR(50),
    STEP_NAME VARCHAR(100) NOT NULL,
    PARAMETERS TEXT,
    SEQUENCE_NO INT NOT NULL,
    STATUS VARCHAR(20),
    PRIMARY KEY (STEP_ID, TC_ID),
    FOREIGN KEY (TC_ID) REFERENCES TC_MASTER(TC_ID),
    FOREIGN KEY (STEP_NAME) REFERENCES STEP_CONFIG(STEP_NAME)
);

CREATE TABLE TC_EXECUTION_LOG (
    EXECUTION_ID BIGSERIAL PRIMARY KEY,
    TC_ID VARCHAR(50),
    STEP_ID INT,
    START_TIME TIMESTAMP,
    END_TIME TIMESTAMP,
    STATUS VARCHAR(20),
    ERROR_MESSAGE TEXT,
    FOREIGN KEY (TC_ID) REFERENCES TC_MASTER(TC_ID)
);

-- Create indexes
CREATE INDEX idx_tc_master_flag ON TC_MASTER(FLAG);
CREATE INDEX idx_tc_steps_tcid ON TC_STEPS(TC_ID);
CREATE INDEX idx_execution_log_tcid ON TC_EXECUTION_LOG(TC_ID);
CREATE INDEX idx_parameter_schema_active ON PARAMETER_SCHEMA(IS_ACTIVE);

-- Insert sample schemas
INSERT INTO PARAMETER_SCHEMA (SCHEMA_ID, SCHEMA_VERSION, SCHEMA_DEFINITION, DESCRIPTION) VALUES
('BASE_PARAMETERS', 1, '{
    "type": "object",
    "required": ["environment", "region"],
    "properties": {
        "environment": {
            "type": "string",
            "enum": ["dev", "test", "prod"]
        },
        "region": {
            "type": "string",
            "pattern": "^[A-Z]{2}-[A-Z]+-\\d+$"
        }
    }
}', 'Base parameters required for all steps');

INSERT INTO PARAMETER_SCHEMA (SCHEMA_ID, SCHEMA_VERSION, SCHEMA_DEFINITION, DESCRIPTION) VALUES
('AIT_PARAMETERS', 1, '{
    "type": "object",
    "required": ["aitNumber"],
    "properties": {
        "aitNumber": {
            "type": "string",
            "pattern": "^AIT\\d{6}$"
        }
    }
}', 'AIT specific parameters');

INSERT INTO PARAMETER_SCHEMA (SCHEMA_ID, SCHEMA_VERSION, SCHEMA_DEFINITION, DESCRIPTION) VALUES
('JAVA_PROCESS_PARAMETERS', 1, '{
    "type": "object",
    "required": ["processName", "server"],
    "properties": {
        "processName": {
            "type": "string",
            "minLength": 1
        },
        "server": {
            "type": "string",
            "format": "hostname"
        },
        "timeout": {
            "type": "integer",
            "minimum": 0,
            "maximum": 3600
        }
    }
}', 'Java process execution parameters');

-- Insert step configurations
INSERT INTO STEP_CONFIG (STEP_NAME, DESCRIPTION, TIMEOUT_SECONDS, MAX_RETRIES) VALUES
('DELETE_INSERT_AIT_SCAN_WINDOW', 'Deletes and inserts AIT scan window records', 300, 3);

INSERT INTO STEP_CONFIG (STEP_NAME, DESCRIPTION, TIMEOUT_SECONDS, MAX_RETRIES) VALUES
('INVOKE_JAVA_PROCESS', 'Invokes a Java process on a remote server', 600, 2);

-- Insert schema mappings
INSERT INTO STEP_SCHEMA_MAPPING (STEP_NAME, SCHEMA_ID, IS_REQUIRED, SEQUENCE_NO) VALUES
('DELETE_INSERT_AIT_SCAN_WINDOW', 'BASE_PARAMETERS', true, 1),
('DELETE_INSERT_AIT_SCAN_WINDOW', 'AIT_PARAMETERS', true, 2),
('INVOKE_JAVA_PROCESS', 'BASE_PARAMETERS', true, 1),
('INVOKE_JAVA_PROCESS', 'JAVA_PROCESS_PARAMETERS', true, 2);

-- Insert sample test case
INSERT INTO TC_MASTER (TC_ID, TC_NAME, FLAG, DESCRIPTION) VALUES
('TC001', 'AIT Scan Window Update', 'ENABLED', 'Updates AIT scan window data');

INSERT INTO TC_STEPS (STEP_ID, TC_ID, STEP_NAME, PARAMETERS, SEQUENCE_NO) VALUES
(1, 'TC001', 'DELETE_INSERT_AIT_SCAN_WINDOW', '{
    "environment": "prod",
    "region": "US-EAST-1",
    "aitNumber": "AIT123456"
}', 1);

INSERT INTO TC_STEPS (STEP_ID, TC_ID, STEP_NAME, PARAMETERS, SEQUENCE_NO) VALUES
(2, 'TC001', 'INVOKE_JAVA_PROCESS', '{
    "environment": "prod",
    "region": "US-EAST-1",
    "processName": "DataCleanup",
    "server": "app-server-1.example.com",
    "timeout": 300
}', 2);





// application.yaml
spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/testcasedb
    username: ${DB_USERNAME}
    password: ${DB_PASSWORD}
    hikari:
      maximum-pool-size: 10
      minimum-idle: 5
      idle-timeout: 300000
  jpa:
    hibernate:
      ddl-auto: validate
    properties:
      hibernate:
        dialect: org.hibernate.dialect.PostgreSQLDialect

camel:
  springboot:
    name: TestCaseProcessor

logging:
  level:
    root: INFO
    com.example.testcaseprocessor: DEBUG

// Domain Models
@Data
@Entity
@Table(name = "PARAMETER_SCHEMA")
public class ParameterSchema {
    @Id
    private String schemaId;
    private Integer schemaVersion;
    @Column(columnDefinition = "TEXT")
    private String schemaDefinition;
    private boolean isActive;
    private String description;
    private LocalDateTime createdDate;
    private LocalDateTime modifiedDate;
}

@Data
@Entity
@Table(name = "STEP_CONFIG")
public class StepConfig {
    @Id
    private String stepName;
    private String description;
    private Integer timeoutSeconds;
    private Integer maxRetries;
    private LocalDateTime createdDate;
    private LocalDateTime modifiedDate;
    
    @OneToMany(mappedBy = "stepName", fetch = FetchType.EAGER)
    @OrderBy("sequenceNo")
    private List<StepSchemaMapping> schemaMappings;
}

@Data
@Entity
@Table(name = "STEP_SCHEMA_MAPPING")
@IdClass(StepSchemaMappingId.class)
public class StepSchemaMapping {
    @Id
    private String stepName;
    @Id
    private String schemaId;
    private boolean isRequired;
    private Integer sequenceNo;
    private LocalDateTime createdDate;
}

// Repositories
@Repository
public interface ParameterSchemaRepository extends JpaRepository<ParameterSchema, String> {
    @Query("""
        SELECT ps FROM ParameterSchema ps
        JOIN StepSchemaMapping ssm ON ps.schemaId = ssm.schemaId
        WHERE ssm.stepName = :stepName AND ps.isActive = true
        ORDER BY ssm.sequenceNo
    """)
    List<ParameterSchema> findActiveSchemasByStepName(String stepName);
}

@Repository
public interface StepConfigRepository extends JpaRepository<StepConfig, String> {
}

// Services
@Service
@Slf4j
public class ParameterValidationService {
    private final JsonSchemaFactory factory = JsonSchemaFactory.byDefault();
    private final Map<String, JsonSchema> schemaCache = new ConcurrentHashMap<>();
    private final ParameterSchemaRepository schemaRepository;
    
    public void validateParameters(String stepName, Map<String, String> parameters) {
        List<ParameterSchema> schemas = schemaRepository.findActiveSchemasByStepName(stepName);
        List<ValidationError> errors = new ArrayList<>();
        
        for (ParameterSchema schema : schemas) {
            try {
                JsonSchema jsonSchema = getOrCreateSchema(schema);
                JsonNode parametersNode = new ObjectMapper().valueToTree(parameters);
                ProcessingReport report = jsonSchema.validate(parametersNode);
                
                if (!report.isSuccess()) {
                    errors.add(new ValidationError(schema.getSchemaId(), report));
                }
            } catch (Exception e) {
                errors.add(new ValidationError(schema.getSchemaId(), e.getMessage()));
            }
        }
        
        if (!errors.isEmpty()) {
            throw new ParameterValidationException(stepName, errors);
        }
    }
    
    private JsonSchema getOrCreateSchema(ParameterSchema schema) {
        String cacheKey = schema.getSchemaId() + "_v" + schema.getSchemaVersion();
        return schemaCache.computeIfAbsent(cacheKey, k -> {
            try {
                return factory.getJsonSchema(JsonLoader.fromString(schema.getSchemaDefinition()));
            } catch (Exception e) {
                throw new SchemaConfigurationException("Invalid schema: " + schema.getSchemaId(), e);
            }
        });
    }
}

// Processors
@Component
@Slf4j
public class TestCaseStepProcessor {
    private final ParameterValidationService validationService;
    private final StepConfigRepository stepConfigRepository;
    
    @Transactional
    public void processStep(String stepName, Map<String, String> parameters) {
        StepConfig stepConfig = stepConfigRepository.findById(stepName)
            .orElseThrow(() -> new IllegalArgumentException("Unknown step: " + stepName));
            
        validationService.validateParameters(stepName, parameters);
        
        // Execute step with retry logic
        RetryPolicy<Object> retryPolicy = RetryPolicy.builder()
            .handle(Exception.class)
            .withMaxRetries(stepConfig.getMaxRetries())
            .withDelay(Duration.ofSeconds(1))
            .withMaxDuration(Duration.ofSeconds(stepConfig.getTimeoutSeconds()))
            .onRetry(e -> log.warn("Retrying step: {}", stepName))
            .onFailure(e -> log.error("Step failed after retries: {}", stepName))
            .build();
            
        Failsafe.with(retryPolicy).run(() -> executeStep(stepName, parameters));
    }
    
    private void executeStep(String stepName, Map<String, String> parameters) {
        switch (stepName) {
            case "DELETE_INSERT_AIT_SCAN_WINDOW":
                executeAitScanWindow(parameters);
                break;
            case "INVOKE_JAVA_PROCESS":
                executeJavaProcess(parameters);
                break;
            default:
                throw new IllegalArgumentException("Unsupported step: " + stepName);
        }
    }
    
    private void executeAitScanWindow(Map<String, String> parameters) {
        // Implementation
    }
    
    private void executeJavaProcess(Map<String, String> parameters) {
        // Implementation
    }
}

// Camel Route
@Component
public class TestCaseProcessorRoute extends RouteBuilder {
    @Override
    public void configure() {
        errorHandler(deadLetterChannel("direct:error")
            .maximumRedeliveries(3)
            .redeliveryDelay(1000)
            .backOffMultiplier(2)
            .useExponentialBackOff());
            
        from("sql:SELECT * FROM TC_MASTER WHERE FLAG = 'ENABLED'?delay=5000")
            .routeId("testCaseProcessor")
            .split(body())
            .streaming()
            .process(exchange -> {
                String tcId = exchange.getIn().getBody(Map.class).get("TC_ID").toString();
                Thread.startVirtualThread(() -> {
                    processTestCase(tcId);
                });
            });
            
        from("direct:error")
            .process(exchange -> {
                Exception cause = exchange.getProperty(Exchange.EXCEPTION_CAUGHT, Exception.class);
                log.error("Processing failed", cause);
            })
            .to("log:error");
    }
}
