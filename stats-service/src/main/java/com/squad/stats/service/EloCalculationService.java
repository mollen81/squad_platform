package com.squad.stats.service;

import org.springframework.stereotype.Service;

@Service
public class EloCalculationService {

    private static final int MIN_ELO = 100;

    // Elo calculation in registration moment
    public int calculateInitialElo(int playtimeHours) {
        if (playtimeHours < 50) return 900;
        if (playtimeHours < 300) return 1000;
        if (playtimeHours < 800) return 1150;
        if (playtimeHours < 1500) return 1300;
        return 1450;
    }

    // Elo calculation after match
    public int calculateMatchElo(int currentElo, boolean isWin, String role,
                                 int kills, int deaths, int revives, int destroyedVehicles) {

        // 1. Базовые очки за исход матча
        int baseDelta = isWin ? 25 : -25;

        // 2. Личные показатели
        int kdDiff = kills - deaths;
        double kdRatio = deaths > 0 ? (double) kills / deaths : kills;

        String safeRole = role != null ? role : "Rifleman";

        // 3. Расчет личного модификатора в зависимости от роли (балансировка вашей формулы под рамки одного матча)
        double performanceDelta = switch (safeRole) {

            // Пехоте важны фраги и немного поднятия
            case "Rifleman", "Ambusher", "Raider", "Automatic Rifleman", "Machine Gunner" ->
                    kdDiff * 1.3 + revives * 0.2;

            // Медику фраги почти не дают бонуса, главный упор на поднятия
            case "Medic" ->
                    kdDiff * 0.35 + revives * 1.5;

            // Снайпер наказывается за K/D ниже 1.0 и получает сильный буст за высокий K/D
            case "Sniper", "Marksman" ->
                    (kdRatio - 1.0) * 5.0 + kdDiff * 0.8;

            case "Grenadier", "Scout" ->
                    (kdRatio - 1.0) * 2.0 + kdDiff * 0.5 + revives * 0.1;

            // Командирам даем фиксированный бонус за организацию (компенсирует просадки по K/D)
            case "Squad Leader" ->
                    5.0 + kdDiff * 1.0;

            case "Lead Crewman", "Crewman" ->
                    kdDiff * 1.5;

            case "Lead Pilot" ->
                    4.0; // Пилоту сложно считать K/D, даем статический плюс

            // Трубам важнее уничтожать технику, чем пехоту
            case "Light Anti-Tank" ->
                    kdDiff * 1.0 + destroyedVehicles * 2.5;

            case "Heavy Anti-Tank" ->
                    kdDiff * 0.8 + destroyedVehicles * 3.25;

            case "Combat Engineer", "Sapper", "Saboteur" ->
                    kdDiff * 1.15 + destroyedVehicles * 2.0;

            case "Infiltrator" ->
                    kdDiff * 1.3 + destroyedVehicles * 1.5;

            default ->
                    kdDiff * 1.0 + revives * 0.1;
        };

        // 4. Суммируем базовую дельту и личный перфоманс
        int finalDelta = baseDelta + (int) Math.round(performanceDelta);

        // 5. Жесткие лимиты: чтобы за один матч нельзя было получить +200 или -200 ELO (кап +/- 50)
        finalDelta = Math.clamp(finalDelta, -50, 50);

        return Math.max(MIN_ELO, currentElo + finalDelta);
    }
}