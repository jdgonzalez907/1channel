# Spec Delta

## ADDED Requirements

### Requirement: Nombre del contacto en el evento simulado

El simulador SHALL ofrecer un campo opcional `display_name`, visible y enviado únicamente para `message.received`. El campo SHALL NOT ser obligatorio para enviar el evento.

#### Scenario: Campo visible en recepción

- **WHEN** el agente selecciona `message.received`
- **THEN** el simulador muestra el campo `display_name` como opcional

#### Scenario: Campo oculto en otros eventos

- **WHEN** el agente selecciona `message.edited`, `message.deleted` o `message.read`
- **THEN** el simulador no muestra el campo `display_name`
