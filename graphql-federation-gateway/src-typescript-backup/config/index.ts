/**
 * Revolutionary ERP GraphQL Federation Gateway Configuration
 * Enterprise-grade configuration management with environment-specific settings
 */

import { readFileSync } from 'fs';
import { join } from 'path';

export interface GatewayConfig {
  // Server Configuration
  server: {
    port: number;
    host: string;
    cors: {
      origin: string[] | boolean;
      credentials: boolean;
    };
    helmet: {
      contentSecurityPolicy: boolean;
      crossOriginEmbedderPolicy: boolean;
    };
  };

  // Apollo Federation Configuration
  federation: {
    supergraphSdl?: string;
    introspectionEnabled: boolean;
    playgroundEnabled: boolean;
    schemaRegistryUrl?: string;
    apolloKey?: string;
    apolloGraphRef?: string;
  };

  // Security Configuration (Zero-Trust Integration)
  security: {
    jwt: {
      secret: string;
      issuer: string;
      audience: string;
      algorithms: string[];
    };
    mtls: {
      enabled: boolean;
      clientCertPath?: string;
      clientKeyPath?: string;
      caPath?: string;
      rejectUnauthorized: boolean;
    };
    spiffe: {
      enabled: boolean;
      socketPath?: string;
      workloadApiAddress?: string;
    };
    rateLimiting: {
      windowMs: number;
      maxRequests: number;
      skipSuccessfulRequests: boolean;
    };
  };

  // Caching Configuration
  cache: {
    redis: {
      host: string;
      port: number;
      password?: string;
      db: number;
      keyPrefix: string;
      ttl: number;
    };
    l1Cache: {
      enabled: boolean;
      maxSize: number;
      ttl: number;
    };
  };

  // Subscription Configuration
  subscriptions: {
    kafka: {
      brokers: string[];
      clientId: string;
      groupId: string;
      ssl: boolean;
      sasl?: {
        mechanism: string;
        username: string;
        password: string;
      };
    };
    redis: {
      host: string;
      port: number;
      password?: string;
    };
    websocket: {
      path: string;
      keepAlive: number;
      connectionInitWaitTimeout: number;
    };
  };

  // Monitoring Configuration
  monitoring: {
    prometheus: {
      enabled: boolean;
      port: number;
      path: string;
    };
    tracing: {
      enabled: boolean;
      jaegerEndpoint?: string;
      serviceName: string;
      sampleRate: number;
    };
    logging: {
      level: string;
      format: string;
      destination: string;
    };
  };

  // Query Optimization Configuration
  optimization: {
    queryComplexityLimit: number;
    queryDepthLimit: number;
    queryTimeout: number;
    dataloaderEnabled: boolean;
    gpu: {
      enabled: boolean;
      resourceManagerUrl?: string;
      maxConcurrentOperations: number;
    };
  };

  // Subgraph Configuration
  subgraphs: Record<string, SubgraphConfig>;
}

export interface SubgraphConfig {
  name: string;
  url: string;
  retries: number;
  timeout: number;
  healthCheckPath?: string;
  headers?: Record<string, string>;
}

// Environment-specific configuration loading
class ConfigurationManager {
  private config: GatewayConfig;

  constructor() {
    this.config = this.loadConfiguration();
  }

  public getConfig(): GatewayConfig {
    return this.config;
  }

