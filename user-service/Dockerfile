# Этап 1: Сборка и распаковка слоев
FROM eclipse-temurin:21-jdk AS builder
WORKDIR /build

# 1. Копируем общие файлы сборщика из корня
COPY gradlew .
COPY gradle ./gradle
COPY settings.gradle .
COPY build.gradle .

# 2. Копируем конфигурацию и исходный код именно user-service
COPY user-service/build.gradle ./user-service/
COPY user-service/src ./user-service/src

# 3. Собираем только user-service
RUN ./gradlew :user-service:clean :user-service:build -x test --no-daemon

# 4. Распаковываем JAR на слои через layertools
WORKDIR /build/user-service/build/libs
RUN java -Djarmode=layertools -jar *[!plain].jar extract

# Этап 2: Финальный легковесный Runtime
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app

RUN addgroup -S spring && adduser -S spring -G spring
USER spring:spring

COPY --from=builder /build/user-service/build/libs/dependencies/ ./
COPY --from=builder /build/user-service/build/libs/spring-boot-loader/ ./
COPY --from=builder /build/user-service/build/libs/snapshot-dependencies/ ./
COPY --from=builder /build/user-service/build/libs/application/ ./

EXPOSE 8081
EXPOSE 9091

ENTRYPOINT ["java", "org.springframework.boot.loader.launch.JarLauncher"]