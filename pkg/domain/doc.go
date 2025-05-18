// ABOUTME: Domain layer package containing core business entities and logic
// ABOUTME: Provides single source of truth for all domain types used across the application

/*
Package domain contains the core business entities and domain logic for Magellai.

This package serves as the single source of truth for all domain types and is designed
to be independent of infrastructure concerns. It contains the essential business
entities that are shared across multiple packages in the application.

The domain layer follows Domain-Driven Design principles and provides:
  - Core business entities (Session, Message, Conversation, etc.)
  - Value objects (MessageRole, AttachmentType, etc.)
  - Domain logic and validation
  - Clean interfaces for use by other layers

All other packages in the application should depend on these domain types rather
than defining their own versions. This eliminates type duplication and provides
a clear separation between business logic and infrastructure concerns.

Architecture:
  - Domain Layer (this package): Core business entities and logic
  - Application Layer (pkg/repl, pkg/command): Use case orchestration
  - Infrastructure Layer (pkg/storage, pkg/llm): Technical implementations

Key types:
  - Session: Represents a complete chat session with conversation history
  - Message: Individual message within a conversation
  - Conversation: Manages the state of an interactive conversation
  - Attachment: Multimodal content attached to messages
  - Provider/Model: LLM provider and model configurations
*/
package domain
