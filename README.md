# Ingenieria de Software I - Trabajo práctico especial ADD

En este repositorio combinamos toda la documentación del diseño arquitectónico que fuimos haciendo en las iteraciones con una prueba de concepto en Go para mostrar que la arquitectura es totalmente viable y funciona.

Nos enfocamos en diseñar la arquitectura del sistema paso a paso: 

1. **Análisis Inicial:** Armamos el diagrama de contexto, definimos el alcance y fijamos los Atributos de Calidad (QA), Requerimientos funcionales (RF), Restricciones (CON) y Preocupaciones (CNR) clave para el sistema.
2. **Diseño con ADD:**
   * **Iteración 1:** Definimos los módulos principales de la arquitectura.
   * **Iteración 2:** Diseñamos la ingesta asíncrona para procesar correos.
   * **Iteración 3:** Definimos el esquema de datos y cómo guardar los itinerarios.
   * **Iteración 4:** Aislamos la integración con agencias externas usando la táctica Split Module para que facilitar la incorporación de nuevas agencias.
3. **Decisiones de Arquitectura (ADRs):** Dejamos asentadas las decisiones más importantes (ADR-01 a ADR-04).
4. **Contratos Spec-Driven:** Definimos las APIs antes de programar (`openapi.yaml` para la parte REST y `asyncapi.yaml` para los eventos asíncronos).
5. **Prototipo de Viabilidad:** Armamos un esqueleto modular en Go y un API Gateway básico para verificar que los endpoints responden según el contrato.

## Prueba de prototipo
* Para levantar el API Gateway
- go run cmd/apigateway/main.go
* Probar la respuesta del servidor
- curl http://localhost:8080/itinerarios

## Integrantes:
- Abril Colonna
- Constanza Etchepare
- Agustina Gil
