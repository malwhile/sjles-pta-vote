export interface Poll {
  id: number;
  question: string;
  created_at: string;
  updated_at: string;
  expires_at: string;
  member_yes: number;
  member_no: number;
  non_member_yes: number;
  non_member_no: number;
}

export interface Member {
  Name: string;
  Email: string;
}

export interface LoginResponse {
  success: boolean;
  token: string;
  error?: string;
}

export interface MembersViewResponse {
  success: boolean;
  members: Member[];
  error?: string;
}
