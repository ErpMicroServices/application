/**
 * Revolutionary ERP GraphQL Federation Gateway
 * Main Entry Point for the Think Different ERP System
 * 
 * This gateway revolutionizes how enterprises interact with ERP systems
 * by providing a unified GraphQL API across all business domains.
 */

import { config, validateConfiguration } from './config';
import { RevolutionaryERPGateway } from './gateway';
import { Logger } from './utils/logger';
import { gracefulShutdown } from './utils/shutdown';

// Setup global error handlers
process.on('uncaughtException', (error: Error) => {
  console.error('💥 Uncaught Exception:', error);
  process.exit(1);
});

process.on('unhandledRejection', (reason: unknown, promise: Promise<unknown>) => {
  console.error('💥 Unhandled Rejection at:', promise, 'reason:', reason);
  process.exit(1);
});

/**
 * Main application bootstrap function
 */
async function bootstrap(): Promise<void> {
  const logger = new Logger('Bootstrap');
  
  try {
    // Display revolutionary startup banner
    displayStartupBanner();
    
    // Validate configuration
    logger.info('🔧 Validating configuration...');
    validateConfiguration(config);
    logger.info('✅ Configuration validated successfully');
    
    // Initialize the Revolutionary ERP Gateway
    logger.info('🚀 Initializing Revolutionary ERP Gateway...');
    const gateway = new RevolutionaryERPGateway(config);
    
    // Setup graceful shutdown
    gracefulShutdown(async () => {
      logger.info('🔄 Initiating graceful shutdown...');
      await gateway.stop();
      logger.info('✅ Gateway shutdown complete');
    });
    
    // Start the gateway
    logger.info('🎯 Starting gateway services...');
    await gateway.start();
    
    // Success message
    logger.info('🌟 Revolutionary ERP Gateway is now operational!');
    logger.info(`📊 Serving unified GraphQL API for ${Object.keys(config.subgraphs).length} business domains`);
    logger.info('🎉 Think Different - Your ERP experience has been revolutionized!');
    
  } catch (error) {
    logger.error('💥 Failed to start Revolutionary ERP Gateway:', error);
    process.exit(1);
  }
}

/**
 * Display the revolutionary startup banner
 */
function displayStartupBanner(): void {
  const banner = `
  ╔══════════════════════════════════════════════════════════════════════════════╗
  ║                                                                              ║
  ║                    🚀 REVOLUTIONARY ERP GATEWAY 🚀                          ║
  ║                                                                              ║
  ║                          Think Different ERP System                         ║
  ║                     Unified GraphQL Federation Gateway                      ║
  ║                                                                              ║
  ║  ┌─────────────────────────────────────────────────────────────────────┐   ║
  ║  │  🌟 Enterprise Features:                                            │   ║
  ║  │   • Apollo Federation v2 with supergraph composition              │   ║
  ║  │   • Real-time subscriptions with Kafka integration                │   ║
  ║  │   • Zero-trust security with mTLS and SPIFFE                      │   ║
  ║  │   • Multi-level caching with intelligent invalidation             │   ║
  ║  │   • GPU-accelerated query optimization                            │   ║
  ║  │   • Comprehensive monitoring and observability                    │   ║
  ║  │   • Field-level authorization with business context               │   ║
  ║  │   • Polyglot persistence integration                              │   ║
  ║  └─────────────────────────────────────────────────────────────────────┘   ║
  ║                                                                              ║
  ║  🎯 Business Domains: People & Orgs • E-commerce • Accounting • Products    ║
  ║     Orders • Invoices • Shipments • HR • Work Effort • Analytics           ║
  ║                                                                              ║
  ║                    "Here's to the crazy ones, the misfits,                  ║
  ║                   the rebels, the troublemakers, the round                  ║
  ║                    pegs in the square holes... Because the                  ║
  ║                     people who are crazy enough to think                    ║
  ║                    they can change the world are the ones                   ║
  ║                              who do." - Apple                               ║
  ║                                                                              ║
  ╚══════════════════════════════════════════════════════════════════════════════╝
  
  Environment: ${process.env.NODE_ENV || 'development'}
  Version: ${process.env.APP_VERSION || '1.0.0'}
  Node.js: ${process.version}
  Platform: ${process.platform}
  Architecture: ${process.arch}
  
  `;
  
  console.log(banner);
}

/**
 * Display environment information
 */
function displayEnvironmentInfo(): void {
  const logger = new Logger('Environment');
  
  logger.info('🔧 Configuration Summary:');
  logger.info(`   Server: ${config.server.host}:${config.server.port}`);
  logger.info(`   Subgraphs: ${Object.keys(config.subgraphs).length} domains`);
  logger.info(`   Security: mTLS=${config.security.mtls.enabled}, SPIFFE=${config.security.spiffe.enabled}`);
  logger.info(`   Caching: Redis=${config.cache.redis.host}:${config.cache.redis.port}, L1=${config.cache.l1Cache.enabled}`);
  logger.info(`   Monitoring: Prometheus=${config.monitoring.prometheus.enabled}, Tracing=${config.monitoring.tracing.enabled}`);
  logger.info(`   Subscriptions: Kafka=${config.subscriptions.kafka.brokers.join(',')}`);
  logger.info(`   GPU: Enabled=${config.optimization.gpu.enabled}`);
  
  // Log domain-specific information
  logger.info('🏢 Business Domains:');
  Object.values(config.subgraphs).forEach(subgraph => {
    logger.info(`   ${subgraph.name}: ${subgraph.url}`);
  });
}

/**
 * Health check for container orchestration
 */
export async function healthCheck(): Promise<{ status: string; timestamp: Date }> {
  try {
    // Implement basic health check logic
    return {
      status: 'healthy',
      timestamp: new Date(),
    };
  } catch (error) {
    return {
      status: 'unhealthy',
      timestamp: new Date(),
    };
  }
}

/**
 * Readiness check for container orchestration
 */
export async function readinessCheck(): Promise<{ status: string; timestamp: Date }> {
  try {
    // Implement readiness check logic
    return {
      status: 'ready',
      timestamp: new Date(),
    };
  } catch (error) {
    return {
      status: 'not ready',
      timestamp: new Date(),
    };
  }
}

/**
 * Development mode utilities
 */
if (process.env.NODE_ENV === 'development') {
  // Enable additional development features
  process.env.DEBUG = process.env.DEBUG || 'erp:*';
  
  // Display additional environment info in development
  displayEnvironmentInfo();
}

// Start the application if this file is run directly
if (require.main === module) {
  bootstrap().catch((error) => {
    console.error('💥 Bootstrap failed:', error);
    process.exit(1);
  });
}

// Export for testing and module usage
export { bootstrap, config };
export default bootstrap;