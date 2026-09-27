# Spec Delta

## REMOVED Requirements

### Requirement: API pública del módulo de agentes

**Reason**: El módulo se renombra a `users`; la identidad del sistema se modela como un usuario compartido y no como un concepto propio de conversaciones.

**Migration**: Usar la capability `users`, que expone `UsersAPI` con la validación de existencia por identificador.

### Requirement: Rehidratar agente

**Reason**: El módulo se renombra a `users` y la rehidratación pasa a describirse como rehidratación de usuario.

**Migration**: Usar la rehidratación de la capability `users`.
