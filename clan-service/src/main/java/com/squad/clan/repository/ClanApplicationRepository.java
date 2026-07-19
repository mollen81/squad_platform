package com.squad.clan.repository;

import com.squad.clan.entity.ClanApplication;
import com.squad.clan.enums.ApplicationStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface ClanApplicationRepository extends JpaRepository<ClanApplication, UUID> {

    List<ClanApplication> findAllByClanIdAndStatus(UUID clanId, ApplicationStatus status);

    boolean existsByUserIdAndClanIDAndStatus(UUID userId, UUID clanId, ApplicationStatus status);
}
