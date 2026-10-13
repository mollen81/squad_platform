package com.squad.event.service;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.passay.data.CharacterData;
import org.passay.data.EnglishCharacterData;
import org.passay.generate.PasswordGenerator;
import org.passay.rule.CharacterRule;
import org.springframework.stereotype.Service;


@Service
@Slf4j
@RequiredArgsConstructor
public class PasswordGeneratorService {

    private final String ERROR_CODE = "error: password generation is invalid";

    public String generatePassword(int length) {
        return generateRandomSpecialCharacters(length);
    }

    private String generateRandomSpecialCharacters(int length) {

        CharacterData lowerCaseChars = EnglishCharacterData.LowerCase;
        CharacterRule lowerCaseRule = new CharacterRule(lowerCaseChars, 2 );

        CharacterData upperCaseChars = EnglishCharacterData.UpperCase;
        CharacterRule upperCaseRule = new CharacterRule(upperCaseChars, 2);

        CharacterData digitChars = EnglishCharacterData.Digit;
        CharacterRule digitCharsRule = new CharacterRule(digitChars, 2);

        CharacterData specialChars = new CharacterData() {
            @Override
            public String getErrorCode() {
                return ERROR_CODE;
            }

            @Override
            public String getCharacters() {
                return "!@#$%^&*()_+";
            }
        };

        CharacterRule specialCharacterRule = new CharacterRule(specialChars, 2);


        org.passay.generate.PasswordGenerator generator = new PasswordGenerator(
                length,
                lowerCaseRule, upperCaseRule, digitCharsRule, specialCharacterRule);

        return generator.generate().toString();
    }
}
