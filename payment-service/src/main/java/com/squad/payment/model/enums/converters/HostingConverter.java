package com.squad.payment.model.enums.converters;

import com.squad.payment.model.enums.Hosting;
import jakarta.persistence.AttributeConverter;
import jakarta.persistence.Converter;

@Converter(autoApply = true)
public class HostingConverter implements AttributeConverter<Hosting, String> {
    @Override
    public String convertToDatabaseColumn(Hosting attribute) {
        return attribute == null ? null : attribute.name();
    }

    @Override
    public Hosting convertToEntityAttribute(String dbData) {
        return dbData == null ? null : Hosting.fromString(dbData);
    }
}
