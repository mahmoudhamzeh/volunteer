const TOKEN_KEY = "mahak_token";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  const token = getToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);
  if (!(init.body instanceof FormData) && !headers.has("Content-Type") && init.body) {
    headers.set("Content-Type", "application/json");
  }
  const res = await fetch(path, { ...init, headers });
  if (!res.ok) {
    let msg = res.statusText;
    try {
      const body = await res.json();
      msg = body.error || msg;
    } catch {
      /* ignore */
    }
    throw new ApiError(res.status, msg);
  }
  if (res.status === 204) return undefined as T;
  const ct = res.headers.get("content-type") || "";
  if (ct.includes("application/json")) return res.json();
  return undefined as T;
}

export async function downloadAuth(path: string, filename: string) {
  const token = getToken();
  const res = await fetch(path, { headers: token ? { Authorization: `Bearer ${token}` } : {} });
  if (!res.ok) throw new Error("download failed");
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

export async function openAuth(path: string) {
  const token = getToken();
  const res = await fetch(path, { headers: token ? { Authorization: `Bearer ${token}` } : {} });
  if (!res.ok) throw new Error("open failed");
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  window.open(url, "_blank");
}

export const api = {
  login: (email: string, password: string) =>
    request<{ token: string; user: User }>(" /api/v1/auth/login".trim(), {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  register: (email: string, password: string, full_name: string) =>
    request<{ token: string; user: User }>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password, full_name, role: "volunteer" }),
    }),
  me: () => request<{ user: User; volunteer?: Volunteer }>("/api/v1/me"),
  updateProfile: (body: Partial<Volunteer> & { skill_ids?: string[]; skill_categories?: string[] }) =>
    request<Volunteer>("/api/v1/volunteers/me", { method: "PUT", body: JSON.stringify(body) }),
  submitProfile: () => request<Volunteer>("/api/v1/volunteers/me/submit", { method: "POST" }),
  myAvailability: () => request<Availability[]>("/api/v1/volunteers/me/availability"),
  setAvailability: (slots: Availability[]) =>
    request("/api/v1/volunteers/me/availability", { method: "PUT", body: JSON.stringify({ slots }) }),
  myDocs: () => request<DocumentFile[]>("/api/v1/volunteers/me/documents"),
  uploadDoc: (kind: string, file: File) => {
    const fd = new FormData();
    fd.append("kind", kind);
    fd.append("file", file);
    return request<DocumentFile>("/api/v1/volunteers/me/documents", { method: "POST", body: fd });
  },
  tasks: () => request<{ items: Task[]; total: number }>("/api/v1/tasks"),
  acceptTask: (id: string) => request(`/api/v1/tasks/${id}/accept`, { method: "POST" }),
  myAssignments: () => request<Assignment[]>("/api/v1/assignments/me"),
  rateAssignment: (id: string, rating: number, comment: string) =>
    request(`/api/v1/assignments/${id}/rate`, { method: "POST", body: JSON.stringify({ rating, comment }) }),
  missions: () => request<Mission[]>("/api/v1/missions"),
  startMission: (id: string) => request(`/api/v1/missions/${id}/start`, { method: "POST" }),
  missionProgress: (id: string) =>
    request(`/api/v1/missions/${id}/progress`, { method: "POST", body: JSON.stringify({ increment: 1 }) }),
  myMissions: () => request<MissionProgress[]>("/api/v1/missions/me"),
  myCerts: () => request<Certificate[]>("/api/v1/certificates/me"),
  notifications: () => request<Notification[]>("/api/v1/notifications"),
  verify: (code: string) => request<Certificate>(`/api/v1/certificates/verify/${code}`),
  dashboard: () => request<Dashboard>("/api/v1/admin/dashboard"),
  adminVolunteers: (q = "") => request<{ items: Volunteer[]; total: number }>(`/api/v1/admin/volunteers${q}`),
  adminVolunteer: (id: string) =>
    request<{ volunteer: Volunteer; documents: DocumentFile[]; availability: Availability[] }>(
      `/api/v1/admin/volunteers/${id}`,
    ),
  review: (id: string, action: string, reason = "") =>
    request(`/api/v1/admin/volunteers/${id}/review`, { method: "POST", body: JSON.stringify({ action, reason }) }),
  adminTasks: () => request<{ items: Task[]; total: number }>("/api/v1/admin/tasks?limit=100"),
  createTask: (body: unknown) => request("/api/v1/admin/tasks", { method: "POST", body: JSON.stringify(body) }),
  updateTask: (id: string, body: unknown) =>
    request(`/api/v1/admin/tasks/${id}`, { method: "PUT", body: JSON.stringify(body) }),
  adminAssignments: (q = "") =>
    request<{ items: Assignment[]; total: number }>(`/api/v1/admin/assignments${q}`),
  attendance: (id: string) => request(`/api/v1/admin/assignments/${id}/attendance`, { method: "POST" }),
  complete: (id: string, body: { discipline: number; expertise: number; ethics: number; comment: string }) =>
    request(`/api/v1/admin/assignments/${id}/complete`, { method: "POST", body: JSON.stringify(body) }),
  issueCert: (id: string) => request(`/api/v1/admin/assignments/${id}/certificate`, { method: "POST" }),
  issueAggregated: (id: string) =>
    request(`/api/v1/admin/volunteers/${id}/certificates/aggregated`, { method: "POST" }),
  adminMissions: () => request<Mission[]>("/api/v1/admin/missions"),
  createMission: (body: unknown) => request("/api/v1/admin/missions", { method: "POST", body: JSON.stringify(body) }),
  ranking: () => request<RankingRow[]>("/api/v1/admin/reports/ranking?limit=50"),
  skills: () => request<Record<string, number>>("/api/v1/admin/reports/skills"),
  skillCatalog: () => request<SkillGroup[]>("/api/v1/skills"),
  proposeSkill: (group_id: string, title: string) =>
    request<SkillProposal>("/api/v1/volunteers/me/skill-proposals", {
      method: "POST",
      body: JSON.stringify({ group_id, title }),
    }),
  mySkillProposals: () => request<SkillProposal[]>("/api/v1/volunteers/me/skill-proposals"),
  adminSkillCatalog: () => request<SkillGroup[]>("/api/v1/admin/skills"),
  createSkillGroup: (title: string, slug = "") =>
    request<SkillGroup>("/api/v1/admin/skills/groups", {
      method: "POST",
      body: JSON.stringify({ title, slug }),
    }),
  createCatalogSkill: (group_id: string, title: string) =>
    request<SkillItem>("/api/v1/admin/skills", {
      method: "POST",
      body: JSON.stringify({ group_id, title }),
    }),
  updateCatalogSkill: (id: string, body: { title?: string; status?: string; group_id?: string }) =>
    request<SkillItem>(`/api/v1/admin/skills/${id}`, {
      method: "PUT",
      body: JSON.stringify(body),
    }),
  adminSkillProposals: (status = "pending") =>
    request<SkillProposal[]>(`/api/v1/admin/skills/proposals${status ? `?status=${status}` : ""}`),
  reviewSkillProposal: (id: string, body: { action: string; title?: string; group_id?: string; admin_note?: string }) =>
    request<SkillProposal>(`/api/v1/admin/skills/proposals/${id}/review`, {
      method: "POST",
      body: JSON.stringify(body),
    }),
};

