# Spec Delta

## ADDED Requirements

### Requirement: Etiqueta del contacto en la consola

La consola SHALL mostrar la etiqueta de presentación del contacto en cada item de la lista y en el encabezado del detalle, en lugar de mostrar únicamente su identificador externo.

#### Scenario: Item con etiqueta

- **WHEN** una conversación aparece en la lista
- **THEN** su item muestra la etiqueta de presentación del contacto

#### Scenario: Encabezado con etiqueta

- **WHEN** el agente abre una conversación
- **THEN** el encabezado del detalle muestra la etiqueta de presentación del contacto

### Requirement: Disposición de tres paneles

La consola SHALL disponer de tres paneles: la lista de conversaciones, el detalle de la conversación y la persona del contacto. El panel de la persona SHALL mostrarse junto al detalle cuando hay una conversación seleccionada.

#### Scenario: Tres paneles con conversación seleccionada

- **WHEN** el agente selecciona una conversación
- **THEN** la consola muestra la lista, el detalle y el panel de la persona

#### Scenario: Sin conversación seleccionada

- **WHEN** no hay conversación seleccionada
- **THEN** la consola no muestra el panel de la persona

### Requirement: Panel de la persona en el detalle

Al abrir una conversación, la consola SHALL mostrar un tercer panel con la persona asociada al contacto. Cuando el contacto tenga persona, la consola SHALL precargar sus datos. Cuando no la tenga, SHALL mostrar el campo `identification_number` con una acción de buscar. Buscar SHALL consultar `GET /v1/personal-information/{identification_number}`: si la persona existe, la consola SHALL precargar sus datos; si no existe, SHALL mostrar los campos vacíos. La consola SHALL ofrecer una acción de guardar que haga `PUT /v1/contacts/{id}/personal-information` con el `identification_number` y los datos; tras el éxito SHALL reflejar la persona y la etiqueta actualizada. El documento SHALL ser editable aunque el contacto ya tenga persona, de modo que se pueda corregir o reasociar el contacto a otra persona. La consola SHALL NOT ofrecer borrar la persona ni desasociarla del contacto.

#### Scenario: Precarga con persona asociada

- **WHEN** el agente abre una conversación cuyo contacto ya tiene persona
- **THEN** el panel muestra los datos de la persona precargados

#### Scenario: Búsqueda encontrada

- **WHEN** el contacto no tiene persona y el agente busca un documento existente
- **THEN** el panel precarga los datos de esa persona

#### Scenario: Búsqueda no encontrada

- **WHEN** el contacto no tiene persona y el agente busca un documento inexistente
- **THEN** el panel deja los campos vacíos y habilita el guardado

#### Scenario: Guardar la persona

- **WHEN** el agente guarda el documento y los datos
- **THEN** la consola persiste la persona, asocia el contacto y refleja la etiqueta actualizada

#### Scenario: Editable en conversación finalizada

- **WHEN** el agente abre una conversación `resolved` o `expired`
- **THEN** el panel de la persona permite precargar, buscar, guardar y modificar

#### Scenario: Sin borrar ni desasociar

- **WHEN** el panel muestra la persona de un contacto
- **THEN** no ofrece acciones de borrar la persona ni de desasociarla

#### Scenario: Documento editable para reasociar

- **WHEN** el panel tiene una persona asociada o encontrada
- **THEN** el campo `identification_number` sigue siendo editable

#### Scenario: Cambiar la persona del contacto

- **WHEN** el contacto ya tiene una persona y el agente escribe un documento distinto y guarda
- **THEN** la consola guarda (upsert) la persona de ese documento y asocia el contacto a ella
- **AND** la persona anterior deja de estar referenciada por ese contacto

#### Scenario: Error al guardar

- **WHEN** el guardado responde 400, 404, 422 u otro error
- **THEN** la consola muestra el error y conserva el panel visible

#### Scenario: Validación en el panel

- **WHEN** el agente no completa el `identification_number`, o completa un dato que supera su máximo de grafemas (100; 254 para email y dirección)
- **THEN** la consola indica el error de validación y no envía el guardado inválido

#### Scenario: Datos opcionales vacíos

- **WHEN** el agente deja vacíos `first_name`, `last_name`, `phone_number`, `email` o `address` y guarda
- **THEN** la consola envía el guardado con el documento y esos datos ausentes
