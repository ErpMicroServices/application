/**
 * Revolutionary ERP GraphQL Federation Gateway Type Definitions
 * Enterprise-grade type definitions for the federation gateway
 */

import { Request, Response } from 'express';
import { User } from 'jsonwebtoken';

// Extended Request interface for GraphQL context
export interface GraphQLRequest extends Request {
  user?: AuthenticatedUser;
  correlationId?: string;
  traceId?: string;
  spanId?: string;
}

// Enhanced User interface with ERP-specific data
export interface AuthenticatedUser extends User {
  id: string;
  username: string;
  email?: string;
  roles: string[];
  permissions: string[];
  organizationId?: string;
  departmentId?: string;
  employeeId?: string;
  customerId?: string;
  supplierId?: string;
  territory?: {
    id: string;
    name: string;
    regions: string[];
  };
  preferences: {
    language: string;
    timezone: string;
    currency: string;
  };
  session: {
    sessionId: string;
    loginTime: Date;
    lastActivity: Date;
    ipAddress: string;
    userAgent: string;
  };
  spiffeId?: string;
  priority: 'low' | 'normal' | 'high' | 'critical';
  tier: 'basic' | 'professional' | 'enterprise';
}

// GraphQL Context with enterprise features
export interface GraphQLContext {
  req: GraphQLRequest;
  res: Response;
  user?: AuthenticatedUser;
  correlationId: string;
  traceId: string;
  connectionId?: string; // For subscriptions
  dataloaders: DataloaderCollection;
  cache: CacheManager;
  metrics: MetricsCollector;
  gpu?: GPUResource;
  startTime: number;
  operationType?: 'query' | 'mutation' | 'subscription';
  operationName?: string;
  span?: any; // OpenTelemetry span
}

// Dataloader collection for N+1 optimization
export interface DataloaderCollection {
  [key: string]: any; // DataLoader instances will be added dynamically
}

// Cache management interface
export interface CacheManager {
  get<T>(key: string): Promise<T | null>;
  set<T>(key: string, value: T, options?: CacheOptions): Promise<void>;
  delete(key: string): Promise<void>;
  invalidateByTags(tags: string[]): Promise<void>;
  stats(): CacheStats;
}

export interface CacheOptions {
  ttl?: number; // Time to live in seconds
  tags?: string[]; // Tags for intelligent invalidation
  compress?: boolean;
  serializationStrategy?: 'json' | 'msgpack' | 'protobuf';
}

export interface CacheStats {
  hits: number;
  misses: number;
  hitRate: number;
  size: number;
  memoryUsage: number;
}

// Metrics collection interface
export interface MetricsCollector {
  incrementCounter(name: string, labels?: Record<string, string>, value?: number): void;
  recordHistogram(name: string, value: number, labels?: Record<string, string>): void;
  setGauge(name: string, value: number, labels?: Record<string, string>): void;
  startTimer(name: string, labels?: Record<string, string>): () => void;
}

// GPU Resource management
export interface GPUResource {
  id: string;
  allocated: boolean;
  memoryAllocated: number;
  estimatedDuration: number;
  priority: 'low' | 'normal' | 'high';
}

// Query complexity analysis
export interface QueryComplexity {
  total: number;
  requiresGPU: boolean;
  estimatedDuration: number; // milliseconds
  memoryRequirement: number; // MB
  fields: FieldComplexity[];
}

export interface FieldComplexity {
  fieldName: string;
  points: number;
  requiresGPU: boolean;
  estimatedDuration: number;
  memoryRequirement: number;
  subgraph: string;
}

// Subscription management
export interface SubscriptionMetadata {
  queryHash: string;
  subscribers: Set<string>;
  resourceRequirements: ResourceRequirements;
  securityContext: AuthenticatedUser;
  createdAt: Date;
  lastActivity: Date;
}

export interface ResourceRequirements {
  requiresGPU: boolean;
  estimatedMemory: number;
  estimatedCPU: number;
  networkBandwidth: number;
}

// Event sourcing integration
export interface DomainEvent {
  id: string;
  type: string;
  aggregateId: string;
  aggregateType: string;
  version: number;
  timestamp: Date;
  data: any;
  metadata: EventMetadata;
}

export interface EventMetadata {
  correlationId: string;
  causationId?: string;
  userId: string;
  source: string;
  traceId: string;
}

// Command execution result
export interface CommandResult {
  success: boolean;
  aggregateId: string;
  version: number;
  events: DomainEvent[];
  errors?: string[];
}

// Business entity interfaces for cross-domain relationships
export interface BusinessEntity {
  id: string;
  type: string;
  version: number;
  createdAt: Date;
  updatedAt: Date;
  createdBy: string;
  updatedBy: string;
}

// Party (Person/Organization) - Core entity across domains
export interface Party extends BusinessEntity {
  partyType: 'PERSON' | 'ORGANIZATION';
  names: PartyName[];
  contactMechanisms: ContactMechanism[];
  relationships: PartyRelationship[];
  roles: PartyRole[];
  classifications: PartyClassification[];
}

export interface PartyName {
  id: string;
  name: string;
  nameType: string;
  fromDate: Date;
  thruDate?: Date;
}

export interface ContactMechanism {
  id: string;
  type: string;
  value: string;
  fromDate: Date;
  thruDate?: Date;
}

