package com.squad.gateway.record;

import java.util.List;

public record Clan(
        String id,
        String name,
        String tag,
        String description,
        String requirements,
        String avatarUrl,
        boolean isRecruiting,
        String status,
        int totalElo,
        int minElo,
        String createdAt,
        List<ClanMember> members
) {
}