export type User = { id: string; email: string; role: "volunteer" | "admin" | "operator" };
export type Volunteer = {
  id: string;
  user_id: string;
  full_name: string;
  national_id: string;
  phone: string;
  phone2?: string;
  province?: string;
  city: string;
  address?: string;
  plaque?: string;
  unit?: string;
  bio: string;
  skill_categories: string[];
  skill_ids?: string[];
  skills?: VolunteerSkill[];
  proposals?: SkillProposal[];
  education_level?: string;
  education_field: string;
  medical_license: string;
  birth_date?: string;
  status: string;
  rejection_reason: string;
  average_score: number;
  total_hours: number;
  completed_tasks: number;
};
export type VolunteerSkill = {
  skill_id: string;
  title: string;
  group_id: string;
  group_slug: string;
  group_title: string;
};
export type SkillItem = {
  id: string;
  group_id: string;
  title: string;
  status: string;
  group_title?: string;
};
export type SkillGroup = {
  id: string;
  slug: string;
  title: string;
  sort_order: number;
  skills: SkillItem[];
};
export type SkillProposal = {
  id: string;
  volunteer_id: string;
  volunteer_name?: string;
  group_id: string;
  group_title: string;
  title: string;
  status: string;
  admin_note: string;
  created_skill_id?: string;
  created_at: string;
};
export type Availability = { weekday: number; start_time: string; end_time: string };
export type DocumentFile = { id: string; kind: string; file_name: string; mime_type: string; created_at: string };
export type Task = {
  id: string;
  title: string;
  description: string;
  location: string;
  starts_at: string;
  ends_at: string;
  capacity: number;
  reserved_count: number;
  hour_weight: number;
  required_skills: string[];
  required_skill_ids?: string[];
  min_score: number;
  required_education: string;
  status: string;
};
export type Assignment = {
  id: string;
  task_id: string;
  volunteer_id: string;
  status: string;
  volunteer_rating?: number;
  composite_score?: number;
  hours_awarded: number;
  task?: { title: string; location: string; starts_at: string; hour_weight: number };
  volunteer?: { full_name: string };
};
export type Mission = {
  id: string;
  title: string;
  description: string;
  kind: string;
  hour_weight: number;
  deadline_hours?: number;
  target_count: number;
  webhook_event?: string;
  status: string;
};
export type MissionProgress = {
  id: string;
  mission_id: string;
  status: string;
  progress: number;
  mission?: Mission;
};
export type Certificate = {
  id: string;
  verification_code: string;
  title: string;
  hours: number;
  volunteer_name?: string;
  kind: string;
  issued_at: string;
  authentic?: boolean;
};
export type Notification = { id: string; title: string; body: string; read: boolean; created_at: string };
export type Dashboard = {
  total_volunteers: number;
  pending_volunteers: number;
  approved_volunteers: number;
  open_tasks: number;
  active_assignments: number;
  completed_this_month: number;
  participation_rate: number;
  total_hours: number;
  skill_distribution: Record<string, number>;
};
export type RankingRow = {
  volunteer_id: string;
  full_name: string;
  city: string;
  average_score: number;
  total_hours: number;
  completed_tasks: number;
  status: string;
};
