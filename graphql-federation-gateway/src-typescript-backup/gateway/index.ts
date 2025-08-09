/**
 * Revolutionary ERP GraphQL Federation Gateway
 * Enterprise-grade Apollo Federation v2 implementation
 */

import { ApolloGateway, RemoteGraphQLDataSource } from '@apollo/gateway';
import { ApolloServer } from '@apollo/server';
import { expressMiddleware } from '@apollo/server/express4';
import { ApolloServerPluginDrainHttpServer } from '@apollo/server/plugin/drainHttpServer';
import { ApolloServerPluginLandingPageLocalDefault } from '@apollo/server/plugin/landingPageLocal';
import { buildSubgraphSchema } from '@apollo/subgraph';
import express from 'express';
import { createServer } from 'http';
import { WebSocketServer } from 'ws';
import { useServer } from 'graphql-ws/lib/use/ws';
import cors from 'cors';
import helmet from 'helmet';
import jwt from 'jsonwebtoken';
import https from 'https';
import { readFileSync } from 'fs';

import { config, GatewayConfig, SubgraphConfig } from '../config';
import {
  GraphQLContext,
  GraphQLRequest,
  AuthenticatedUser,
  QueryComplexity,
  SubscriptionMetadata,
} from '../types';
import { SecurityManager } from '../security';
import { CacheManager } from '../cache';
import { MonitoringManager } from '../monitoring';
import { SubscriptionManager } from '../subscriptions';
import { DataloaderManager } from '../dataloaders';
import { QueryComplexityAnalyzer } from '../optimization';
import { Logger } from '../utils/logger';

export class RevolutionaryERPGateway {
  private app: express.Application;
  private httpServer: any;
  private wsServer: WebSocketServer;
  private apolloServer: ApolloServer<GraphQLContext>;
  private gateway: ApolloGateway;
  
  private securityManager: SecurityManager;
  private cacheManager: CacheManager;
  private monitoringManager: MonitoringManager;
  private subscriptionManager: SubscriptionManager;
  private dataloaderManager: DataloaderManager;
  private complexityAnalyzer: QueryComplexityAnalyzer;
  private logger: Logger;

  constructor(private config: GatewayConfig) {
    this.logger = new Logger('RevolutionaryERPGateway');
    this.initializeManagers();
    this.setupExpress();
    this.setupGateway();
    this.setupApolloServer();
  }

  private initializeManagers(): void {
    this.securityManager = new SecurityManager(this.config.security);
    this.cacheManager = new CacheManager(this.config.cache);
    this.monitoringManager = new MonitoringManager(this.config.monitoring);
    this.subscriptionManager = new SubscriptionManager(this.config.subscriptions);
    this.dataloaderManager = new DataloaderManager();
    this.complexityAnalyzer = new QueryComplexityAnalyzer(this.config.optimization);
  }

  private setupExpress(): void {
    this.app = express();
    this.httpServer = createServer(this.app);
    
    // Security middleware
    this.app.use(helmet(this.config.server.helmet));
    
    // CORS configuration
    this.app.use(cors(this.config.server.cors));
    
    // Request correlation and tracing
    this.app.use(this.correlationMiddleware.bind(this));
    
    // Health check endpoint
    this.app.get('/health', this.healthCheck.bind(this));
    this.app.get('/ready', this.readinessCheck.bind(this));
    
    // Metrics endpoint
    if (this.config.monitoring.prometheus.enabled) {
      this.app.get('/metrics', this.monitoringManager.getMetricsHandler());
    }

    // GraphQL-specific middleware will be added after Apollo Server setup
  }

  private setupGateway(): void {
    this.gateway = new ApolloGateway({
      serviceList: this.buildServiceList(),
      buildService: this.buildSubgraphService.bind(this),
      experimental_didResolveQueryPlan: this.didResolveQueryPlan.bind(this),
      experimental_didFailComposition: this.didFailComposition.bind(this),
    });
  }

  private buildServiceList(): Array<{ name: string; url: string }> {
    return Object.values(this.config.subgraphs).map((subgraph: SubgraphConfig) => ({
      name: subgraph.name,
      url: subgraph.url,
    }));
  }

