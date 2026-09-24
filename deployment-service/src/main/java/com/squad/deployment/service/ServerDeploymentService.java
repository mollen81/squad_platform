package com.squad.deployment.service;

import com.squad.deployment.dto.ServerDeploymentFinishedEvent;
import com.squad.deployment.dto.VpsPurchasedEvent;
import com.squad.deployment.kafka.ServerDeploymentFinishedProducer;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.schmizz.sshj.SSHClient;
import net.schmizz.sshj.sftp.SFTPClient;
import net.schmizz.sshj.transport.verification.PromiscuousVerifier;
import net.schmizz.sshj.xfer.InMemorySourceFile;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.util.List;

@Service
@Slf4j
@RequiredArgsConstructor
public class ServerDeploymentService {

    @Autowired
    private final ServerDeploymentFinishedProducer kafkaProducer;

    @Async
    public void deploySquadServer(VpsPurchasedEvent event) {
        try (SSHClient ssh = new SSHClient()){
            ssh.addHostKeyVerifier(new PromiscuousVerifier());
            ssh.connect(event.ip());
            ssh.authPassword("root", event.rootPassword());

            log.info("[{}] Connected. Docker installation...", event.ip());
            executeCommand(ssh, "apt-get update && apt-get install -y docker.io docker-compose-v2");

            log.info("[{}] Working directories creation...", event.ip());
            executeCommand(ssh, "mkdir -p /opt/squad-server/config");

            log.info("[{}] Config generation and installing...", event.ip());
            try (SFTPClient sftp = ssh.newSFTPClient()) {
                String adminsConfig = generateAdminsConfig(event);
                uploadTextAsFile(sftp, adminsConfig, "/opt/squad-server/config/Admins.cfg");

                log.info("[{}] Генерация белого списка игроков...", event.ip());
                String whitelistJson = generateWhitelistJson(event.whitelistSteamIds());
                uploadTextAsFile(sftp, whitelistJson, "/opt/squadjs/config/whitelist.json");

                String dockerCompose = generateDockerCompose(event);
                uploadTextAsFile(sftp, dockerCompose, "/opt/squad-server/docker-compose.yml");
            }

            log.info("[{}] Containers start (SteamCMD downloading the game)...", event.ip());
            executeCommand(ssh, "cd /opt/squad-server && docker compose up -d");

            log.info("[{}] Deployment is finished!", event.eventId());

            // TODO: Отправить в Kafka событие ServerOnlineEvent(eventId, ipAddress) для фронтенда
            kafkaProducer.sendServerDeploymentFinishedEvent(ServerDeploymentFinishedEvent.builder()
                            .serverName(event.serverConfig().serverName())
                            .password(event.serverConfig().password())
                            .build());

        }
        catch (Exception e) {
            log.error("Server deployment error for event {}: {}", event.eventId(), e.getMessage(), e);
            // TODO: Отправить в Kafka событие ServerDeploymentFailedEvent, чтобы биллинг отменил сервер
        }
    }

    private void executeCommand(SSHClient ssh, String command) throws IOException {
        try (var session = ssh.startSession()) {
            var cmd = session.exec(command);
            cmd.join();
            if (cmd.getExitStatus() != null && cmd.getExitStatus() != 0) {
                throw new RuntimeException("Command ended with error: " + command);
            }
        }
    }

    private String generateAdminsConfig(VpsPurchasedEvent event) {
        StringBuilder sb = new StringBuilder();
        sb.append("// Auto-generated Admins config for Event: ").append(event.eventId()).append("\n");
        sb.append("Group=SuperAdmin:kick,ban,config,cameraman,immunity,manageserver,chat,cheat,clientui\n\n");

        for (var admin : event.admins()) {
            sb.append("Admin=").append(admin.steamId()).append(":").append(admin.role()).append("\n");
        }
        return sb.toString();
    }

    private String generateWhitelistJson(List<String> steamIds) {
        return "[\n  \"" + String.join("\",\n  \"", steamIds) + "\"\n]";
    }

    private void uploadTextAsFile(SFTPClient sftp, String content, String remotePath) throws IOException {
        sftp.put(new InMemorySourceFile() {
            @Override public String getName() { return "memory-file"; }
            @Override public long getLength() { return content.getBytes().length; }
            @Override public InputStream getInputStream() { return new ByteArrayInputStream(content.getBytes()); }
        }, remotePath);
    }

    private String generateDockerCompose(VpsPurchasedEvent event) {
        return """
            version: '3.8'
            services:
              squad-server:
                image: cm2network/steamcmd:latest
                container_name: squad-server
                network_mode: host
                environment:
                  - SERVER_NAME=%s
                  - SERVER_PASSWORD=%s
                  - MAP=%s
                volumes:
                  - ./config:/server/config
                restart: unless-stopped
            """.formatted(
                event.serverConfig().serverName(),
                event.serverConfig().password(),
                event.serverConfig().map()
        );
    }
}
