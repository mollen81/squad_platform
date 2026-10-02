package com.squad.payment.repository;

import com.squad.payment.model.domain.VpsHosting;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface VpsHostingRepository extends JpaRepository<VpsHosting, String> {
    Optional<VpsHosting> findByHostingId(String hostingId);
    Optional<VpsHosting> findByHostingName(String hostingName);
}
