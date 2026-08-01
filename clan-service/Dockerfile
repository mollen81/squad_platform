FROM public.ecr.aws/docker/library/eclipse-temurin:21 AS builder
WORKDIR /app

COPY gradlew .
COPY gradle gradle
COPY build.gradle .
COPY settings.gradle .

COPY clan-service clan-service
COPY stats-service stats-service

RUN chmod +x ./gradlew
RUN ./gradlew :clan-service:bootJar --no-daemon

FROM eclipse-temurin:21
WORKDIR /app

RUN groupadd -r appgroup && useradd -r -g appgroup appuser
USER appuser

COPY --from=builder /app/clan-service/build/libs/*.jar app.jar

EXPOSE 8084
EXPOSE 9094

ENTRYPOINT ["java", "-XX:+UseG1GC", "-XX:MaxRAMPercentage=65.0", "-jar", "app.jar"]