const API_BASE = '/api';

interface ApiError {
  error: string;
}

class ApiClient {
  private token: string | null = null;

  constructor() {
    this.token = localStorage.getItem('token');
  }

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
  }

  getToken(): string | null {
    return this.token;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
    skipAuthRedirect = false
  ): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
    });

    if (response.status === 401 && !skipAuthRedirect) {
      this.setToken(null);
      window.location.href = '/login';
      throw new Error('Unauthorized');
    }

    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.error || 'Request failed');
    }

    return response.json();
  }

  // Auth
  async login(username: string, password: string): Promise<LoginResponse> {
    const result = await this.request<LoginResponse>(
      '/auth/login',
      {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      },
      true // Skip auth redirect for login endpoint
    );
    if (result.token) {
      this.setToken(result.token);
    }
    return result;
  }

  async verify2FALogin(tempToken: string, code: string): Promise<{ token: string; user: User }> {
    const result = await this.request<{ token: string; user: User }>(
      '/auth/2fa/verify',
      {
        method: 'POST',
        body: JSON.stringify({ temp_token: tempToken, code }),
      },
      true // Skip auth redirect
    );
    this.setToken(result.token);
    return result;
  }

  async logout() {
    this.setToken(null);
  }

  async getProfile(): Promise<User> {
    return this.request<User>('/auth/me');
  }

  async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    await this.request('/auth/change-password', {
      method: 'PUT',
      body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
    });
  }

  // 2FA
  async get2FAStatus(): Promise<TwoFactorStatusResponse> {
    return this.request<TwoFactorStatusResponse>('/admin/2fa/status');
  }

  async setup2FA(): Promise<TwoFactorSetupResponse> {
    return this.request<TwoFactorSetupResponse>('/admin/2fa/setup', {
      method: 'POST',
    });
  }

  async verify2FA(code: string): Promise<TwoFactorVerifyResponse> {
    return this.request<TwoFactorVerifyResponse>('/admin/2fa/verify', {
      method: 'POST',
      body: JSON.stringify({ code }),
    });
  }

  async disable2FA(code: string): Promise<{ message: string }> {
    return this.request<{ message: string }>('/admin/2fa/disable', {
      method: 'POST',
      body: JSON.stringify({ code }),
    });
  }

  async regenerateBackupCodes(code: string): Promise<TwoFactorBackupCodesResponse> {
    return this.request<TwoFactorBackupCodesResponse>('/admin/2fa/backup-codes', {
      method: 'POST',
      body: JSON.stringify({ code }),
    });
  }

  // Users (admin routes)
  async getUsers(): Promise<User[]> {
    const result = await this.request<{ users: User[] }>('/admin/users');
    return result.users ?? [];
  }

  async createUser(data: CreateUserRequest): Promise<User> {
    return this.request<User>('/admin/users', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateUser(id: string, data: UpdateUserRequest): Promise<User> {
    return this.request<User>(`/admin/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteUser(id: string): Promise<void> {
    await this.request(`/admin/users/${id}`, { method: 'DELETE' });
  }

  // Projects (admin routes)
  async getProjects(): Promise<Project[]> {
    const result = await this.request<{ projects: Project[] }>('/admin/projects');
    return result.projects ?? [];
  }

  async getProject(id: string): Promise<Project> {
    const result = await this.request<{ project: Project }>(`/admin/projects/${id}`);
    return result.project;
  }

  async createProject(data: CreateProjectRequest): Promise<{ project: Project; api_key: string }> {
    return this.request<{ project: Project; api_key: string }>('/admin/projects', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateProject(id: string, data: UpdateProjectRequest): Promise<Project> {
    return this.request<Project>(`/admin/projects/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteProject(id: string): Promise<void> {
    await this.request(`/admin/projects/${id}`, { method: 'DELETE' });
  }

  async rotateApiKey(projectId: string): Promise<{ api_key: string }> {
    return this.request<{ api_key: string }>(`/admin/projects/${projectId}/rotate-key`, {
      method: 'POST',
    });
  }

  // Project Members (admin routes)
  async getProjectMembers(projectId: string): Promise<ProjectMember[]> {
    const result = await this.request<{ members: ProjectMember[] }>(`/admin/projects/${projectId}/members`);
    return result.members ?? [];
  }

  async addProjectMember(projectId: string, userId: string, role: string): Promise<void> {
    await this.request(`/admin/projects/${projectId}/members`, {
      method: 'POST',
      body: JSON.stringify({ user_id: userId, role }),
    });
  }

  async updateProjectMemberRole(projectId: string, userId: string, role: string): Promise<void> {
    await this.request(`/admin/projects/${projectId}/members/${userId}`, {
      method: 'PUT',
      body: JSON.stringify({ role }),
    });
  }

  async removeProjectMember(projectId: string, userId: string): Promise<void> {
    await this.request(`/admin/projects/${projectId}/members/${userId}`, { method: 'DELETE' });
  }

  // Logs (admin routes)
  async getLogs(params: LogsQueryParams): Promise<LogsResponse> {
    const searchParams = new URLSearchParams();
    if (params.project_id) searchParams.set('project_id', params.project_id.toString());
    if (params.level) searchParams.set('level', params.level);
    if (params.search) searchParams.set('search', params.search);
    if (params.from) searchParams.set('from', params.from);
    if (params.to) searchParams.set('to', params.to);
    if (params.page) searchParams.set('page', params.page.toString());
    if (params.limit) searchParams.set('limit', params.limit.toString());

    return this.request<LogsResponse>(`/admin/logs?${searchParams.toString()}`);
  }

  async getLog(id: string): Promise<LogEntry> {
    return this.request<LogEntry>(`/admin/logs/${id}`);
  }

  async deleteLogs(ids: string[]): Promise<void> {
    await this.request('/admin/logs', {
      method: 'DELETE',
      body: JSON.stringify({ ids }),
    });
  }

  // Channels (admin routes)
  async getChannels(projectId: string): Promise<Channel[]> {
    const result = await this.request<{ channels: Channel[] }>(`/admin/projects/${projectId}/channels`);
    return result.channels ?? [];
  }

  async createChannel(projectId: string, data: CreateChannelRequest): Promise<Channel> {
    return this.request<Channel>(`/admin/projects/${projectId}/channels`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateChannel(channelId: string, data: UpdateChannelRequest): Promise<Channel> {
    return this.request<Channel>(`/admin/channels/${channelId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteChannel(channelId: string): Promise<void> {
    await this.request(`/admin/channels/${channelId}`, { method: 'DELETE' });
  }

  async testChannel(channelId: string): Promise<void> {
    await this.request(`/admin/channels/${channelId}/test`, { method: 'POST' });
  }

  // Push Subscriptions
  async subscribePush(subscription: PushSubscriptionRequest): Promise<void> {
    await this.request('/push/subscribe', {
      method: 'POST',
      body: JSON.stringify(subscription),
    });
  }

  async unsubscribePush(endpoint: string): Promise<void> {
    await this.request('/push/unsubscribe', {
      method: 'POST',
      body: JSON.stringify({ endpoint }),
    });
  }

  async getVapidPublicKey(): Promise<{ public_key: string }> {
    return this.request<{ public_key: string }>('/push/vapid-key');
  }

  async testPushNotification(): Promise<{ message: string; sent_to: number }> {
    return this.request<{ message: string; sent_to: number }>('/push/test', {
      method: 'POST',
    });
  }

  // Stats (admin routes)
  async getStats(): Promise<DashboardStats> {
    return this.request<DashboardStats>('/admin/stats/overview');
  }

  async getProjectStats(projectId: string): Promise<ProjectStats> {
    return this.request<ProjectStats>(`/admin/stats/projects/${projectId}`);
  }

  // Version (public)
  async getVersion(): Promise<VersionInfo> {
    return this.request<VersionInfo>('/version', {}, true);
  }

  async checkForUpdates(): Promise<UpdateCheckInfo> {
    return this.request<UpdateCheckInfo>('/version/check', {}, true);
  }
}

// Types
export interface User {
  id: string;
  username: string;
  name: string;
  role: 'ADMIN' | 'USER';
  is_active: boolean;
  two_factor_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface LoginResponse {
  token?: string;
  user?: User;
  requires_2fa?: boolean;
  temp_token?: string;
}

export interface TwoFactorSetupResponse {
  secret: string;
  qr_code: string;
}

export interface TwoFactorStatusResponse {
  enabled: boolean;
  backup_codes_count: number;
}

export interface TwoFactorVerifyResponse {
  message: string;
  backup_codes?: string[];
}

export interface TwoFactorBackupCodesResponse {
  backup_codes: string[];
}

export interface CreateUserRequest {
  username: string;
  password: string;
  name: string;
  role: string;
}

export interface UpdateUserRequest {
  name?: string;
  role?: string;
  is_active?: boolean;
}

export type ProjectIconType = 'initials' | 'icon' | 'image';

export interface Project {
  id: string;
  name: string;
  description: string;
  icon_type: ProjectIconType;
  icon_value: string;
  api_key?: string;
  api_key_prefix?: string;
  is_active?: boolean;
  created_at: string;
  updated_at: string;
  user_role?: string;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  icon_type?: ProjectIconType;
  icon_value?: string;
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  icon_type?: ProjectIconType;
  icon_value?: string;
}

export interface ProjectMember {
  id: string;
  user_id: string;
  project_id: string;
  role: string;
  created_at: string;
  user?: {
    id: string;
    email: string;
    name: string;
    role: string;
    is_active: boolean;
  };
}

export interface LogEntry {
  id: string;
  project_id: string;
  project_name?: string;
  level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'CRITICAL';
  message: string;
  metadata?: Record<string, unknown>;
  source?: string;
  created_at: string;
}

export interface LogsQueryParams {
  project_id?: string | number;
  level?: string;
  search?: string;
  from?: string;
  to?: string;
  page?: number;
  limit?: number;
}

export interface LogsResponse {
  logs: LogEntry[];
  total: number;
  page: number;
  limit: number;
}

export interface Channel {
  id: string;
  project_id: string;
  type: 'PUSH' | 'TELEGRAM' | 'DISCORD';
  name: string;
  config: Record<string, unknown>;
  min_level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'CRITICAL';
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateChannelRequest {
  type: string;
  name: string;
  config: Record<string, unknown>;
  min_level: string;
}

export interface UpdateChannelRequest {
  name?: string;
  config?: Record<string, unknown>;
  min_level?: string;
  is_active?: boolean;
}

export interface DashboardStats {
  total_projects: number;
  total_logs: number;
  logs_today: number;
  logs_by_level: Record<string, number>;
  recent_logs: LogEntry[];
}

export interface ProjectStats {
  total_logs: number;
  logs_today: number;
  logs_by_level: Record<string, number>;
  logs_trend: { date: string; count: number }[];
}

export interface PushSubscriptionRequest {
  endpoint: string;
  keys: {
    p256dh: string;
    auth: string;
  };
}

export interface VersionInfo {
  version: string;
  build_time: string;
  git_commit: string;
}

export interface UpdateCheckInfo {
  current_version: string;
  latest_version: string;
  update_available: boolean;
  release_url?: string;
  release_notes?: string;
  published_at?: string;
}

export const api = new ApiClient();
