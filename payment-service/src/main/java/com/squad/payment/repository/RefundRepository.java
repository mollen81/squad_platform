package com.squad.payment.repository;

import com.squad.payment.model.domain.Refund;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface RefundRepository extends JpaRepository<Refund, String> {
    Optional<Refund> findByRefundId(String refundId);
    Optional<Refund> findByPaymentId(String paymentId);
}
