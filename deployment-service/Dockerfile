# Этап 1: Сборка и распаковка слоев
FROM eclipse-temurin:21-jdk AS builder
WORKDIR /build

# Копируем только файлы конфигурации Gradle для кэширования зависимостей
COPY gradlew .
COPY gradle ./gradle
COPY settings.gradle .
COPY build.gradle .

# Копируем исходный код и собираем проект
COPY deployment-service/build.gradle ./deployment-service/
COPY deployment-service/src ./deployment-service/src

RUN ./gradlew :deployment-service:clean :deployment-service:build -x test --no-daemon

# 4. Переходим в папку с собранным артефактом и распаковываем JAR на слои
WORKDIR /build/deployment-service/build/libs
RUN java -Djarmode=layertools -jar *.jar extract

# Этап 2: Легковесный Runtime
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app

# Создаем пользователя без root-прав для безопасности
RUN addgroup -S spring && adduser -S spring -G spring
USER spring:spring

# Копируем слои из папки сборки конкретного модуля
COPY --from=builder /build/deployment-service/build/libs/dependencies/ ./
COPY --from=builder /build/deployment-service/build/libs/spring-boot-loader/ ./
COPY --from=builder /build/deployment-service/build/libs/snapshot-dependencies/ ./
COPY --from=builder /build/deployment-service/build/libs/application/ ./

EXPOSE 8087

ENTRYPOINT ["java", "org.springframework.boot.loader.launch.JarLauncher"]