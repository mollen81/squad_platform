FROM eclipse-temurin:21-jdk AS builder
WORKDIR /app

COPY gradlew .
COPY gradle gradle
COPY build.gradle .
COPY settings.gradle .
COPY user-service/build.gradle user-service/

RUN chmod +x ./gradlew

RUN ./gradlew :user-service:dependencies --no-daemon || true

COPY user-service/src user-service/src

RUN ./gradlew :user-service:bootJar --no-daemon -x test

FROM eclipse-temurin:21-jre
WORKDIR /app

RUN groupadd -r appgroup && useradd -r -g appgroup appuser
COPY --from=builder --chown=appuser:appgroup /app/user-service/build/libs/*-ALPHA.jar app.jar

USER appuser

EXPOSE 8081
EXPOSE 9091

ENTRYPOINT ["java", "-XX:+UseG1GC", "-XX:MaxRAMPercentage=55.0", "-jar", "app.jar"]