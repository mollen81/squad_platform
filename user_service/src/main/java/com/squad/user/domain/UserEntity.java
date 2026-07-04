package com.squad.user.domain;

import jakarta.persistence.*;
import lombok.*;

import java.time.OffsetDateTime;

@Entity
@Table(name = "users")
@Getter
@Setter
@AllArgsConstructor
@NoArgsConstructor
@Builder
public class UserEntity {
    @Id
    private String id;

    @Column(name = "steam_id", unique = true, nullable = false)
    private String steamId;

    @Column(nullable = false)
    private String nickname;

    @Column(name = "avatar_url")
    private String avatarUrl;

    @Column(name = "joined_at", updatable = false)
    private OffsetDateTime joinedAt;

    @Column(name = "last_login_at")
    private OffsetDateTime lastLoginAt;

    @PrePersist
    protected void onCreate() {
        this.joinedAt = OffsetDateTime.now();
    }
}
