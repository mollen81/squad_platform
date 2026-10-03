package com.squad.payment.model.domain;

import com.squad.payment.model.enums.Hosting;
import com.squad.payment.model.enums.converters.HostingConverter;
import jakarta.persistence.Convert;
import jakarta.persistence.Entity;
import jakarta.persistence.Table;
import lombok.*;

@Entity
@Table(name = "vps_hosting")
@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
@Builder
public class VdsHosting extends BaseEntity {

    @Convert(converter = HostingConverter.class)
    private Hosting hostingName;

    private String apiToken;
}