  private buildSubgraphService({ url, name }: { url: string; name: string }) {
    return new EnhancedRemoteGraphQLDataSource({
      url,
      name,
      config: this.config.subgraphs[name],
      securityManager: this.securityManager,
      cacheManager: this.cacheManager,
      monitoringManager: this.monitoringManager,
      logger: this.logger,
    });
  }

  private async setupApolloServer(): Promise<void> {
    // Setup WebSocket server for subscriptions
    this.wsServer = new WebSocketServer({
      server: this.httpServer,
      path: this.config.subscriptions.websocket.path,
    });

    const { schema } = await this.gateway.load();

    // Create subscription server
    const subscriptionServer = useServer(
      {
        schema,
        context: async (ctx) => {
          return this.createSubscriptionContext(ctx);
        },
        onConnect: this.onSubscriptionConnect.bind(this),
        onDisconnect: this.onSubscriptionDisconnect.bind(this),
        onSubscribe: this.onSubscriptionSubscribe.bind(this),
      },
      this.wsServer
    );

    this.apolloServer = new ApolloServer<GraphQLContext>({
      gateway: this.gateway,
      plugins: [
        // Drain HTTP server on shutdown
        ApolloServerPluginDrainHttpServer({ httpServer: this.httpServer }),
        
        // Drain WebSocket server on shutdown
        {
          async serverWillStart() {
            return {
              async drainServer() {
                await subscriptionServer.dispose();
              },
            };
          },
        },

        // Landing page
        this.config.federation.playgroundEnabled
          ? ApolloServerPluginLandingPageLocalDefault({ footer: false })
          : undefined,

        // Custom plugins
        this.createMonitoringPlugin(),
        this.createComplexityPlugin(),
        this.createCachingPlugin(),
        this.createSecurityPlugin(),
      ].filter(Boolean),

      introspection: this.config.federation.introspectionEnabled,
      csrfPrevention: true,
      cache: 'bounded',
      
      // Enhanced error formatting
      formatError: this.formatError.bind(this),
    });

    await this.apolloServer.start();
  }

  private async correlationMiddleware(
    req: GraphQLRequest,
    res: express.Response,
    next: express.NextFunction
  ): Promise<void> {
    // Generate correlation ID if not present
    req.correlationId = req.headers['x-correlation-id'] as string ||
                       this.generateCorrelationId();
    
    // Set response headers
    res.set('X-Correlation-ID', req.correlationId);
    res.set('X-Service', 'erp-federation-gateway');
    
    next();
  }

  private async createGraphQLContext({
    req,
    res,
  }: {
    req: GraphQLRequest;
    res: express.Response;
  }): Promise<GraphQLContext> {
    const startTime = Date.now();
    
    // Extract user from JWT token
    const user = await this.extractUserFromRequest(req);
    
    // Create tracing span
    const span = this.monitoringManager.startSpan('graphql.request', {
      'user.id': user?.id,
      'correlation.id': req.correlationId,
    });

    // Create context with all enterprise features
    const context: GraphQLContext = {
      req,
      res,
      user,
      correlationId: req.correlationId || this.generateCorrelationId(),
      traceId: span?.traceId || this.generateTraceId(),
      dataloaders: this.dataloaderManager.createLoaders(user),
      cache: this.cacheManager,
      metrics: this.monitoringManager.getMetricsCollector(),
      startTime,
      span,
    };

    return context;
  }

  private async createSubscriptionContext(ctx: any): Promise<GraphQLContext> {
    const connectionParams = ctx.connectionParams || {};
    const token = connectionParams.authorization || connectionParams.Authorization;
    
    const user = await this.validateSubscriptionAuth(token);
    
    return {
      req: null as any, // Not applicable for subscriptions
      res: null as any, // Not applicable for subscriptions
      user,
      correlationId: this.generateCorrelationId(),
      traceId: this.generateTraceId(),
      connectionId: ctx.connectionId,
      dataloaders: this.dataloaderManager.createLoaders(user),
      cache: this.cacheManager,
      metrics: this.monitoringManager.getMetricsCollector(),
      startTime: Date.now(),
    };
  }

