export type UserRole = 'ROLE_USER' | 'ROLE_CLAN_LEADER' | 'ROLE_ADMIN';

export interface UserPrincipal {
  id: string;
  username: string;
  email: string;
  steamId: string | null;
  rating: number;
  role: UserRole;
  squadHours?: number;
}

export interface AuthSession {
  token: string;
  principal: UserPrincipal;
}
