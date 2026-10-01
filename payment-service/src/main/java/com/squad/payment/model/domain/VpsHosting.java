package com.squad.payment.model.domain;

import com.squad.payment.model.enums.Hosting;
import com.squad.payment.model.enums.converters.HostingConverter;
import jakarta.persistence.Convert;
import jakarta.persistence.Entity;
import jakarta.persistence.Table;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "vps_hosting")
@AllArgsConstructor
@NoArgsConstructor
@Builder
public class VpsHosting extends BaseEntity {

    @Convert(converter = HostingConverter.class)
    private Hosting hosting;

    private String apiToken;
}
