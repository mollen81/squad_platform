package com.squad.event.model.domain;

import com.squad.event.model.enums.EventServerStatus;
import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.Instant;
import java.util.UUID;

@Entity
@Table(name = "event_server")
@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
@Builder
public class EventServer {
    @Id
    @Column(name = "event_id", nullable = false)
    private UUID eventId;

    @Enumerated(EnumType.STRING)
    @Column(length = 32, nullable = false)
    private EventServerStatus status;

    @Column(name = "ip_address")
    private String ipAddress;

    private String password;

    @Column(name = "deployed_at")
    private Instant deployedAt;

    @CreationTimestamp
    @Column(name = "created_at", nullable = false)
    private Instant createdAt;

    @UpdateTimestamp
    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;

    public boolean isDataAvailable() {
        return ipAddress != null && password != null &&
                (status == EventServerStatus.DEPLOYED || status == EventServerStatus.READY);
    }
}
