package com.squad.payment.model.domain;

import com.squad.payment.model.enums.PaymentStatus;
import com.squad.payment.model.enums.converters.PaymentStatusConverter;
import jakarta.persistence.*;
import lombok.*;

import java.math.BigDecimal;
import java.util.List;

@Entity
@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
@Builder
public class Payment extends BaseEntity {
    private BigDecimal amount;

    @Column(name = "status")
    @Convert(converter = PaymentStatusConverter.class)
    private PaymentStatus paymentStatus;

    private String errorMessage;

    @Column(name = "event_id", nullable = false)
    private String eventId;

    // Aeza service id
    @Column(name = "external_service_id")
    private Long externalServiceId;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "vps_hosting_id")
    private VdsHosting vdsHosting;

    @OneToMany(mappedBy = "payment", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<Refund> refunds;
}