  private loadConfiguration(): GatewayConfig {
    const env = process.env.NODE_ENV || 'development';
    
    const baseConfig: GatewayConfig = {
      server: {
        port: parseInt(process.env.PORT || '4000', 10),
        host: process.env.HOST || '0.0.0.0',
        cors: {
          origin: process.env.CORS_ORIGIN?.split(',') || ['http://localhost:3000'],
          credentials: process.env.CORS_CREDENTIALS === 'true',
        },
        helmet: {
          contentSecurityPolicy: process.env.CSP_ENABLED !== 'false',
          crossOriginEmbedderPolicy: process.env.COEP_ENABLED !== 'false',
        },
      },

      federation: {
        introspectionEnabled: process.env.INTROSPECTION_ENABLED === 'true',
        playgroundEnabled: process.env.PLAYGROUND_ENABLED === 'true',
        schemaRegistryUrl: process.env.APOLLO_SCHEMA_REGISTRY_URL,
        apolloKey: process.env.APOLLO_KEY,
        apolloGraphRef: process.env.APOLLO_GRAPH_REF,
      },

      security: {
        jwt: {
          secret: process.env.JWT_SECRET || 'your-super-secret-jwt-key',
          issuer: process.env.JWT_ISSUER || 'erp-federation-gateway',
          audience: process.env.JWT_AUDIENCE || 'erp-microservices',
          algorithms: (process.env.JWT_ALGORITHMS || 'RS256').split(','),
        },
        mtls: {
          enabled: process.env.MTLS_ENABLED === 'true',
          clientCertPath: process.env.MTLS_CLIENT_CERT_PATH,
          clientKeyPath: process.env.MTLS_CLIENT_KEY_PATH,
          caPath: process.env.MTLS_CA_PATH,
          rejectUnauthorized: process.env.MTLS_REJECT_UNAUTHORIZED !== 'false',
        },
        spiffe: {
          enabled: process.env.SPIFFE_ENABLED === 'true',
          socketPath: process.env.SPIFFE_SOCKET_PATH,
          workloadApiAddress: process.env.SPIFFE_WORKLOAD_API_ADDRESS,
        },
        rateLimiting: {
          windowMs: parseInt(process.env.RATE_LIMIT_WINDOW_MS || '60000', 10),
          maxRequests: parseInt(process.env.RATE_LIMIT_MAX_REQUESTS || '100', 10),
          skipSuccessfulRequests: process.env.RATE_LIMIT_SKIP_SUCCESS === 'true',
        },
      },

      cache: {
        redis: {
          host: process.env.REDIS_HOST || 'localhost',
          port: parseInt(process.env.REDIS_PORT || '6379', 10),
          password: process.env.REDIS_PASSWORD,
          db: parseInt(process.env.REDIS_DB || '0', 10),
          keyPrefix: process.env.REDIS_KEY_PREFIX || 'erp-gql:',
          ttl: parseInt(process.env.REDIS_TTL || '300', 10),
        },
        l1Cache: {
          enabled: process.env.L1_CACHE_ENABLED !== 'false',
          maxSize: parseInt(process.env.L1_CACHE_MAX_SIZE || '1000', 10),
          ttl: parseInt(process.env.L1_CACHE_TTL || '60', 10),
        },
      },

      subscriptions: {
        kafka: {
          brokers: (process.env.KAFKA_BROKERS || 'localhost:9092').split(','),
          clientId: process.env.KAFKA_CLIENT_ID || 'erp-federation-gateway',
          groupId: process.env.KAFKA_GROUP_ID || 'graphql-subscriptions',
          ssl: process.env.KAFKA_SSL === 'true',
          sasl: process.env.KAFKA_SASL_USERNAME ? {
            mechanism: process.env.KAFKA_SASL_MECHANISM || 'PLAIN',
            username: process.env.KAFKA_SASL_USERNAME,
            password: process.env.KAFKA_SASL_PASSWORD || '',
          } : undefined,
        },
        redis: {
          host: process.env.SUBSCRIPTION_REDIS_HOST || process.env.REDIS_HOST || 'localhost',
          port: parseInt(process.env.SUBSCRIPTION_REDIS_PORT || process.env.REDIS_PORT || '6379', 10),
          password: process.env.SUBSCRIPTION_REDIS_PASSWORD || process.env.REDIS_PASSWORD,
        },
        websocket: {
          path: process.env.WEBSOCKET_PATH || '/graphql',
          keepAlive: parseInt(process.env.WEBSOCKET_KEEP_ALIVE || '30000', 10),
          connectionInitWaitTimeout: parseInt(process.env.WEBSOCKET_INIT_TIMEOUT || '10000', 10),
        },
      },

      monitoring: {
        prometheus: {
          enabled: process.env.PROMETHEUS_ENABLED !== 'false',
          port: parseInt(process.env.PROMETHEUS_PORT || '9090', 10),
          path: process.env.PROMETHEUS_PATH || '/metrics',
        },
        tracing: {
          enabled: process.env.TRACING_ENABLED === 'true',
          jaegerEndpoint: process.env.JAEGER_ENDPOINT,
          serviceName: process.env.SERVICE_NAME || 'erp-federation-gateway',
          sampleRate: parseFloat(process.env.TRACING_SAMPLE_RATE || '0.1'),
        },
        logging: {
          level: process.env.LOG_LEVEL || 'info',
          format: process.env.LOG_FORMAT || 'json',
          destination: process.env.LOG_DESTINATION || 'console',
        },
      },

      optimization: {
        queryComplexityLimit: parseInt(process.env.QUERY_COMPLEXITY_LIMIT || '1000', 10),
        queryDepthLimit: parseInt(process.env.QUERY_DEPTH_LIMIT || '10', 10),
        queryTimeout: parseInt(process.env.QUERY_TIMEOUT || '30000', 10),
        dataloaderEnabled: process.env.DATALOADER_ENABLED !== 'false',
        gpu: {
          enabled: process.env.GPU_ENABLED === 'true',
          resourceManagerUrl: process.env.GPU_RESOURCE_MANAGER_URL,
          maxConcurrentOperations: parseInt(process.env.GPU_MAX_CONCURRENT || '5', 10),
        },
      },

      subgraphs: this.loadSubgraphConfigs(),
    };

    // Load environment-specific overrides
    return this.applyEnvironmentOverrides(baseConfig, env);
  }

