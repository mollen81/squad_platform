package com.squad.clan.repository;

import com.squad.clan.entity.Clan;
import com.squad.clan.entity.ClanMember;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface ClanRepository extends JpaRepository<Clan, UUID> {

    boolean existsByName(String name);

    @Query("SELECT c FROM Clan c LEFT JOIN FETCH c.members WHERE c.id = :clanId")
    Optional<Clan> findByIdWithAllMembers(@Param("clanId") UUID clanId);
}
