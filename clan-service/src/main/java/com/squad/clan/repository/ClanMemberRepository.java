package com.squad.clan.repository;

import com.squad.clan.entity.ClanMember;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;
import java.util.UUID;

@Repository
public interface ClanMemberRepository extends JpaRepository<ClanMember, UUID> {
    Optional<ClanMember> findByUserId(UUID userId);
    int countByClanId(UUID clanId);
}