  private loadSubgraphConfigs(): Record<string, SubgraphConfig> {
    return {
      'people-and-organizations': {
        name: 'people-and-organizations',
        url: process.env.PEOPLE_ORGS_SERVICE_URL || 'http://localhost:8081/graphql',
        retries: parseInt(process.env.PEOPLE_ORGS_RETRIES || '3', 10),
        timeout: parseInt(process.env.PEOPLE_ORGS_TIMEOUT || '5000', 10),
        healthCheckPath: '/health',
      },
      'e-commerce': {
        name: 'e-commerce',
        url: process.env.ECOMMERCE_SERVICE_URL || 'http://localhost:8082/graphql',
        retries: parseInt(process.env.ECOMMERCE_RETRIES || '3', 10),
        timeout: parseInt(process.env.ECOMMERCE_TIMEOUT || '5000', 10),
        healthCheckPath: '/health',
      },
      'accounting-and-budgeting': {
        name: 'accounting-and-budgeting',
        url: process.env.ACCOUNTING_SERVICE_URL || 'http://localhost:8083/graphql',
        retries: parseInt(process.env.ACCOUNTING_RETRIES || '3', 10),
        timeout: parseInt(process.env.ACCOUNTING_TIMEOUT || '5000', 10),
        healthCheckPath: '/health',
      },
      'products': {
        name: 'products',
        url: process.env.PRODUCTS_SERVICE_URL || 'http://localhost:8084/graphql',
        retries: parseInt(process.env.PRODUCTS_RETRIES || '3', 10),
        timeout: parseInt(process.env.PRODUCTS_TIMEOUT || '5000', 10),
        healthCheckPath: '/health',
      },
      'orders': {
        name: 'orders',
        url: process.env.ORDERS_SERVICE_URL || 'http://localhost:8085/graphql',
        retries: parseInt(process.env.ORDERS_RETRIES || '3', 10),
        timeout: parseInt(process.env.ORDERS_TIMEOUT || '5000', 10),
        healthCheckPath: '/health',
      },
      'invoices': {
        name: 'invoices',
        url: process.env.INVOICES_SERVICE_URL || 'http://localhost:8086/graphql',
        retries: parseInt(process.env.INVOICES_RETRIES || '3', 10),
        timeout: parseInt(process.env.INVOICES_TIMEOUT || '5000', 10),
        healthCheckPath: '/health',
      },
      'shipments': {
        name: 'shipments',
        url: process.env.SHIPMENTS_SERVICE_URL || 'http://localhost:8087/graphql',
        retries: parseInt(process.env.SHIPMENTS_RETRIES || '3', 10),
        timeout: parseInt(process.env.SHIPMENTS_TIMEOUT || '5000', 10),
        healthCheckPath: '/health',
      },
      'human-resources': {
        name: 'human-resources',
        url: process.env.HR_SERVICE_URL || 'http://localhost:8088/graphql',
        retries: parseInt(process.env.HR_RETRIES || '3', 10),
        timeout: parseInt(process.env.HR_TIMEOUT || '5000', 10),
        healthCheckPath: '/health',
      },
      'work-effort': {
        name: 'work-effort',
        url: process.env.WORK_EFFORT_SERVICE_URL || 'http://localhost:8089/graphql',
        retries: parseInt(process.env.WORK_EFFORT_RETRIES || '3', 10),
        timeout: parseInt(process.env.WORK_EFFORT_TIMEOUT || '5000', 10),
        healthCheckPath: '/health',
      },
      'analytics': {
        name: 'analytics',
        url: process.env.ANALYTICS_SERVICE_URL || 'http://localhost:8090/graphql',
        retries: parseInt(process.env.ANALYTICS_RETRIES || '3', 10),
        timeout: parseInt(process.env.ANALYTICS_TIMEOUT || '5000', 10),
        healthCheckPath: '/health',
      },
    };
  }

