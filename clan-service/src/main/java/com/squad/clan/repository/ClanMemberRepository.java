package com.squad.clan.repository;

import com.squad.clan.entity.ClanMember;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface ClanMemberRepository extends JpaRepository<ClanMember, UUID> {
    Optional<ClanMember> findByUserId(UUID userId);
    int countByClanId(UUID clanId);

    @Query("SELECT c FROM Clan c LEFT JOIN FETCH c.members m WHERE c.id = :clanId")
    List<ClanMember> findAllClanMembersByClanId(UUID clanId);
}