export interface PartyRelationship {
  id: string;
  fromPartyId: string;
  toPartyId: string;
  relationshipType: string;
  fromDate: Date;
  thruDate?: Date;
}

export interface PartyRole {
  id: string;
  partyId: string;
  roleType: string;
  fromDate: Date;
  thruDate?: Date;
}

export interface PartyClassification {
  id: string;
  partyId: string;
  classificationType: string;
  value: string;
  fromDate: Date;
  thruDate?: Date;
}

// Product - Core entity for commerce domains
export interface Product extends BusinessEntity {
  sku: string;
  name: string;
  description?: string;
  categoryId: string;
  brandId?: string;
  status: ProductStatus;
  pricing: ProductPricing[];
  inventory: InventoryInfo[];
  attributes: ProductAttribute[];
}

export interface ProductPricing {
  id: string;
  priceType: string;
  amount: number;
  currency: string;
  fromDate: Date;
  thruDate?: Date;
}

export interface InventoryInfo {
  warehouseId: string;
  quantityOnHand: number;
  quantityReserved: number;
  quantityAvailable: number;
  lastUpdated: Date;
}

export interface ProductAttribute {
  attributeId: string;
  value: string;
  fromDate: Date;
  thruDate?: Date;
}

export type ProductStatus = 'ACTIVE' | 'INACTIVE' | 'DISCONTINUED' | 'DRAFT';

// Order - Core entity for order management
export interface Order extends BusinessEntity {
  orderNumber: string;
  customerId: string;
  orderDate: Date;
  requestedDeliveryDate?: Date;
  status: OrderStatus;
  currency: string;
  totalAmount: number;
  lineItems: OrderLineItem[];
  addresses: OrderAddress[];
  payments: OrderPayment[];
}

export interface OrderLineItem {
  id: string;
  productId: string;
  quantity: number;
  unitPrice: number;
  totalPrice: number;
  status: OrderLineItemStatus;
}

export interface OrderAddress {
  type: 'BILLING' | 'SHIPPING';
  contactMechanismId: string;
}

export interface OrderPayment {
  id: string;
  paymentMethodId: string;
  amount: number;
  status: PaymentStatus;
}

export type OrderStatus = 'DRAFT' | 'SUBMITTED' | 'APPROVED' | 'IN_PROGRESS' | 'COMPLETED' | 'CANCELLED';
export type OrderLineItemStatus = 'PENDING' | 'ALLOCATED' | 'SHIPPED' | 'DELIVERED' | 'CANCELLED';
export type PaymentStatus = 'PENDING' | 'AUTHORIZED' | 'CAPTURED' | 'FAILED' | 'REFUNDED';

// Analytics and reporting types
export interface AnalyticsQuery {
  dimensions: string[];
  metrics: string[];
  filters: AnalyticsFilter[];
  timeRange: TimeRange;
  granularity: TimeGranularity;
}

export interface AnalyticsFilter {
  field: string;
  operator: 'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte' | 'in' | 'nin';
  value: any;
}

export interface TimeRange {
  start: Date;
  end: Date;
}

export type TimeGranularity = 'hour' | 'day' | 'week' | 'month' | 'quarter' | 'year';

// Error handling types
export interface BusinessLogicException extends Error {
  code: string;
  context?: any;
  severity: 'low' | 'medium' | 'high' | 'critical';
}

export interface ValidationException extends Error {
  field: string;
  value: any;
  constraints: string[];
}

// Health check types
export interface HealthCheckResult {
  status: 'healthy' | 'unhealthy' | 'degraded';
  timestamp: Date;
  duration: number;
  checks: ComponentHealthCheck[];
}

export interface ComponentHealthCheck {
  component: string;
  status: 'healthy' | 'unhealthy' | 'degraded';
  message?: string;
  timestamp: Date;
  duration: number;
  metadata?: any;
}

// Configuration types (re-exported for convenience)
export * from '../config';

// Utility types
export type Maybe<T> = T | null | undefined;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };

// Generic response wrapper
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  errors?: ApiError[];
  metadata?: ResponseMetadata;
}

export interface ApiError {
  code: string;
  message: string;
  field?: string;
  context?: any;
}

export interface ResponseMetadata {
  requestId: string;
  timestamp: Date;
  duration: number;
  version: string;
}

// Pagination types
export interface PageInfo {
  hasNextPage: boolean;
  hasPreviousPage: boolean;
  startCursor?: string;
  endCursor?: string;
}

export interface Connection<T> {
  edges: Edge<T>[];
  pageInfo: PageInfo;
  totalCount?: number;
}

export interface Edge<T> {
  node: T;
  cursor: string;
}

export interface PaginationInput {
  first?: number;
  after?: string;
  last?: number;
  before?: string;
}

// Federation-specific types
export interface EntityReference {
  __typename: string;
  [key: string]: any;
}

export interface FederationContext {
  isReference: boolean;
  representationKeys: string[];
  originalContext?: GraphQLContext;
}

// Real-time subscription types
export interface SubscriptionFilter {
  [field: string]: any;
}

export interface SubscriptionOptions {
  filter?: SubscriptionFilter;
  debounceMs?: number;
  bufferSize?: number;
  authentication?: boolean;
}

export interface SubscriptionEvent<T = any> {
  type: string;
  data: T;
  timestamp: Date;
  source: string;
  metadata?: any;
}