  private async extractUserFromRequest(req: GraphQLRequest): Promise<AuthenticatedUser | undefined> {
    const authHeader = req.headers.authorization;
    if (!authHeader) return undefined;

    const token = authHeader.replace('Bearer ', '');
    return this.securityManager.validateJWTToken(token);
  }

  private async validateSubscriptionAuth(token?: string): Promise<AuthenticatedUser | undefined> {
    if (!token) return undefined;
    
    const cleanToken = token.replace('Bearer ', '');
    return this.securityManager.validateJWTToken(cleanToken);
  }

  private createMonitoringPlugin() {
    return this.monitoringManager.createApolloPlugin();
  }

  private createComplexityPlugin() {
    return this.complexityAnalyzer.createApolloPlugin();
  }

  private createCachingPlugin() {
    return this.cacheManager.createApolloPlugin();
  }

  private createSecurityPlugin() {
    return this.securityManager.createApolloPlugin();
  }

  private formatError(formattedError: any, error: any): any {
    // Enhanced error formatting with security considerations
    const isDevelopment = process.env.NODE_ENV === 'development';
    
    this.logger.error('GraphQL Error:', {
      error: formattedError.message,
      path: formattedError.path,
      locations: formattedError.locations,
      extensions: formattedError.extensions,
      originalError: isDevelopment ? error.stack : undefined,
    });

    // Record error metrics
    this.monitoringManager.recordError(formattedError, error);

    // Return formatted error (remove sensitive data in production)
    return {
      ...formattedError,
      extensions: {
        ...formattedError.extensions,
        // Remove internal error details in production
        ...(isDevelopment ? { stack: error.stack } : {}),
      },
    };
  }

  private async didResolveQueryPlan(options: any): Promise<void> {
    // Query plan analysis and optimization
    this.logger.debug('Query plan resolved:', {
      operation: options.request.operationName,
      queryHash: options.request.queryHash,
      planNodes: options.queryPlan?.node ? this.countQueryPlanNodes(options.queryPlan.node) : 0,
    });

    // Record query plan metrics
    this.monitoringManager.recordQueryPlan(options);
  }

  private didFailComposition(errors: any[]): void {
    this.logger.error('Schema composition failed:', { errors });
    this.monitoringManager.recordCompositionFailure(errors);
    
    // Could implement notification system here
    throw new Error(`Schema composition failed: ${errors.map(e => e.message).join(', ')}`);
  }

  private countQueryPlanNodes(node: any): number {
    if (!node) return 0;
    
    let count = 1;
    if (node.nodes) {
      count += node.nodes.reduce((sum: number, child: any) => sum + this.countQueryPlanNodes(child), 0);
    }
    
    return count;
  }

  private async onSubscriptionConnect(ctx: any): Promise<boolean> {
    try {
      const user = await this.validateSubscriptionAuth(ctx.connectionParams?.authorization);
      if (!user) {
        this.logger.warn('Subscription connection rejected: No valid authentication');
        return false;
      }

      this.logger.info('Subscription connection established', {
        userId: user.id,
        connectionId: ctx.connectionId,
      });

      return true;
    } catch (error) {
      this.logger.error('Subscription connection error:', error);
      return false;
    }
  }

  private async onSubscriptionDisconnect(ctx: any): Promise<void> {
    this.logger.info('Subscription connection closed', {
      connectionId: ctx.connectionId,
    });

    // Cleanup resources
    await this.subscriptionManager.cleanupConnection(ctx.connectionId);
  }

  private async onSubscriptionSubscribe(ctx: any, message: any): Promise<boolean> {
    try {
      // Validate subscription complexity
      const complexity = await this.complexityAnalyzer.analyzeSubscription(
        message.payload.query,
        message.payload.variables
      );

      if (complexity.total > this.config.optimization.queryComplexityLimit) {
        this.logger.warn('Subscription rejected: Complexity limit exceeded', {
          complexity: complexity.total,
          limit: this.config.optimization.queryComplexityLimit,
        });
        return false;
      }

      return true;
    } catch (error) {
      this.logger.error('Subscription subscribe error:', error);
      return false;
    }
  }

