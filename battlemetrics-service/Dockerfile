FROM eclipse-temurin:21 AS builder
WORKDIR /app

COPY gradlew .
COPY gradle gradle
COPY build.gradle .
COPY settings.gradle .

COPY battlemetrics-service battlemetrics-service

RUN chmod +x ./gradlew
RUN ./gradlew :battlemetrics-service:bootJar --no-daemon

FROM eclipse-temurin:21
WORKDIR /app

RUN groupadd -r appgroup && useradd -r -g appgroup appuser
USER appuser

COPY --from=builder /app/battlemetrics-service/build/libs/*.jar app.jar

EXPOSE 8082
EXPOSE ${SPRING_KAFKA_BOOTSTRAP_SERVERS_PORT}

ENTRYPOINT ["java", "-XX:+UseG1GC", "-XX:MaxRAMPercentage=55.0", "-jar", "app.jar"]