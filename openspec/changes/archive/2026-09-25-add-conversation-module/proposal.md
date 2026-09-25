# Proposal

## Why

1Channel es un CRM conversacional multi-canal. Necesitamos un modulo de conversaciones que permita a los contactos iniciar conversaciones con la empresa para soporte, ventas, cambios, envios, y otras interacciones cortas. Cada conversacion tiene un solo agente activo y puede expirar segun la configuracion de la plataforma.

## What Changes

- Nuevo modulo `conversations` en `internal/module/conversations/`
- Aggregate root `Conversation` con campos para gestionar el ciclo de vida de la conversacion
- Entity `Message` para representar los mensajes dentro de una conversacion
- Value Objects para estados: `ConversationStatus`, `MessageStatus`, `MessageType`
- Constructores globales con validaciones de negocio:
  - UUID no puede ser Nil
  - Conversacion debe tener al menos un contacto o agente
  - Texto de mensaje debe tener entre 1 y 1000 caracteres visuales (grapheme clusters)
- Getters en una linea para todos los campos
- Errores de dominio scoped por archivo con mensajes en lenguaje de negocio
- Constantes para todos los valores (sin magic numbers)
- Cada archivo de test corresponde a su archivo fuente exacto

## Capabilities

### New Capabilities

- `conversation`: Gestion de conversaciones entre contactos y agentes. Incluye creacion de conversaciones por contactos, asignacion de agentes, mensajes, y cierre de conversaciones (resuelta o expirada).

### Modified Capabilities

(ninguna - modulo nuevo)

## Impact

- Nuevo directorio `internal/module/conversations/domain/`
- Dependencias:
  - `uuid` package de Go 1.27 (stdlib)
  - `github.com/rivo/uniseg` para conteo de grapheme clusters
- No afecta codigo existente - es un modulo nuevo