  private applyEnvironmentOverrides(config: GatewayConfig, env: string): GatewayConfig {
    switch (env) {
      case 'production':
        return {
          ...config,
          federation: {
            ...config.federation,
            introspectionEnabled: false,
            playgroundEnabled: false,
          },
          monitoring: {
            ...config.monitoring,
            tracing: {
              ...config.monitoring.tracing,
              enabled: true,
              sampleRate: 0.05, // Lower sample rate in production
            },
          },
          optimization: {
            ...config.optimization,
            queryComplexityLimit: 800, // Stricter in production
            queryDepthLimit: 8,
          },
        };

      case 'staging':
        return {
          ...config,
          federation: {
            ...config.federation,
            introspectionEnabled: true,
            playgroundEnabled: true,
          },
          monitoring: {
            ...config.monitoring,
            tracing: {
              ...config.monitoring.tracing,
              enabled: true,
              sampleRate: 0.1,
            },
          },
        };

      case 'development':
      default:
        return {
          ...config,
          federation: {
            ...config.federation,
            introspectionEnabled: true,
            playgroundEnabled: true,
          },
          security: {
            ...config.security,
            mtls: {
              ...config.security.mtls,
              enabled: false, // Disable mTLS in development
            },
          },
        };
    }
  }

  // Hot reload configuration (for development)
  public reloadConfiguration(): void {
    this.config = this.loadConfiguration();
  }
}

// Singleton instance
const configManager = new ConfigurationManager();
export const config = configManager.getConfig();
export const reloadConfig = () => configManager.reloadConfiguration();

// Configuration validation
export function validateConfiguration(cfg: GatewayConfig): void {
  const requiredEnvVars = [
    'JWT_SECRET',
  ];

  const missing = requiredEnvVars.filter(envVar => !process.env[envVar]);
  if (missing.length > 0) {
    throw new Error(`Missing required environment variables: ${missing.join(', ')}`);
  }

  // Additional validation logic
  if (cfg.security.mtls.enabled) {
    if (!cfg.security.mtls.clientCertPath || !cfg.security.mtls.clientKeyPath) {
      throw new Error('mTLS is enabled but certificate paths are not configured');
    }
  }

  if (cfg.monitoring.tracing.enabled && !cfg.monitoring.tracing.jaegerEndpoint) {
    throw new Error('Tracing is enabled but Jaeger endpoint is not configured');
  }
}

export default config;