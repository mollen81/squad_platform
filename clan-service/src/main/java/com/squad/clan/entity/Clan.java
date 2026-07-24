package com.squad.clan.entity;

import com.squad.clan.enums.ClanStatus;
import jakarta.persistence.*;
import lombok.*;
import org.hibernate.annotations.CreationTimestamp;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

@Entity
@Table(name = "clans")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class Clan {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    @Column(nullable = false, unique = true, length = 50)
    private String name;

    @Column(nullable = false, length = 3)
    private String tag;

    @Column(columnDefinition = "TEXT")
    private String description;

    @Column(columnDefinition = "TEXT")
    private String requirements;

    @Column(name = "min_elo")
    @Builder.Default
    private final int minElo = 500;

    @Column(name = "avatar_url")
    private String avatar_url;

    @Column(name = "is_recruiting")
    @Builder.Default
    private Boolean isRecruiting = true;

    @Enumerated(EnumType.STRING)
    @Builder.Default
    private ClanStatus clanStatus;

    @Column(name = "total_elo")
    @Builder.Default
    private Integer totalElo = 0;

    @Column(name = "created_at", updatable = false)
    @CreationTimestamp
    private LocalDateTime createdAt;

    @OneToMany(mappedBy = "clan", cascade = CascadeType.ALL, orphanRemoval = true)
    @Builder.Default
    private List<ClanMember> members = new ArrayList<>();
}
