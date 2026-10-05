# Этап 1: Сборка и распаковка слоев
FROM eclipse-temurin:21-jdk AS builder
WORKDIR /build

# Копируем только файлы конфигурации Gradle для кэширования зависимостей
COPY gradlew .
COPY gradle ./gradle
COPY settings.gradle .
COPY build.gradle .

# Копируем исходный код и собираем проект
COPY event-service/build.gradle ./event-service/
COPY event-service/src ./event-service/src

RUN ./gradlew :event-service:clean :event-service:build -x test --no-daemon

# 4. Переходим в папку с собранным артефактом и распаковываем JAR на слои
WORKDIR /build/event-service/build/libs
RUN java -Djarmode=layertools -jar *.jar extract

# Этап 2: Легковесный Runtime
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app

# Создаем пользователя без root-прав для безопасности
RUN addgroup -S spring && adduser -S spring -G spring
USER spring:spring

# Копируем слои из папки сборки конкретного модуля
COPY --from=builder /build/event-service/build/libs/dependencies/ ./
COPY --from=builder /build/event-service/build/libs/spring-boot-loader/ ./
COPY --from=builder /build/event-service/build/libs/snapshot-dependencies/ ./
COPY --from=builder /build/event-service/build/libs/application/ ./

EXPOSE 8086
EXPOSE 9096

ENTRYPOINT ["java", "org.springframework.boot.loader.launch.JarLauncher"]