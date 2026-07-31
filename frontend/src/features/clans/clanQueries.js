import { gql } from '@apollo/client';

export const GET_CLAN_WITH_MEMBERS = gql`
  query GetClanWithMembers($clanId: ID!) {
    getClanWithMembers(clanId: $clanId) {
      id
      name
      tag
      description
      requirements
      avatarUrl
      isRecruiting
      status
      totalElo
      minElo
      createdAt
      members {
        id
        userId
        role
      }
    }
  }
`;

export const APPLY_TO_CLAN = gql`
  mutation ApplyToClan(
    $userId: String!
    $clanId: ID!
    $socialLink: String
    $experienceText: String
  ) {
    applyToClan(
      userId: $userId
      clanId: $clanId
      socialLink: $socialLink
      experienceText: $experienceText
    ) {
      applicationId
      message
    }
  }
`;