  private async healthCheck(req: express.Request, res: express.Response): Promise<void> {
    try {
      const health = await this.performHealthCheck();
      res.status(health.status === 'healthy' ? 200 : 503).json(health);
    } catch (error) {
      res.status(503).json({
        status: 'unhealthy',
        timestamp: new Date(),
        error: 'Health check failed',
      });
    }
  }

  private async readinessCheck(req: express.Request, res: express.Response): Promise<void> {
    try {
      const isReady = await this.isReady();
      res.status(isReady ? 200 : 503).json({
        status: isReady ? 'ready' : 'not ready',
        timestamp: new Date(),
      });
    } catch (error) {
      res.status(503).json({
        status: 'not ready',
        timestamp: new Date(),
        error: 'Readiness check failed',
      });
    }
  }

  private async performHealthCheck(): Promise<any> {
    // Implement comprehensive health check
    const checks = await Promise.allSettled([
      this.checkSubgraphHealth(),
      this.cacheManager.healthCheck(),
      this.subscriptionManager.healthCheck(),
      this.monitoringManager.healthCheck(),
    ]);

    const results = checks.map((check, index) => ({
      component: ['subgraphs', 'cache', 'subscriptions', 'monitoring'][index],
      status: check.status === 'fulfilled' ? 'healthy' : 'unhealthy',
      error: check.status === 'rejected' ? check.reason.message : undefined,
    }));

    const overallStatus = results.every(r => r.status === 'healthy') ? 'healthy' : 'unhealthy';

    return {
      status: overallStatus,
      timestamp: new Date(),
      checks: results,
    };
  }

  private async checkSubgraphHealth(): Promise<boolean> {
    // Check if all subgraphs are accessible
    const healthChecks = Object.values(this.config.subgraphs).map(async (subgraph) => {
      try {
        // Implement actual health check call to subgraph
        return true;
      } catch {
        return false;
      }
    });

    const results = await Promise.all(healthChecks);
    return results.every(Boolean);
  }

  private async isReady(): Promise<boolean> {
    // Check if gateway is ready to serve requests
    return this.apolloServer !== undefined && this.gateway !== undefined;
  }

  private generateCorrelationId(): string {
    return `gw-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  private generateTraceId(): string {
    return `trace-${Date.now()}-${Math.random().toString(36).substr(2, 16)}`;
  }

  public async start(): Promise<void> {
    try {
      // Setup GraphQL middleware
      this.app.use(
        '/graphql',
        express.json({ limit: '10mb' }),
        expressMiddleware(this.apolloServer, {
          context: this.createGraphQLContext.bind(this),
        })
      );

      // Start the HTTP server
      const port = this.config.server.port;
      const host = this.config.server.host;
      
      await new Promise<void>((resolve) => {
        this.httpServer.listen(port, host, () => {
          this.logger.info(`🚀 Revolutionary ERP GraphQL Federation Gateway ready!`, {
            graphqlUrl: `http://${host}:${port}/graphql`,
            subscriptionsUrl: `ws://${host}:${port}${this.config.subscriptions.websocket.path}`,
            healthUrl: `http://${host}:${port}/health`,
            metricsUrl: this.config.monitoring.prometheus.enabled 
              ? `http://${host}:${port}/metrics` 
              : 'disabled',
          });
          resolve();
        });
      });

      // Start background services
      await this.subscriptionManager.start();
      await this.monitoringManager.start();

      this.logger.info('🎯 All systems operational - Think Different ERP Gateway is live!');

    } catch (error) {
      this.logger.error('Failed to start gateway:', error);
      throw error;
    }
  }

  public async stop(): Promise<void> {
    try {
      this.logger.info('Shutting down Revolutionary ERP Gateway...');

      // Graceful shutdown sequence
      await this.apolloServer.stop();
      await this.subscriptionManager.stop();
      await this.monitoringManager.stop();
      await this.cacheManager.disconnect();
      
      this.httpServer.close();
      this.wsServer.close();

      this.logger.info('Gateway shutdown complete');
    } catch (error) {
      this.logger.error('Error during shutdown:', error);
      throw error;
    }
  }
}

