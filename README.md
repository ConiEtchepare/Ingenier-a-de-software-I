# Ingenieria de Software I - Trabajo práctico especial ADD

En este repositorio combinamos toda la documentación del diseño arquitectónico que fuimos haciendo en las iteraciones con una prueba de concepto en Go para mostrar que la arquitectura es totalmente viable y funciona.

Nos enfocamos en diseñar la arquitectura del sistema paso a paso: 

1. **Análisis Inicial:** Armamos el diagrama de contexto, definimos el alcance y fijamos los Atributos de Calidad (QA), Requerimientos funcionales (RF), Restricciones (CON) y Preocupaciones (CNR) clave para el sistema.
2. **Diseño con ADD:**
   * **Iteración 1:** Definimos los módulos principales de la arquitectura.
   * **Iteración 2:** Diseñamos la ingesta asíncrona para procesar correos.
   * **Iteración 3:** Definimos el esquema de datos y cómo guardar los itinerarios.
   * **Iteración 4:** Aislamos la integración con agencias externas usando la táctica Split Module para que facilitar la incorporación de nuevas agencias.
   * **Iteración 5 (2da Entrega):** Integración y escalabilidad del **Asistente Virtual / Chatbot** como microservicio *stateless* con gestión de sesiones en caché.
   * **Iteración 6 (2da Entrega):** Refinamiento interno del Asistente Virtual mediante Adaptador NLU encapsulado y política híbrida de escalamiento hacia el Enrutador de Agencias.
3. **Decisiones de Arquitectura (ADRs):** Documentación atómica bajo formato MADR (ADR-01 a ADR-13) en la Wiki del repositorio.
4. **Contratos Spec-Driven:** Especificación formal previa a la implementación en `api/openapi.yaml` (endpoints REST sincrónicos) y `api/asyncapi.yaml` (eventos asíncronos).
5. **Prototipo Funcional en Go:** Implementación modular verificable del API Gateway, el Servicio Central de Itinerarios, el Enrutador de Agencias, el Asistente Virtual (Chatbot) y una interfaz web de demostración.

## Prueba de prototipo
* Para levantar el API Gateway: go run cmd/apigateway/main.go. El servidor iniciará en http://localhost:8080.
* Consultar itinerarios activos (GET /itinerarios)
  - curl http://localhost:8080/itinerarios
* Enviar consultas simples al chatbot (POST /assistant/messages)
  - curl -X POST http://localhost:8080/assistant/messages \ -H "Content-Type: application/json" \ -d '{"mensaje": "¿Cuál es el estado de mi vuelo?"}'
* Probar escalamiento complejo a la agencia (POST /assistant/messages)
  - curl -X POST http://localhost:8080/assistant/messages \ -H "Content-Type: application/json" \ -d '{"mensaje": "Perdí mi conexión por tormenta y necesito cancelar"}'

## Prueba desde el Navegador
Abrir http://localhost:8080 en el navegador para interactuar con el panel de viajes activos y el Asistente Virtual, probando los tres flujos principales:
- Consulta de estado: Responde en primer nivel leyendo las reservas del viaje seleccionado (CONSULTA_RESUELTA).
- Reprogramación menor: Actualiza el horario/check-in de la reserva directamente en el panel (REPROGRAMACION_MENOR).
- Caso complejo / Emergencia: Deriva automáticamente al Enrutador de Agencias, selecciona el adaptador de la agencia preferida del usuario y genera un ticket adjuntando el contexto de la conversación (ESCALADO_AGENCIA).

## Integrantes:
- Abril Colonna
- Constanza Etchepare
- Agustina Gil
