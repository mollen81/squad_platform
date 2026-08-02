const GATEWAY_URL = 'http://localhost:8080/graphql';

export const fetchGraphQL = async (query, variables = {}) => {
    const token = localStorage.getItem('squad_jwt');

    const headers = {
        'Content-Type': 'application/json',
    };

    // Если токен есть, прикрепляем его (понадобится нам в будущем для защищенных эндпоинтов)
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    try {
        const response = await fetch(GATEWAY_URL, {
            method: 'POST',
            headers,
            body: JSON.stringify({ query, variables }),
        });

        const json = await response.json();

        if (json.errors) {
            throw new Error(json.errors[0].message);
        }

        return json.data;
    } catch (error) {
        console.error("GraphQL Request Failed:", error);
        throw error;
    }
};

export const QUERIES = {
    GET_STATS: `
    query getStats($userId: String!) {
      getPlayerStats(userId: $userId) {
        eloRating
        kills
        deaths
        revives
        favouriteRole
        totalPlaytimeHours
      }
    }
  `
};

export const MUTATIONS = {
    LOGIN_STEAM: `
    mutation login($paramsJson: String!) {
      loginWithSteam(openidParamsJson: $paramsJson) {
        userId
        steamId
        token
        isNewUser
      }
    }
  `
};