/**
 * Enhanced Remote GraphQL Data Source with enterprise features
 */
class EnhancedRemoteGraphQLDataSource extends RemoteGraphQLDataSource {
  constructor(private options: {
    url: string;
    name: string;
    config: SubgraphConfig;
    securityManager: SecurityManager;
    cacheManager: CacheManager;
    monitoringManager: MonitoringManager;
    logger: Logger;
  }) {
    super({ url: options.url });
  }

  willSendRequest({ request, context }: { request: any; context: GraphQLContext }) {
    // Add authentication headers
    this.addAuthenticationHeaders(request, context);
    
    // Add tracing headers
    this.addTracingHeaders(request, context);
    
    // Add correlation headers
    request.http.headers.set('X-Correlation-ID', context.correlationId);
    request.http.headers.set('X-Source-Service', 'federation-gateway');
    
    // Add user context (encrypted)
    if (context.user) {
      const encryptedUserContext = this.options.securityManager.encryptUserContext(
        context.user,
        this.options.name
      );
      request.http.headers.set('X-User-Context', encryptedUserContext);
    }

    // Setup mTLS if enabled
    if (this.options.securityManager.isMTLSEnabled()) {
      request.http.agent = this.options.securityManager.createMTLSAgent(this.options.name);
    }

    // Record metrics
    this.options.monitoringManager.recordSubgraphRequest(this.options.name, {
      operation: request.query,
      variables: request.variables,
    });
  }

  didReceiveResponse({ response, request, context }: { response: any; request: any; context: GraphQLContext }) {
    // Record response metrics
    this.options.monitoringManager.recordSubgraphResponse(this.options.name, {
      status: response.http.status,
      duration: Date.now() - context.startTime,
    });

    // Handle caching
    if (this.shouldCache(request, response)) {
      this.cacheSubgraphResponse(request, response, context);
    }

    return response;
  }

  didEncounterError(error: any, request: any, context: GraphQLContext) {
    this.options.logger.error(`Subgraph ${this.options.name} error:`, {
      error: error.message,
      request: request.query,
      variables: request.variables,
      correlationId: context.correlationId,
    });

    // Record error metrics
    this.options.monitoringManager.recordSubgraphError(this.options.name, error);

    return error;
  }

  private addAuthenticationHeaders(request: any, context: GraphQLContext): void {
    // Add service-to-service JWT
    const serviceToken = this.options.securityManager.generateServiceToken({
      issuer: 'federation-gateway',
      audience: this.options.name,
      subject: 'gateway-service',
    });
    
    request.http.headers.set('Authorization', `Bearer ${serviceToken}`);
  }

  private addTracingHeaders(request: any, context: GraphQLContext): void {
    if (context.traceId) {
      request.http.headers.set('X-Trace-ID', context.traceId);
    }
    if (context.span?.spanId) {
      request.http.headers.set('X-Span-ID', context.span.spanId);
    }
  }

  private shouldCache(request: any, response: any): boolean {
    // Implement intelligent caching logic
    return (
      response.http.status === 200 &&
      request.query.includes('query') &&
      !request.query.includes('mutation')
    );
  }

  private async cacheSubgraphResponse(request: any, response: any, context: GraphQLContext): Promise<void> {
    const cacheKey = this.generateCacheKey(request);
    await this.options.cacheManager.set(cacheKey, response.data, {
      ttl: 300, // 5 minutes
      tags: [`subgraph:${this.options.name}`],
    });
  }

  private generateCacheKey(request: any): string {
    // Generate deterministic cache key
    const queryHash = this.options.securityManager.hashQuery(request.query);
    const variablesHash = this.options.securityManager.hashVariables(request.variables);
    return `subgraph:${this.options.name}:${queryHash}:${variablesHash}`;
  }
}

export default RevolutionaryERPGateway;