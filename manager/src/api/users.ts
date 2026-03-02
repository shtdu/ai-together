import { apiClient } from './client';
import type { User } from '../types/models';

export interface CreateUserRequest {
  email: string;
  name: string;
  password: string;
  role: 'manager' | 'member';
}

export interface UpdateUserRequest {
  name?: string;
  role?: 'manager' | 'member';
  password?: string;
}

export interface UsersListResponse {
  users: User[];
}

export const usersApi = {
  listUsers: async (): Promise<UsersListResponse> => {
    const response = await apiClient.get('/api/v1/users');
    return response.data;
  },

  createUser: async (data: CreateUserRequest): Promise<User> => {
    const response = await apiClient.post('/api/v1/users', data);
    return response.data;
  },

  updateUser: async (userId: number, data: UpdateUserRequest): Promise<User> => {
    const response = await apiClient.put(`/api/v1/users/${userId}`, data);
    return response.data;
  },

  deleteUser: async (userId: number): Promise<void> => {
    await apiClient.delete(`/api/v1/users/${userId}`);
  },
};
