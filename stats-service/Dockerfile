FROM eclipse-temurin:21 AS builder
WORKDIR /app

COPY gradlew .
COPY gradle gradle
COPY build.gradle .
COPY settings.gradle .

COPY stats-service stats-service

RUN chmod +x ./gradlew
RUN ./gradlew :stats-service:bootJar --no-daemon

FROM eclipse-temurin:21
WORKDIR /app

RUN groupadd -r appgroup && useradd -r -g appgroup appuser
USER appuser

COPY --from=builder /app/stats-service/build/libs/*.jar app.jar

EXPOSE 8081
EXPOSE 9091

ENTRYPOINT ["java", "-XX:+UseG1GC", "-XX:MaxRAMPercentage=55.0", "-jar", "app.jar"]