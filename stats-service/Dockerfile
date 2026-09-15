# Этап 1: Сборка и распаковка слоев
FROM eclipse-temurin:21-jdk-alpine AS builder
WORKDIR /build

# Копируем только файлы конфигурации Gradle для кэширования зависимостей
COPY build.gradle settings.gradle gradlew ./
COPY gradle ./gradle
RUN ./gradlew dependencies --no-daemon || true

# Копируем исходный код и собираем проект
COPY src ./src
RUN ./gradlew clean build -x test --no-daemon

# Извлекаем слои из собранного fat-jar
RUN java -Djarmode=layertools -jar build/libs/*.jar extract

# Этап 2: Легковесный Runtime
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app

# Создаем пользователя без root-прав для безопасности
RUN addgroup -S spring && adduser -S spring -G spring
USER spring:spring

# Копируем слои в порядке частоты их изменения
COPY --from=builder /build/dependencies/ ./
COPY --from=builder /build/spring-boot-loader/ ./
COPY --from=builder /build/snapshot-dependencies/ ./
COPY --from=builder /build/application/ ./

EXPOSE 8080

ENTRYPOINT ["java", "org.springframework.boot.loader.launch.JarLauncher"]