FROM eclipse-temurin:21-jdk AS builder
WORKDIR /app

COPY gradlew .
COPY gradle gradle
COPY build.gradle .
COPY settings.gradle .

COPY user-service user-service
COPY stats-service stats-service
COPY gateway gateway

RUN chmod +x ./gradlew
RUN ./gradlew :gateway:bootJar --no-daemon

FROM eclipse-temurin:21
WORKDIR /app

RUN groupadd -r appgroup && useradd -r -g appgroup appuser
USER appuser

COPY --from=builder /app/gateway/build/libs/*.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-XX:+UseG1GC", "-XX:MaxRAMPercentage=55.0", "-jar", "app.jar"]